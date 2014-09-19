package main

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestVertigo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vertigo Suite")
}

var _ = AfterSuite(func() {
	err := os.Remove("settings.json")
	if err != nil {
		panic(err)
	}
})
