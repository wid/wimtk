package main

import (
	"context"
	"io/ioutil"
	"os"
	"path"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	filesToPublish := os.Args[1:]

	filenameContentMapping := createFilenameContentMapping(filesToPublish)
	deleteIfExist("pucon")
	createConfigmap("pucon", filenameContentMapping)
}

func createFilenameContentMapping(filesToPublish []string) map[string]string {

	filenameContentMapping := make(map[string]string)
	for _, filename := range filesToPublish {
		content, err := ioutil.ReadFile(filename)
		panicErr(err)
		filenameContentMapping[path.Base(filename)] = string(content)
	}
	return filenameContentMapping
}

func getNamespace() string {
	namespace, _ := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	return string(namespace)
}

func deleteIfExist(name string) {
	clientset := getConfiguredClientSet()
	existing, _ := clientset.CoreV1().ConfigMaps(getNamespace()).Get(context.TODO(), name, metav1.GetOptions{})
	if existing != nil {
		err := clientset.CoreV1().ConfigMaps(getNamespace()).Delete(context.TODO(), name, metav1.DeleteOptions{})
		panicErr(err)
	}
}

func getConfiguredClientSet() *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	panicErr(err)
	return clientset
}

func createConfigmap(name string, filenameContentMapping map[string]string) {
	clientset := getConfiguredClientSet()
	configMap := &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "pucon",
		},
		Data: filenameContentMapping,
	}
	_, err := clientset.CoreV1().ConfigMaps(getNamespace()).Create(context.TODO(), configMap, metav1.CreateOptions{})
	panicErr(err)
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}
