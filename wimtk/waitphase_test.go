package main

import (
	"testing"
)

func TestWaitPod(t *testing.T) {
	waitPodsPhase([]string{"regex-pod.*", "pod2"}, "ready")
}
