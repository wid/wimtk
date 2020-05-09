package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

func waitPods(podRegexps []string, statusWatched string) {
	var eventArrivedWaitGroup sync.WaitGroup
	for _, podRegexp := range podRegexps {
		stop := waitPod(podRegexp, statusWatched, &eventArrivedWaitGroup)
		defer close(stop)
	}
	eventArrivedWaitGroup.Wait()
}

func waitPod(podRegexp string, statusWatched string, eventArrivedWaitGroup *sync.WaitGroup) chan struct{} {
	fmt.Printf("Waiting %v: %v\n", podRegexp, statusWatched)
	clientset := getConfiguredClientSet()
	watchlist := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", getNamespace(), fields.Everything())

	eventArrivedWaitGroup.Add(1)
	_, controller := cache.NewInformer(
		watchlist,
		&v1.Pod{},
		time.Second*0,
		cache.ResourceEventHandlerFuncs{
			UpdateFunc: func(oldObj, newObj interface{}) {
				if pod, ok := newObj.(*v1.Pod); ok {
					match, _ := regexp.MatchString(podRegexp, pod.Name)
					if !match {
						return
					}
					fmt.Printf("Update from %v: %v", pod.Name, pod.Status.Phase)
					if pod.Status.Phase == v1.PodPhase(statusWatched) {
						fmt.Printf(" => Done\n")
						defer eventArrivedWaitGroup.Done()
					} else {
						fmt.Printf("\n")
					}

				}
			},
		},
	)
	stop := make(chan struct{})
	go controller.Run(stop)
	return stop
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
