package kubernetes

import (
	test "knative.dev/serving/test"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"knative.dev/serving/pkg/client/clientset/versioned"
)

type Kubernetes struct {
	Kubernetes *kubernetes.Clientset
	Knative    *test.ServingClients
}

func CreateKubernetesClients() *Kubernetes {
	kubeconf, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	clients, err := newKnativeServingClients(kubeconf, "default")
	if err != nil {
		panic(err)
	}
	return clients
}

func newKnativeServingClients(cfg *rest.Config, namespace string) (*Kubernetes, error) {
	cfg.QPS = 100
	cfg.Burst = 200
	cs, err := versioned.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &Kubernetes{
		Kubernetes: clientset,
		Knative: &test.ServingClients{
			Configs:   cs.ServingV1().Configurations(namespace),
			Revisions: cs.ServingV1().Revisions(namespace),
			Routes:    cs.ServingV1().Routes(namespace),
			Services:  cs.ServingV1().Services(namespace),
		},
	}, nil
}
