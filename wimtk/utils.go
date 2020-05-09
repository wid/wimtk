package main

import (
	"context"
	"fmt"
	"io/ioutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func getConfiguredClientSet() *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	panicErr(err)
	return clientset
}

func getNamespace() string {
	namespace, _ := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	return string(namespace)
}

func deleteIfExist(name string) {
	clientset := getConfiguredClientSet()
	_, err := clientset.CoreV1().ConfigMaps(getNamespace()).Get(context.TODO(), name, metav1.GetOptions{})
	if err == nil {
		VerboseF("Deleting %v", name)
		err := clientset.CoreV1().ConfigMaps(getNamespace()).Delete(context.TODO(), name, metav1.DeleteOptions{})
		panicErr(err)
	}
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}

// VerboseF Call Printf if verbose == true
func VerboseF(format string, a ...interface{}) (n int, err error) {
	if verbose {
		return fmt.Printf(format, a...)
	}
	return 0, nil
}
