package main

import (
	"fmt"
	"os"
	"testing"

	r "github.com/dancannon/gorethink"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = BeforeSuite(func() {
	if os.Getenv("DEV") == "" {
		fmt.Println("This server doesn't seem to be in development environment.")
		fmt.Println("After running the test suite, your vertigo database will be nuked.")
		fmt.Println("Do not proceed unless you are okay with all your data to be removed.")
		fmt.Println("Otherwise, please specify DEV environment variable and rerun.")
		os.Exit(1)
	}
})

func TestVertigo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vertigo Suite")
}

var _ = AfterSuite(func() {
	os.Remove("settings.json")
	session, _ := r.Connect(r.ConnectOpts{
		Address:  os.Getenv("RDB_HOST") + ":" + os.Getenv("RDB_PORT"),
		Database: "vertigo",
	})
	r.DbDrop("vertigo").Run(session)
})
