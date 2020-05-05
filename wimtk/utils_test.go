package main

import (
	"testing"
)

func TestGetNamespace(t *testing.T) {
	if getNamespace() != "wimtk" {
		t.Fail()
	}
}
