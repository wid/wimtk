package main

import (
	"testing"
)

func TestGetNamespace(t *testing.T) {
	if getNamespace() != "pucon" {
		t.Fail()
	}
}

func TestCreateFilenameContentMapping(t *testing.T) {
	filenameContentMapping := createFilenameContentMapping([]string{"fixture/a.txt", "fixture/b.txt", "fixture/c.txt"})
	if filenameContentMapping["a.txt"] != "a content\n" {
		t.Fail()
	}
}

func TestCreateConfigmap(t *testing.T) {
	deleteIfExist("pucon")
	filenameContentMapping := createFilenameContentMapping([]string{"fixture/a.txt", "fixture/b.txt", "fixture/c.txt"})
	createConfigmap("pucon", filenameContentMapping)
}
