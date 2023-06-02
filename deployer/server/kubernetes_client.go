package server

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	knative "knative.dev/serving/pkg/apis/serving/v1"
	test "knative.dev/serving/test/v1"
)

const (
	NAMESPACE = "default"
	TIMEOUT   = 180 * time.Second
)

func (s *Server) CheckIfServiceExists(ctx context.Context, username string) (string, error) {
	// Create Deployment
	route, err := s.Clients.Knative.Routes.Get(ctx, username, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return route.Status.URL.String(), nil
}

func (s *Server) DeployContainer(ctx context.Context, username string) (string, error) {
	vclaim := username + "-vclaim"
	volume := username + "-volume"
	s.CreatePersistentVolume(ctx, vclaim, volume, username)
	s.CreatePersistentVolumeClaim(ctx, vclaim, volume, username)
	sgx_resources := corev1.ResourceRequirements{}
	if s.Config.SGX_ACTIVATE {
		sgx_resources = corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				"sgx.intel.com/epc": resource.MustParse("2Mi"),
			},
		}
	}

	service := knative.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      username,
			Namespace: "default",
		},
		Spec: knative.ServiceSpec{
			ConfigurationSpec: knative.ConfigurationSpec{
				Template: knative.RevisionTemplateSpec{
					Spec: knative.RevisionSpec{
						PodSpec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  username,
									Image: s.Config.Image,
									Ports: []corev1.ContainerPort{
										{
											ContainerPort: 8081,
										},
									},
									Env: []corev1.EnvVar{
										{
											Name:  "USE_INSECURE_HTTP",
											Value: "true",
										},
										{
											Name:  "FRONTEND_URL",
											Value: s.Config.FRONTEND_URL,
										},
										{
											Name:  "FRONTEND_HOST",
											Value: s.Config.FRONTEND_HOST,
										},
										{
											Name:  "BACKEND_URL",
											Value: s.Config.BACKEND_URL,
										},
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											MountPath: "/data",
											Name:      "data",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "data",
									VolumeSource: corev1.VolumeSource{
										PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
											ClaimName: username + "-vclaim",
										},
									},
								},
							},
						},
					},
				},
			},
			RouteSpec: knative.RouteSpec{},
		},
	}
	// Create Deployment

	service.Spec.ConfigurationSpec.Template.Spec.PodSpec.Containers[0].Resources = sgx_resources

	result, err := s.Clients.Knative.Services.Create(ctx, &service, metav1.CreateOptions{})
	if err != nil {
		return "", err
	}
	fmt.Printf("Waiting for Service to transition to Ready...")
	if err := test.WaitForServiceState(s.Clients.Knative, result.Name, test.IsServiceReady, "ServiceIsReady"); err != nil {
		return "", err
	}

	fmt.Printf("Checking to ensure Service Status is populated for Ready service")
	if err := test.CheckServiceState(s.Clients.Knative, result.Name, func(s *knative.Service) (bool, error) {
		if s.Status.URL == nil || s.Status.URL.Host == "" {
			return false, fmt.Errorf("url is not present in Service status: %v", s)
		}
		if s.Status.LatestCreatedRevisionName == "" {
			return false, fmt.Errorf("lastCreatedRevision is not present in Service status: %v", s)
		}
		if s.Status.LatestReadyRevisionName == "" {
			return false, fmt.Errorf("lastReadyRevision is not present in Service status: %v", s)

		}
		if s.Status.ObservedGeneration != 1 {
			return false, fmt.Errorf("observedGeneration is not 1 in Service status: %v", s)
		}
		return true, nil
	}); err != nil {
		return "", err
	}
	res, err := s.Clients.Knative.Routes.Get(ctx, username, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	fmt.Printf("Created deployment %q.\n", result.Status.URL.String())
	return res.Status.URL.String(), nil
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

func (s *Server) CreatePersistentVolume(ctx context.Context, vclaim_name string, volume_name string, user string) {

	volume := corev1.PersistentVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolume",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: volume_name,
			Labels: map[string]string{
				"type": "local",
			},
		},
		Spec: corev1.PersistentVolumeSpec{
			StorageClassName: "manual",
			Capacity: corev1.ResourceList{
				"storage": resource.MustParse("10Mi"),
			},
			AccessModes: []corev1.PersistentVolumeAccessMode{"ReadWriteOnce"},
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				HostPath: &corev1.HostPathVolumeSource{Path: "/elonwallet/user/" + user},
			},
			ClaimRef: &corev1.ObjectReference{
				Name:      vclaim_name,
				Namespace: "default",
			},
		},
	}
	s.Clients.Kubernetes.CoreV1().PersistentVolumes().Create(ctx, &volume, metav1.CreateOptions{})
}

func (s *Server) CreatePersistentVolumeClaim(ctx context.Context, vclaim_name string, volume_name string, user string) {
	manual := "manual"
	volumeClaim := corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: vclaim_name,
			Labels: map[string]string{
				"type": "local",
			},
			Namespace: "default",
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &manual,
			AccessModes:      []corev1.PersistentVolumeAccessMode{"ReadWriteOnce"},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": resource.MustParse("5Mi"),
				},
			},
			VolumeName: volume_name,
		},
	}
	s.Clients.Kubernetes.CoreV1().PersistentVolumeClaims("default").Create(ctx, &volumeClaim, metav1.CreateOptions{})
}
