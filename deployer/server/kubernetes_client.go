package server

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	knative "knative.dev/serving/pkg/apis/serving/v1"
)

const (
	NAMESPACE = "default"
	TIMEOUT   = 180 * time.Second
)

func (s *Server) CheckIfServiceExists(ctx context.Context, username string) (bool, error) {
	// Create Deployment
	_, err := s.clients.Services.Get(ctx, username, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *Server) DeployContainer(ctx context.Context, username string) error {

	//"resources": map[string]interface{}{
	//	"limits": map[string]interface{}{
	//		"sgx.intel.com/epc": "2Mi",
	//	},
	//},

	service := knative.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: username,
		},
		Spec: knative.ServiceSpec{
			ConfigurationSpec: knative.ConfigurationSpec{
				Template: knative.RevisionTemplateSpec{
					Spec: knative.RevisionSpec{
						PodSpec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  username,
									Image: s.image,
									Ports: []corev1.ContainerPort{
										{
											ContainerPort: 8080,
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
	result, err := s.clients.Services.Create(ctx, &service, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	fmt.Printf("Created deployment %q.\n", result.GetName())
	return nil
}

func (s *Server) RemoveContainer(ctx context.Context, username string) error {
	fmt.Printf("name: " + username)
	err := s.clients.Services.Delete(ctx, username, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	fmt.Printf("deployment deleted")
	return nil
}
