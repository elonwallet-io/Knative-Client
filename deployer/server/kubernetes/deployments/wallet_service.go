package deployments

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	knative "knative.dev/serving/pkg/apis/serving/v1"
)

func CreateWalletService(ctx context.Context, sgx_active bool, username, frontend_url, frontend_host, backend_url, image string) knative.Service {
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
									Image: image,
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
											Value: frontend_url,
										},
										{
											Name:  "FRONTEND_HOST",
											Value: frontend_host,
										},
										{
											Name:  "BACKEND_URL",
											Value: backend_url,
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
	isSGXActivated(sgx_active, &service)
	return service
}

func isSGXActivated(sgx_active bool, service *knative.Service) {
	if sgx_active {
		sgx_driver_mount := []corev1.VolumeMount{
			{
				Name:      "dev-sgx-enclave",
				MountPath: "/dev/sgx/enclave",
			},
			{
				Name:      "dev-sgx-enclave",
				MountPath: "/dev/sgx_enclave",
			},
			{
				Name:      "dev-sgx-provision",
				MountPath: "/dev/sgx_provision",
			},
		}

		sgx_driver := []corev1.Volume{
			{
				Name: "dev-sgx-provision",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/dev/sgx_provision",
					},
				},
			},
			{
				Name: "dev-sgx-enclave",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/dev/sgx_enclave",
					},
				},
			},
		}
		service.Spec.ConfigurationSpec.Template.Spec.PodSpec.Volumes = append(service.Spec.ConfigurationSpec.Template.Spec.PodSpec.Volumes, sgx_driver...)

		service.Spec.ConfigurationSpec.Template.Spec.PodSpec.Containers[0].VolumeMounts = append(service.Spec.ConfigurationSpec.Template.Spec.PodSpec.Containers[0].VolumeMounts, sgx_driver_mount...)
		privileged := true
		service.Spec.ConfigurationSpec.Template.Spec.PodSpec.Containers[0].SecurityContext = &corev1.SecurityContext{Privileged: &privileged}
	}
}
