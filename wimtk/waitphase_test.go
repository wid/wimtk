package main

import (
	"testing"
)

func NOTestWaitPod(t *testing.T) {
	waitPodsPhase([]string{"regex-pod.*", "pod2"}, "ready")
}
