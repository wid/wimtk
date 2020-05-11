package main

import (
	"context"
	"regexp"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

func waitPods(podRegexps []string, statusWatched string) {
	var waitForAllPodState sync.WaitGroup
	stateChan := make(chan string, len(podRegexps))

	for _, podRegexp := range podRegexps {
		VerboseF("Waiting for %v: %v\n", podRegexp, v1.PodPhase(statusWatched))
		waitForAllPodState.Add(1)
		stop := waitPod(podRegexp, statusWatched, stateChan)
		matchedPodNames := getMatchingPodNames(podRegexp)
		if isOnePodAlrealdyInTargetState(matchedPodNames, statusWatched) {
			stateChan <- podRegexp
			close(stop)
		} else {
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
					defer VerboseF("\n")
					if pod.Status.Phase == v1.PodPhase(statusWatched) {
						VerboseF(" => Done")
						stateChan <- podRegexp
					}
				}
			},
		},
	)
	stop := make(chan struct{})
	go controller.Run(stop)
	return stop
}

func getMatchingPodNames(podRegexp string) []string {
	matchedPodNames := []string{}
	clientset := getConfiguredClientSet()
	podList, err := clientset.CoreV1().Pods(getNamespace()).List(context.TODO(), metav1.ListOptions{})
	panicErr(err)

	for _, pod := range podList.Items {
		match, _ := regexp.MatchString(podRegexp, pod.Name)
		if !match {
			DebugF("%v does NOT matches %v => Ignored\n", pod.Name, podRegexp)
			continue
		}
		matchedPodNames = append(matchedPodNames, pod.Name)
		DebugF("%v MATCHES %v ==> Added to checklist\n", pod.Name, podRegexp)
	}

	return matchedPodNames
}

func isOnePodAlrealdyInTargetState(podNames []string, statusWatched string) bool {
	clientset := getConfiguredClientSet()

	for _, podName := range podNames {
		pod, err := clientset.CoreV1().Pods(getNamespace()).Get(context.TODO(), podName, metav1.GetOptions{})
		panicErr(err)
		if pod.Status.Phase == v1.PodPhase(statusWatched) {
			DebugF("%v IS already in %v\n", podName, statusWatched)
			return true
		} else {
			DebugF("%v is NOT already in %v => Waiting for events then\n", podName, statusWatched)
		}
	}
	return false
}

func isPodInList(eventPodName string, podNamesWatched []string) bool {
	for _, podName := range podNamesWatched {
		if eventPodName == podName {
			return true
		}
	}
	return false
}
