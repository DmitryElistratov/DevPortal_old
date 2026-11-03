#!/usr/bin/env bash
# =============================================================================
# Release-Portal — production build for Kubernetes.
# Builds ONLY the portal services (frontend + backend) and PostgreSQL, with NO
# mock data and NO dev stubs (ArgoCD mock / Keycloak / GitLab CE are local-dev
# only and are never included in this release).
#
# Output: a named, versioned release archive inside the repo:
#         release/release-portal-v<VERSION>.tar.gz
#
# Usage:
#   REGISTRY=registry.example.com/team ./release/build-release.sh [version]
#   VERSION env var is also supported; default version = v1.0.0-<gitsha>
# =============================================================================
set -euo pipefail
cd "$(dirname "$0")/.."
REPO_ROOT="$(pwd)"

VERSION="${1:-${VERSION:-v1.0.0-$(git rev-parse --short HEAD)}}"
REGISTRY="${REGISTRY:-}"           # e.g. registry.example.com/team (empty => load into minikube/kind)
PUSH="${PUSH:-false}"              # true => docker push after build
RELEASE_NAME="release-portal-${VERSION}"
OUT_DIR="$REPO_ROOT/release"
STAGE="$OUT_DIR/$RELEASE_NAME"
ARCHIVE="$OUT_DIR/${RELEASE_NAME}.tar.gz"

FRONTEND_IMAGE="${REGISTRY:+$REGISTRY/}release-portal-frontend:${VERSION}"
BACKEND_IMAGE="${REGISTRY:+$REGISTRY/}release-portal-backend:${VERSION}"

echo "==> Building release $RELEASE_NAME"
[ -d "$STAGE" ] && rm -rf "$STAGE"
mkdir -p "$STAGE"/{images,k8s}

# ---- 1. Build container images (production, no mocks) ----------------------
echo "==> Building backend image: $BACKEND_IMAGE"
docker build -f backend/Dockerfile -t "$BACKEND_IMAGE" backend

echo "==> Building frontend image: $FRONTEND_IMAGE"
# Build context = repo root; .dockerignore excludes node_modules/dist/mocks/scripts
docker build -f Dockerfile.k8s -t "$FRONTEND_IMAGE" .

if [ "$PUSH" = "true" ]; then
  echo "==> Pushing images"
  docker push "$BACKEND_IMAGE"
  docker push "$FRONTEND_IMAGE"
fi

# ---- 2. Save images into the release ---------------------------------------
echo "==> Saving images to tarballs"
docker save -o "$STAGE/images/backend.tar" "$BACKEND_IMAGE"
docker save -o "$STAGE/images/frontend.tar" "$FRONTEND_IMAGE"

# ---- 3. Copy k8s manifests + patch image names ------------------------------
cp -r k8s/* "$STAGE/k8s/"
sed -i "s|image: release-portal-backend:v1.0.0.*|image: ${BACKEND_IMAGE}|" "$STAGE/k8s/backend.yaml"
sed -i "s|image: release-portal-frontend:v1.0.0.*|image: ${FRONTEND_IMAGE}|" "$STAGE/k8s/frontend.yaml"

# ---- 4. Release metadata & instructions -------------------------------------
GIT_SHA="$(git rev-parse HEAD 2>/dev/null || echo unknown)"
cat > "$STAGE/RELEASE_INFO.txt" <<INFO
Release:        $RELEASE_NAME
Version:        $VERSION
Git commit:     $GIT_SHA
Built at:       $(date -u +%Y-%m-%dT%H:%M:%SZ)

Images:
  $BACKEND_IMAGE
  $FRONTEND_IMAGE
(both saved as tarballs under images/, loadable via 'docker load')

Contents:
  k8s/            kustomize manifests (namespace, postgres, backend, frontend+ingress)
  images/*.tar    prebuilt production images (NO mocks, NO dev stubs)
  install.sh      one-shot loader/installer

NOTE: before deploying edit k8s/postgres.yaml and k8s/backend.yaml secrets
(POSTGRES_PASSWORD, JWT_SECRET, DATABASE_URL) and the Ingress host.
INFO

cat > "$STAGE/install.sh" <<'INSTALL'
#!/usr/bin/env bash
# Load images and apply manifests. Requires docker + kubectl configured.
set -euo pipefail
cd "$(dirname "$0")"
echo "==> Loading images"
docker load -i images/backend.tar
docker load -i images/frontend.tar
echo "==> Applying manifests"
kubectl apply -k k8s/
echo "==> Waiting for rollouts"
kubectl -n release-portal rollout status deployment/portal-postgres --timeout=180s
kubectl -n release-portal rollout status deployment/portal-backend  --timeout=180s
kubectl -n release-portal rollout status deployment/portal-frontend --timeout=120s
echo "Done. Portal available via Ingress (host: portal.local by default)."
INSTALL
chmod +x "$STAGE/install.sh"

# ---- 5. Pack everything into the named tarball -------------------------------
echo "==> Packing $ARCHIVE"
tar -czf "$ARCHIVE" -C "$OUT_DIR" "$RELEASE_NAME"
rm -rf "$STAGE"

echo ""
echo "Release ready:"
ls -lh "$ARCHIVE"
tar -tzf "$ARCHIVE" | head -20
