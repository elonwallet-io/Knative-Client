package kubernetes

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreatePersistentVolume(ctx context.Context, vclaim_name string, volume_name string, user string) corev1.PersistentVolume {
	return corev1.PersistentVolume{
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
}

func CreatePersistentVolumeClaim(ctx context.Context, vclaim_name string, volume_name string, user string) corev1.PersistentVolumeClaim {
	manual := "manual"
	return corev1.PersistentVolumeClaim{
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
}
