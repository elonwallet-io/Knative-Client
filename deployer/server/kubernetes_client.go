package server

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"knative.dev/pkg/reconciler"
	"knative.dev/pkg/test/logging"
	knative "knative.dev/serving/pkg/apis/serving/v1"
	"knative.dev/serving/test"
)

const (
	NAMESPACE = "default"
	TIMEOUT   = 180 * time.Second
)

func (s *Server) CheckIfServiceExists(ctx context.Context, username string) (string, error) {
	// Create Deployment
	route, err := s.clients.Routes.Get(ctx, username, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return route.Status.URL.String(), nil
}

func (s *Server) DeployContainer(ctx context.Context, username string) (string, error) {
	//"resources": map[string]interface{}{
	//	"limits": map[string]interface{}{
	//		"sgx.intel.com/epc": "2Mi",
	//	},
	//},
	runTime := "kata-qemu"

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
							RuntimeClassName: &runTime,
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
		return "", err
	}
	fmt.Printf("Waiting for Service to transition to Ready...")
	if err := WaitForServiceState(s.clients, result.Name, IsServiceReady, "ServiceIsReady"); err != nil {
		return "", err
	}

	fmt.Printf("Checking to ensure Service Status is populated for Ready service")
	if err := CheckServiceState(s.clients, result.Name, func(s *knative.Service) (bool, error) {
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
	res, err := s.clients.Routes.Get(ctx, username, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	fmt.Printf("Created deployment %q.\n", result.Status.URL.String())
	return res.Status.URL.String(), nil
}

func (clients *ServingClients) DeleteServiceForUser(ctx context.Context, username string) []error {
	propPolicy := metav1.DeletePropagationForeground
	dopt := metav1.DeleteOptions{
		PropagationPolicy: &propPolicy,
	}

	var errors []error
	if err := clients.Services.Delete(ctx, username, dopt); err != nil {
		errors = append(errors, err)
	}
	if err := clients.Services.Delete(ctx, username, dopt); err != nil {
		errors = append(errors, err)
	}
	if err := clients.Routes.Delete(ctx, username, dopt); err != nil {
		errors = append(errors, err)
	}

	if err := clients.Configs.Delete(ctx, username, dopt); err != nil {
		errors = append(errors, err)
	}
	return errors
}

// Functionfrom https://github.com/knative/serving/blob/main/test/v1/service.go
func WaitForServiceState(client *ServingClients, name string, inState func(s *knative.Service) (bool, error), desc string) error {
	span := logging.GetEmitableSpan(context.Background(), fmt.Sprintf("WaitForServiceState/%s/%s", name, desc))
	defer span.End()

	var lastState *knative.Service
	waitErr := wait.PollImmediate(test.PollInterval, test.PollTimeout, func() (bool, error) {
		err := reconciler.RetryTestErrors(func(int) (err error) {
			lastState, err = client.Services.Get(context.Background(), name, metav1.GetOptions{})
			return err
		})
		if err != nil {
			return true, err
		}
		return inState(lastState)
	})

	if waitErr != nil {
		return fmt.Errorf("service %q is not in desired state, got: %#v: %w", name, lastState, waitErr)
	}
	return nil
}

// Function from https://github.com/knative/serving/blob/main/test/v1/service.go
func CheckServiceState(client *ServingClients, name string, inState func(s *knative.Service) (bool, error)) error {
	var s *knative.Service
	err := reconciler.RetryTestErrors(func(int) (err error) {
		s, err = client.Services.Get(context.Background(), name, metav1.GetOptions{})
		return err
	})
	if err != nil {
		return err
	}
	if done, err := inState(s); err != nil {
		return err
	} else if !done {
		return fmt.Errorf("service %q is not in desired state, got: %#v", name, s)
	}
	return nil
}

// Function from https://github.com/knative/serving/blob/main/test/v1/service.go
func IsServiceReady(s *knative.Service) (bool, error) {
	return s.IsReady(), nil
}
