package k8_test

import (
	"cluster-codex/internal/config"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestK8(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "K8 Suite")
}

var _ = BeforeSuite(func() {
	config.ConfigureLogger("info") // Initialize the logger once
})
