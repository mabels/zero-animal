package k8s

import (
	"github.com/mabels/zero-animal/config"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func GetConfig(cfg config.K8sCfg) (*rest.Config, error) {
	// loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	// // if you want to change the loading rules (which files in which order), you can do so here
	// configOverrides := &clientcmd.ConfigOverrides{}
	// kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	// config, err := kubeConfig.ClientConfig()
	// if err != nil {
	// 	// Do something
	// }

	var config *rest.Config
	// var err error
	kubeconfig := "/Users/menabe/.kube/config"

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	/*
		tconfig, err := config.TransportConfig()
		if err != nil {
			return nil, err
		}
		config.Transport, err = transport.HTTPWrappersForConfig(tconfig, config.Transport)
	*/
	// config, err := rest.InClusterConfig()
	// if err != nil {
	// panic(err.Error())
	// }
	// creates the clientset
	// clientset, err := kubernetes.NewForConfig(config)
	// if err != nil {
	// 	panic(err.Error())
	// }
	return config, err
}
