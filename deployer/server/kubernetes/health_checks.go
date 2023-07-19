package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	knative "knative.dev/serving/pkg/apis/serving/v1"
	clients "knative.dev/serving/test"
	test "knative.dev/serving/test/v1"
)

const (
	NAMESPACE = "vault"
	TIMEOUT   = 180 * time.Second
)

func waitUntillKnativeServiceIsUp(knative_client *clients.ServingClients, wallet_service *knative.Service) error {
	if err := test.WaitForServiceState(knative_client, wallet_service.Name, test.IsServiceReady, "ServiceIsReady"); err != nil {
		return err
	}
	if err := test.CheckServiceState(knative_client, wallet_service.Name, func(s *knative.Service) (bool, error) {
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
		return err
	}
	return nil
}

// return a condition function that indicates whether the given pod is
// currently running
func isPodRunning(ctx context.Context, c kubernetes.Interface, podName string) wait.ConditionFunc {
	return func() (bool, error) {
		pods, err := c.CoreV1().Pods(NAMESPACE).List(ctx, metav1.ListOptions{})
		if err != nil {
			return false, err
		}
		for _, pod := range pods.Items {
			if strings.HasPrefix(pod.GetName(), podName) {
				switch pod.Status.Phase {
				case core.PodRunning:
					return true, nil
				default:
					return false, nil
				}
			}
		}

		return false, errors.New("pod does not exist")
	}
}

// Poll up to timeout seconds for pod to enter running state.
// Returns an error if the pod never enters the running state.
func waitForPodRunning(ctx context.Context, c kubernetes.Interface, podName string) error {
	return wait.PollImmediate(3*time.Second, TIMEOUT, isPodRunning(ctx, c, podName))
}
