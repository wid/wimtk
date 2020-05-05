package main

import (
	"testing"
)

func TestCreateFilenameContentMapping(t *testing.T) {
	filenameContentMapping := createFilenameContentMapping([]string{"fixture/a.txt", "fixture/b.txt", "fixture/c.txt"})
	if filenameContentMapping["a.txt"] != "a content\n" {
		t.Fail()
	}
}

func TestCreateConfigmap(t *testing.T) {
	deleteIfExist("wimtk")
	filenameContentMapping := createFilenameContentMapping([]string{"fixture/a.txt", "fixture/b.txt", "fixture/c.txt"})
	createConfigmap("wimtk", filenameContentMapping)
}
