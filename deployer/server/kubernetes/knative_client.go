package kubernetes

import (
	"context"

	"backend/kubernetes_client/server/kubernetes/deployments"

	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	knative "knative.dev/serving/pkg/apis/serving/v1"
)

func (kub *Kubernetes) DeployKnativeService(username string, ctx context.Context, service knative.Service) error {
	vclaim_name := username + "-vclaim"
	volume_name := username + "-volume"

	volume := CreatePersistentVolume(ctx, vclaim_name, volume_name, username)
	kub.Kubernetes.CoreV1().PersistentVolumes().Create(ctx, &volume, metav1.CreateOptions{})

	volumeClaim := CreatePersistentVolumeClaim(ctx, vclaim_name, volume_name, username)
	kub.Kubernetes.CoreV1().PersistentVolumeClaims("default").Create(ctx, &volumeClaim, metav1.CreateOptions{})

	service_res, err := kub.Knative.Services.Create(ctx, &service, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	waitUntillKnativeServiceIsUp(kub.Knative, service_res)
	return nil
}

func (kub *Kubernetes) DeleteServiceForUser(ctx context.Context, username string) []error {
	propPolicy := metav1.DeletePropagationForeground
	dopt := metav1.DeleteOptions{
		PropagationPolicy: &propPolicy,
	}

	var errors []error
	if err := kub.Knative.Services.Delete(ctx, username, dopt); err != nil {
		errors = append(errors, err)
	}
	if err := kub.Knative.Services.Delete(ctx, username, dopt); err != nil {
		errors = append(errors, err)
	}
	if err := kub.Knative.Routes.Delete(ctx, username, dopt); err != nil {
		errors = append(errors, err)
	}

	if err := kub.Knative.Configs.Delete(ctx, username, dopt); err != nil {
		errors = append(errors, err)
	}
	return errors
}

func (kub *Kubernetes) ReturnRouteIfServiceExists(ctx context.Context, username string) string {
	route, err := kub.Knative.Routes.Get(ctx, username, metav1.GetOptions{})
	if err != nil {
		return ""
	}
	return route.Status.URL.String()
}

func (kub *Kubernetes) DeployServerlessWalletService(ctx context.Context, sgx_active bool, username, frontend_url, frontend_host, backend_url, image string) (string, error) {
	wallet_service := deployments.CreateWalletService(ctx, sgx_active, username, frontend_url, frontend_host, backend_url, image)

	err := kub.DeployKnativeService(username, ctx, wallet_service)
	if err != nil {
		return "", err
	}
	route, err := kub.Knative.Routes.Get(ctx, username, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	log.Debug().Caller().Str("route", route.Status.URL.String()).Msg("Created Wallet Service")
	return route.Status.URL.String(), nil
}
