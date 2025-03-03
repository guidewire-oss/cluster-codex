package cmd_test

import (
	"cluster-codex/internal/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func TestCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Generate Suite")
}

var _ = BeforeSuite(func() {
	config.ConfigureLogger("info") // Initialize the logger once
})
