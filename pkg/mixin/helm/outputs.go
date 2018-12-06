package helm

import (
	"fmt"
	"os"
	"path/filepath"

	errWrap "github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func (m *Mixin) getSecret(namespace, name, key string) (string, error) {
	if namespace == "" {
		namespace = "default"
	}
	home := os.Getenv("HOME")
	kubecfg := filepath.Join(home, ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubecfg)
	if err != nil {
		return "", fmt.Errorf("couldn't build kubernetes config: %s", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", errWrap.Wrap(err, "couldn't create kubernetes client")
	}
	secret, err := clientset.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return "", fmt.Errorf("error getting secret %s from namespace %s: %s", name, namespace, err)
	}
	val, ok := secret.Data[key]
	if !ok {
		return "", fmt.Errorf("couldn't find key %s in secret", key)
	}
	return string(val), nil
}
