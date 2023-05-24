package client

import (
	"flag"
	"path/filepath"

	"github.com/openinsight-proj/elastic-alert/pkg/utils/logger"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var gClientset *kubernetes.Clientset

type KubeClient struct {
	Client        *kubernetes.Clientset
	Namespace     string
	ConfigmapName string
}

func NewKubeClient(cfg map[string]any) (*KubeClient, error) {
	kc := &KubeClient{}
	ns, ok := cfg["namespace"]
	if ok {
		namespace, ok := ns.(string)
		if !ok {
			panic("NewKubeClient:namespace error")
		}
		kc.Namespace = namespace
	}
	cn, ok := cfg["configmap_name"]
	if ok {
		configmapName, ok := cn.(string)
		if !ok {
			panic("NewKubeClient:configmapName error")
		}
		kc.ConfigmapName = configmapName
	}
	cs, err := GetClientSet()
	if err != nil {
		logger.Logger.Errorf("fialed to create client set %v", err)
		return nil, err
	}

	kc.Client = cs

	return kc, err
}

// GetClientSet get client set
func GetClientSet() (cclientset *kubernetes.Clientset, err error) {
	if gClientset != nil {
		return gClientset, nil
	}
	var kubeconfig *string
	var clientset *kubernetes.Clientset

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "~/.kube/config", "absolute path to the kubeconfig file")
	}

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return clientset, err
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return clientset, err
	}
	gClientset = clientset

	return clientset, nil
}
