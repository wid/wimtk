package main

import (
	"context"
	"sync"
	"time"

	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

func syncMap(namespace string, configMapName string) {
	var waitAlways sync.WaitGroup
	syncConfigMapIfPresent(namespace, configMapName)
	waitAlways.Add(1)
	VerboseF("Watching changed on %v in Namespace %v\n", configMapName, namespace)
	stop := waitConfigMapChange(namespace, configMapName)
	defer close(stop)
	waitAlways.Wait()
}

func syncConfigMapIfPresent(namespace string, configMapName string) {
	clientset := getConfiguredClientSet()
	DebugF("Try to find existing %v on %v\n", configMapName, namespace)
	configMap, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if err == nil {
		VerboseF("Found existing %v on %v => Recreating on current namespace\n", configMapName, namespace)
		deleteIfExist(configMapName)
		createConfigmapFromTemplate(configMap)
	}
}

func waitConfigMapChange(namespace string, configMapName string) chan struct{} {
	clientset := getConfiguredClientSet()
	watchlist := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "configmaps", namespace, fields.Everything())

	_, controller := cache.NewInformer(
		watchlist,
		&v1.ConfigMap{},
		time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if addedConfigMap, ok := obj.(*v1.ConfigMap); ok {
					if addedConfigMap.Name != configMapName {
						return
					}
					VerboseF("Add seen on %v Namespace %v => Replicating Add\n", addedConfigMap.Name, namespace)
					deleteIfExist(configMapName)
					createConfigmapFromTemplate(addedConfigMap)
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				if originalConfigMap, ok := newObj.(*v1.ConfigMap); ok {
					if originalConfigMap.Name != configMapName {
						return
					}
					VerboseF("Update seen on %v Namespace %v => Syncing data\n", originalConfigMap.Name, namespace)
					updateConfigmapFromTemplate(originalConfigMap)
				}
			},
			DeleteFunc: func(obj interface{}) {
				if deletedConfigMap, ok := obj.(*v1.ConfigMap); ok {
					if deletedConfigMap.Name != configMapName {
						return
					}
					VerboseF("Delete seen on %v Namespace %v => Removing\n", deletedConfigMap.Name, namespace)
					deleteIfExist(configMapName)
				}
			},
		},
	)
	stop := make(chan struct{})
	go controller.Run(stop)
	return stop
}

func createConfigmapFromTemplate(originalConfigMap *apiv1.ConfigMap) {
	clientset := getConfiguredClientSet()

	configMap := &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: originalConfigMap.ObjectMeta.Name,
		},
		Data: originalConfigMap.Data,
	}

	_, err := clientset.CoreV1().ConfigMaps(getNamespace()).Create(context.TODO(), configMap, metav1.CreateOptions{})
	VerboseF("Created %v\n", originalConfigMap.ObjectMeta.Name)
	panicErr(err)
}

func updateConfigmapFromTemplate(originalConfigMap *apiv1.ConfigMap) {
	clientset := getConfiguredClientSet()

	configMap := &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: originalConfigMap.ObjectMeta.Name,
		},
		Data: originalConfigMap.Data,
	}

	_, err := clientset.CoreV1().ConfigMaps(getNamespace()).Update(context.TODO(), configMap, metav1.UpdateOptions{})
	VerboseF("Updated %v\n", originalConfigMap.ObjectMeta.Name)
	panicErr(err)
}
