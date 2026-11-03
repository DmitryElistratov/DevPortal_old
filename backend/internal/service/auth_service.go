package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/company/release-portal/internal/model"
	"github.com/company/release-portal/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo       *repository.UserRepository
	permissionRepo *repository.PermissionRepository
	jwtSecret      string
}

func NewAuthService(userRepo *repository.UserRepository, permissionRepo *repository.PermissionRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		permissionRepo: permissionRepo,
		jwtSecret:      jwtSecret,
	}
}

func (s *AuthService) Login(username, password string) (*model.AuthResponse, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{Token: token, User: *user}, nil
}

func (s *AuthService) Register(req *model.RegisterRequest) (*model.AuthResponse, error) {
	// Validate input
	if len(req.Password) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}
	if len(req.Username) < 3 {
		return nil, errors.New("username must be at least 3 characters")
	}

	// Check if user already exists
	existing, err := s.userRepo.GetByUsername(req.Username)
	if err == nil && existing != nil {
		return nil, errors.New("username already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	role := req.Role
	if role == "" {
		role = "developer"
	}

	user := &model.User{
		ID:           generateID(),
		Username:     req.Username,
		Email:        req.Email,
		FullName:     req.FullName,
		Password:     string(hashedPassword),
		Role:         role,
		AuthProvider: "local",
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{Token: token, User: *user}, nil
}

func (s *AuthService) GetByID(id string) (*model.User, error) {
	return s.userRepo.GetByID(id)
}

func (s *AuthService) ValidateToken(tokenStr string) (*model.User, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return s.userRepo.GetByID(userID)
}

func (s *AuthService) generateToken(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":  user.ID,
		"name": user.FullName,
		"role": user.Role,
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// LoginOAuth authenticates a user coming from an external identity provider
// (Keycloak/OIDC, GitHub OAuth, GitLab OAuth). The caller handler is
// responsible for verifying the provider token/authorization code; this method
// only maps the verified provider identity onto a local user record (creating
// one on first login) and issues a portal JWT.
func (s *AuthService) LoginOAuth(provider, externalID, username, email, fullName string) (*model.AuthResponse, error) {
	if provider == "" || externalID == "" {
		return nil, errors.New("provider and external id are required")
	}

	user, err := s.userRepo.GetByExternalID(provider, externalID)
	if err == nil && user != nil {
		token, terr := s.generateToken(user)
		if terr != nil {
			return nil, terr
		}
		return &model.AuthResponse{Token: token, User: *user}, nil
	}

	// Fall back to matching by username (e.g. user was registered locally before).
	existing, uerr := s.userRepo.GetByUsername(username)
	if uerr == nil && existing != nil {
		existing.AuthProvider = provider
		existing.ExternalID = externalID
		if email != "" && existing.Email == "" {
			existing.Email = email
		}
		if fullName != "" && existing.FullName == "" {
			existing.FullName = fullName
		}
		if err := s.userRepo.Update(existing); err != nil {
			return nil, err
		}
		token, terr := s.generateToken(existing)
		if terr != nil {
			return nil, terr
		}
		return &model.AuthResponse{Token: token, User: *existing}, nil
	}

	// First login via provider – provision the account automatically.
	if username == "" {
		username = fmt.Sprintf("%s_%s", provider, externalID)
	}
	role := "developer"
	if admins, aerr := s.userRepo.List(); aerr == nil {
		hasAdmin := false
		for _, u := range admins {
			if u.Role == "admin" {
				hasAdmin = true
				break
			}
		}
		if !hasAdmin {
			role = "admin" // bootstrap: first user becomes admin
		}
	}

	user = &model.User{
		ID:           generateID(),
		Username:     username,
		Email:        email,
		FullName:     fullName,
		Role:         role,
		AuthProvider: provider,
		ExternalID:   externalID,
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}
	return &model.AuthResponse{Token: token, User: *user}, nil
}

// ListUsers returns all portal users (for the Users settings tab).
func (s *AuthService) ListUsers() ([]model.User, error) {
	return s.userRepo.List()
}

func (s *AuthService) GetProjectPermission(userID, serviceID string) (*model.ProjectPermission, error) {
	return s.permissionRepo.GetByUserAndService(userID, serviceID)
}

func (s *AuthService) UpdateProjectPermission(userID, serviceID, role, grantedBy string) error {
	permission := &model.ProjectPermission{
		ID:        generateID(),
		UserID:    userID,
		ServiceID: serviceID,
		Role:      role,
		GrantedBy: grantedBy,
	}
	return s.permissionRepo.Upsert(permission)
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
