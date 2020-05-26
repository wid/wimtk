package main

import (
	"context"
	"fmt"
	"regexp"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

func helmUpKill(release string, podRegexps []string) {
	upgradedSeenChan := make(chan struct{}, len(podRegexps))

	stop := waitReleaseUpgrade(release, upgradedSeenChan)
	defer close(stop)
	for {
		<-upgradedSeenChan
		deleteMatchingPods(podRegexps)
	}
}

func waitReleaseUpgrade(release string, upgradedSeenChan chan<- struct{}) chan struct{} {
	clientset := getConfiguredClientSet()
	watchlist := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "secrets", getNamespace(), fields.Everything())

	_, controller := cache.NewInformer(
		watchlist,
		&v1.Secret{},
		time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(newObj interface{}) {
				if secrets, ok := newObj.(*v1.Secret); ok {
					secretRegexp := fmt.Sprintf("sh.helm.release.v1.%s.*", release)
					match, _ := regexp.MatchString(secretRegexp, secrets.Name)
					if !match {
						return
					}
					VerboseF("Upgrade detected\n")
					upgradedSeenChan <- struct{}{}
				}
			},
		},
	)
	stop := make(chan struct{})
	go controller.Run(stop)
	return stop
}

func deleteMatchingPods(podRegexp []string) []string {
	matchedPodNames := []string{}
	clientset := getConfiguredClientSet()
	podList, err := clientset.CoreV1().Pods(getNamespace()).List(context.TODO(), metav1.ListOptions{})
	panicErr(err)

	for _, podRegexp := range podRegexp {
		for _, pod := range podList.Items {
			match, _ := regexp.MatchString(podRegexp, pod.Name)
			if !match {
				DebugF("%v: does NOT matches %v => Ignored\n", pod.Name, podRegexp)
				continue
			}
			clientset.CoreV1().Pods(getNamespace()).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
			DebugF("%v: MATCHES %v ==> Deleted\n", pod.Name, podRegexp)
		}
	}
	return matchedPodNames
}
