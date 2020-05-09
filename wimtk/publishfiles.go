package main

import (
	"context"
	"io/ioutil"
	"path"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func publishFiles(filesToPublish []string, configMapName string) {
	filenameContentMapping := createFilenameContentMapping(filesToPublish)
	deleteIfExist(configMapName)
	createConfigmap(configMapName, filenameContentMapping)
}

func createFilenameContentMapping(filesToPublish []string) map[string]string {

	filenameContentMapping := make(map[string]string)
	for _, filename := range filesToPublish {
		content, err := ioutil.ReadFile(filename)
		panicErr(err)
		basename := path.Base(filename)
		VerboseF("Provisioning %v for publishing", basename)
		filenameContentMapping[basename] = string(content)
	}
	return filenameContentMapping
}

func createConfigmap(name string, filenameContentMapping map[string]string) {
	clientset := getConfiguredClientSet()
	configMap := &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: filenameContentMapping,
	}
	_, err := clientset.CoreV1().ConfigMaps(getNamespace()).Create(context.TODO(), configMap, metav1.CreateOptions{})
	VerboseF("Created %v", name)
	panicErr(err)
}
