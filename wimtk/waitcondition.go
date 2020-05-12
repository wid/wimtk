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

func waitPodsCondition(podRegexps []string, conditionWatched string, conditionStatus string) {
	var waitForAllPodState sync.WaitGroup
	stateChan := make(chan string, len(podRegexps))

	for _, podRegexp := range podRegexps {
		waitForAllPodState.Add(1)
		stop := waitPhase(podRegexp, conditionWatched, conditionStatus, stateChan)
		matchedPodNames := getMatchingPodNames(podRegexp)
		if areSomePodsAlreadyInTargetCondition(matchedPodNames, conditionWatched, conditionStatus) {
			stateChan <- podRegexp
			close(stop)
		} else {
			VerboseF("%v: Waiting for %v=%v\n", podRegexp, conditionWatched, conditionStatus)
			defer close(stop)
		}
	}
	go listenForPodsCondition(podRegexps, stateChan, &waitForAllPodState)
	waitForAllPodState.Wait()
}

func listenForPodsCondition(podRegexps []string, stateChan chan string, waitForAllPodState *sync.WaitGroup) {
	for {
		podRegexps = filter(<-stateChan, podRegexps)
		if len(podRegexps) == 0 {
			waitForAllPodState.Done()
			return
		}
	}
}

func waitPhase(podRegexp string, conditionWatched string, conditionStatus string, stateChan chan<- string) chan struct{} {
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
					if findCondition(pod, conditionWatched).Status == v1.ConditionStatus(conditionStatus) {
						DebugF("%v: %v IS in %v\n", pod.Name, conditionWatched, conditionStatus)
						stateChan <- podRegexp
					} else {
						DebugF("%v: %v is NOT %v\n", pod.Name, conditionWatched, conditionStatus)
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
			DebugF("%v: does NOT matches %v => Ignored\n", pod.Name, podRegexp)
			continue
		}
		matchedPodNames = append(matchedPodNames, pod.Name)
		DebugF("%v: MATCHES %v ==> Added to checklist\n", pod.Name, podRegexp)
	}

	return matchedPodNames
}

func areSomePodsAlreadyInTargetCondition(podNames []string, conditionWatched string, conditionStatus string) bool {
	clientset := getConfiguredClientSet()

	for _, podName := range podNames {
		pod, err := clientset.CoreV1().Pods(getNamespace()).Get(context.TODO(), podName, metav1.GetOptions{})
		podJSON, _ := json.Marshal(pod)
		DebugF("Current: %v\n", string(podJSON))
		panicErr(err)
		if findCondition(pod, conditionWatched).Status == v1.ConditionStatus(conditionStatus) {
			VerboseF("%v: %v IS already in %v\n", podName, conditionWatched, conditionStatus)
			return true
		} else {
			VerboseF("%v: %v is NOT already in %v\n", podName, conditionWatched, conditionStatus)
		}
	}
	return false
}

func findCondition(pod *v1.Pod, conditionToFind string) *v1.PodCondition {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == v1.PodConditionType(conditionToFind) {
			return &condition
		}
	}
	return &v1.PodCondition{}
}
