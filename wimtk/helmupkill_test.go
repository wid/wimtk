package main

import (
	"testing"
)

func TestHelmUpKill(t *testing.T) {
	helmUpKill("test", []string{"killme"})
}
