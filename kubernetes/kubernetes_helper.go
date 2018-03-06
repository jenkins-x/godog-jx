package kubernetes

import (
	"os"
	"strings"

	"flag"
	"path/filepath"

	"fmt"

	"github.com/jenkins-x/godog-jx/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func GetKubeClient() (*kubernetes.Clientset, error) {

	//// creates the in-cluster config
	//config, err := rest.InClusterConfig()
	//if err != nil {
	//	panic(err.Error())
	//}
	// creates the clientset

	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func CheckPodStatus(app, state, environment string) error {
	c, err := GetKubeClient()
	if err != nil {
		return fmt.Errorf("error getting a Kubernetes client %v", err)
	}
	m := map[string]string{
		"app": app,
	}
	pods, err := c.CoreV1().Pods(environment).List(metav1.ListOptions{LabelSelector: labels.FormatLabels(m)})
	if err != nil {
		return fmt.Errorf("error getting pods in namespace %s: %v", environment, err)
	}
	if pods == nil || len(pods.Items) == 0 {
		return fmt.Errorf("no pods found matching label %s in namespace %s", app, environment)
	}
	matched := false
	for _, p := range pods.Items {
		utils.LogInfof("found app in %s state", p.Status.Phase)
		matched = strings.EqualFold(string(p.Status.Phase), state)
	}

	if matched {
		return nil
	}
	return fmt.Errorf("no pods found in %s state", state)

}
