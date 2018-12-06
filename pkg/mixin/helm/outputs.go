package helm

import (
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"path/filepath"

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
	kubeconfig := flag.String(
		"kubeconfig",
		filepath.Join(home, ".kube", "config"),
		"(optional) absolute path to the kubeconfig file")

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return "", fmt.Errorf("couldn't build kubernetes client: %s", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	secret, err := clientset.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return "", fmt.Errorf("error getting secret %s from namespace %s: %s", name, namespace, err)
	}
	val, ok := secret.Data[key]
	if !ok {
		return "", fmt.Errorf("couldn't find key %s in secret", key)
	}

	result, err := base64.StdEncoding.DecodeString(string(val))
	return string(result), nil
}
