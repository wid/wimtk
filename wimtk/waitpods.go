package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

func waitPods(podNamesWatched []string, statusWatched string) {
	clientset := getConfiguredClientSet()
	watchlist := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", getNamespace(), fields.Everything())
	var eventArrivedWaitGroup sync.WaitGroup
	eventArrivedWaitGroup.Add(1)
	_, controller := cache.NewInformer(
		watchlist,
		&v1.Pod{},
		time.Second*0,
		cache.ResourceEventHandlerFuncs{
			UpdateFunc: func(oldObj, newObj interface{}) {
				if pod, ok := newObj.(*v1.Pod); ok {
					if pod.Status.Phase == v1.PodPhase(statusWatched) {
						logPod(pod)
						defer eventArrivedWaitGroup.Done()
					}

				}
			},
		},
	)
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(stop)
	eventArrivedWaitGroup.Wait()
}

func isPodInList(eventPodName string, podNamesWatched []string) bool {
	for _, podName := range podNamesWatched {
		if eventPodName == podName {
			return true
		}
	}
	return false
}

func logPod(pod *v1.Pod) {
	jsonBytes, err := json.Marshal(pod)
	panicErr(err)
	fmt.Printf("%v\n", string(jsonBytes))
}
