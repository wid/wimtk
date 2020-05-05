package main

import (
	"context"
	"io/ioutil"
	"path"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func publishFiles(filesToPublish []string) {
	filenameContentMapping := createFilenameContentMapping(filesToPublish)
	deleteIfExist("wimtk")
	createConfigmap("wimtk", filenameContentMapping)
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

func createConfigmap(name string, filenameContentMapping map[string]string) {
	clientset := getConfiguredClientSet()
	configMap := &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "wimtk",
		},
		Data: filenameContentMapping,
	}
	_, err := clientset.CoreV1().ConfigMaps(getNamespace()).Create(context.TODO(), configMap, metav1.CreateOptions{})
	panicErr(err)
}
