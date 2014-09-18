package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestVertigo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vertigo Suite")
}
