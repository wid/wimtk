package main

import (
	"testing"
)

func TestWaitPod(t *testing.T) {
	waitPods([]string{"test-pod"}, "ready")
}
