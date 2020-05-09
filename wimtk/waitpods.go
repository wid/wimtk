package main

import (
	"regexp"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

func waitPods(podRegexps []string, statusWatched string) {
	var waitForAllPodState sync.WaitGroup
	stateChan := make(chan string)

	for _, podRegexp := range podRegexps {
		waitForAllPodState.Add(1)
		stop := waitPod(podRegexp, statusWatched, stateChan)
		defer close(stop)
	}
	go listenForPodStates(podRegexps, stateChan, &waitForAllPodState)
	waitForAllPodState.Wait()
}

func listenForPodStates(podRegexps []string, stateChan chan string, waitForAllPodState *sync.WaitGroup) {
	for {
		podRegexps = filter(<-stateChan, podRegexps)
		if len(podRegexps) == 0 {
			waitForAllPodState.Done()
			return
		}
	}
}

func filter(needle string, haystack []string) []string {
	if len(haystack) == 0 {
		return []string{}
	}
	if haystack[0] != needle {
		return append(filter(needle, haystack[1:]), haystack[0])
	}
	return filter(needle, haystack[1:])
}

func waitPod(podRegexp string, statusWatched string, stateChan chan<- string) chan struct{} {
	clientset := getConfiguredClientSet()
	watchlist := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", getNamespace(), fields.Everything())

	VerboseF("Watching for %v: %v ", podRegexp, v1.PodPhase(statusWatched))
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
					VerboseF("Update from %v: %v (Want %v)", pod.Name, pod.Status.Phase, v1.PodPhase(statusWatched))
					if pod.Status.Phase == v1.PodPhase(statusWatched) {
						VerboseF(" => Done\n")
						stateChan <- podRegexp
					} else {
						VerboseF("\n")
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
