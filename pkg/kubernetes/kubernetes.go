package kubernetes

import (
	"fmt"

	"github.com/pkg/errors"

	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// ClientFactory is an interface that knows how to create Kubernetes Clients
type ClientFactory interface {
	GetClient(configPath string) (k8s.Interface, error)
}

type clientFactory struct {
}

func (f *clientFactory) GetClient(configPath string) (k8s.Interface, error) {

	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return nil, fmt.Errorf("couldn't build kubernetes config: %s", err)
	}
	clientset, err := k8s.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't create kubernetes client")
	}
	return clientset, nil
}

// New returns an implementation of the ClientFactory interface
func New() ClientFactory {
	return &clientFactory{}
}
