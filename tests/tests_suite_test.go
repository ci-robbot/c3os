package mos_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/c3os-io/c3os/tests/machine"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "c3os Test Suite")
}

var tempDir string

var machineID string = os.Getenv("MACHINE_ID")

var _ = AfterSuite(func() {
	if os.Getenv("CREATE_VM") == "true" {
		machine.Delete()
	}
})

var _ = BeforeSuite(func() {

	if os.Getenv("CLOUD_INIT") == "" || !filepath.IsAbs(os.Getenv("CLOUD_INIT")) {
		Fail("CLOUD_INIT must be set and must be pointing to a file as an absolute path")
	}

	if machineID == "" {
		machineID = "testvm"
	}

	if os.Getenv("ISO") == "" {
		fmt.Println("ISO missing")
		os.Exit(1)
	}
	if os.Getenv("CREATE_VM") == "true" {
		t, err := ioutil.TempDir("", "")
		Expect(err).ToNot(HaveOccurred())
		tempDir = t

		machine.ID = machineID
		machine.TempDir = t

		sshPort := "2222"

		if os.Getenv("SSH_PORT") != "" {
			sshPort = os.Getenv("SSH_PORT")
		}

		machine.Create(sshPort)
	}
})
