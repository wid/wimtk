package main

import (
	"context"
	"encoding/json"
	"regexp"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

func waitPodsPhase(podRegexps []string, phaseWatched string) {
	var waitForAllPodState sync.WaitGroup
	stateChan := make(chan string, len(podRegexps))

	for _, podRegexp := range podRegexps {
		waitForAllPodState.Add(1)
		stop := waitCondition(podRegexp, phaseWatched, stateChan)
		matchedPodNames := getMatchingPodNames(podRegexp)
		if areSomePodAlreadyInTargetPhase(matchedPodNames, phaseWatched) {
			stateChan <- podRegexp
			close(stop)
		} else {
			VerboseF("%v: Waiting for %v\n", podRegexp, v1.PodPhase(phaseWatched))
			defer close(stop)
		}
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

func waitCondition(podRegexp string, phaseWatched string, stateChan chan<- string) chan struct{} {
	clientset := getConfiguredClientSet()
	watchlist := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", getNamespace(), fields.Everything())

	_, controller := cache.NewInformer(
		watchlist,
		&v1.Pod{},
		time.Second*0,
		cache.ResourceEventHandlerFuncs{
			UpdateFunc: func(oldObj, newObj interface{}) {
				if pod, ok := newObj.(*v1.Pod); ok {
					podJSON, _ := json.Marshal(pod)
					DebugF("Received: %v\n", string(podJSON))
					match, _ := regexp.MatchString(podRegexp, pod.Name)
					if !match {
						return
					}
					if pod.Status.Phase == v1.PodPhase(phaseWatched) {
						VerboseF("%v: Status Phase IS in %v\n", pod.Name, v1.PodPhase(phaseWatched))
						stateChan <- podRegexp
					} else {
						VerboseF("%v: Status Phase is NOT in %v\n", pod.Name, v1.PodPhase(phaseWatched))
					}

				}
			},
		},
	)
	stop := make(chan struct{})
	go controller.Run(stop)
	return stop
}

func areSomePodAlreadyInTargetPhase(podNames []string, phaseWatched string) bool {
	clientset := getConfiguredClientSet()

	for _, podName := range podNames {
		pod, err := clientset.CoreV1().Pods(getNamespace()).Get(context.TODO(), podName, metav1.GetOptions{})
		podJSON, _ := json.Marshal(pod)
		DebugF("Current: %v\n", string(podJSON))
		panicErr(err)
		if pod.Status.Phase == v1.PodPhase(phaseWatched) {
			VerboseF("%v: IS already in %v\n", podName, phaseWatched)
			return true
		} else {
			VerboseF("%v: is NOT already in %v => Waiting for events then\n", podName, phaseWatched)
		}
	}
	return false
}
