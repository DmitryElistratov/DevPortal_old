package service

import (
	"context"
	"fmt"
	"log"

	"github.com/company/release-portal/internal/integration/kubernetes"
)

type KubernetesService struct {
	k8sClient *kubernetes.Client
}

func NewKubernetesService(k8sClient *kubernetes.Client) *KubernetesService {
	return &KubernetesService{k8sClient: k8sClient}
}

// ListPods returns list of pods in a namespace
func (s *KubernetesService) ListPods(ctx context.Context, namespace string) ([]kubernetes.Pod, error) {
	if s.k8sClient == nil {
		return nil, fmt.Errorf("kubernetes client not configured")
	}

	log.Printf("[Kubernetes] Listing pods in namespace: %s", namespace)
	pods, err := s.k8sClient.ListPods(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	log.Printf("[Kubernetes] Found %d pods", len(pods))
	return pods, nil
}

// ListDeployments returns list of deployments in a namespace
func (s *KubernetesService) ListDeployments(ctx context.Context, namespace string) ([]kubernetes.Deployment, error) {
	if s.k8sClient == nil {
		return nil, fmt.Errorf("kubernetes client not configured")
	}

	log.Printf("[Kubernetes] Listing deployments in namespace: %s", namespace)
	deployments, err := s.k8sClient.ListDeployments(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	log.Printf("[Kubernetes] Found %d deployments", len(deployments))
	return deployments, nil
}

// ListServices returns list of services in a namespace
func (s *KubernetesService) ListServices(ctx context.Context, namespace string) ([]kubernetes.Service, error) {
	if s.k8sClient == nil {
		return nil, fmt.Errorf("kubernetes client not configured")
	}

	log.Printf("[Kubernetes] Listing services in namespace: %s", namespace)
	services, err := s.k8sClient.ListServices(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	log.Printf("[Kubernetes] Found %d services", len(services))
	return services, nil
}

// GetPodLogs returns logs from a pod
func (s *KubernetesService) GetPodLogs(ctx context.Context, namespace, podName string) (string, error) {
	if s.k8sClient == nil {
		return "", fmt.Errorf("kubernetes client not configured")
	}

	log.Printf("[Kubernetes] Getting logs from pod: %s/%s", namespace, podName)
	logs, err := s.k8sClient.GetPodLogs(ctx, namespace, podName)
	if err != nil {
		return "", fmt.Errorf("failed to get pod logs: %w", err)
	}

	log.Printf("[Kubernetes] Retrieved logs from pod: %s", podName)
	return logs, nil
}
