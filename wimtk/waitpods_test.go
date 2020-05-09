package main

import (
	"testing"
)

func TestWaitPod(t *testing.T) {
	waitPods([]string{"regex-pod.*", "pod2"}, "ready")
}
