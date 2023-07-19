package kubernetes

import (
	"backend/kubernetes_client/server/kubernetes/deployments"
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kub *Kubernetes) DeployVault(ctx context.Context, image string) error {
	vault_deployment := deployments.CreateVaultDeployment(image)
	_, err := kub.Kubernetes.AppsV1().Deployments("vault").Create(ctx, &vault_deployment, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	vault_service := deployments.CreateVaultService()
	_, err = kub.Kubernetes.CoreV1().Services("vault").Create(ctx, &vault_service, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}
