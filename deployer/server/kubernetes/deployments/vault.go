package deployments

import (
	v1 "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func CreateVaultDeployment(image string) v1.Deployment {
	return v1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "Vault-SGX",
			Labels: map[string]string{"service": "Vault-SGX"},
		},
		Spec: v1.DeploymentSpec{
			Template: core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "vault-sgx",
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:  "vault",
							Image: image,
							Ports: []core.ContainerPort{
								{
									ContainerPort: 8200,
								},
							},
							VolumeMounts: []core.VolumeMount{{
								MountPath: "/etc/sgx_default_qcnl.conf",
								Name:      "qcnl-conf",
								SubPath:   "sgx_default_qcnl.conf",
							},
								{
									Name:      "vault-sgx-data",
									MountPath: "/data/",
								}},
						},
					},
					Volumes: []core.Volume{
						{
							Name: "qcnl-conf",
							VolumeSource: core.VolumeSource{
								ConfigMap: &core.ConfigMapVolumeSource{
									LocalObjectReference: core.LocalObjectReference{Name: "sgx-pccs-config"},
								},
							},
						},
						{
							Name: "vault-sgx-data",
							VolumeSource: core.VolumeSource{HostPath: &core.HostPathVolumeSource{
								Path: "/etc/vault/data",
							}},
						},
					},
				},
			},
		},
	}
}

func CreateVaultService() core.Service {
	return core.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "VAULT-SGX",
		},
		Spec: core.ServiceSpec{
			Selector: map[string]string{
				"service": "Vault-SGX",
			},
			Ports: []core.ServicePort{
				{
					Protocol:   "TCP",
					Port:       8200,
					TargetPort: intstr.FromInt(8200),
				},
			},
		},
	}
}
