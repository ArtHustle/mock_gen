package test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestMockGen(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mock Generator Suite")
}

var mockGeneratorPath string

var _ = BeforeSuite(func() {
	var err error
	mockGeneratorPath, err = gexec.Build("bitbucket.org/arthustle/backend/tools/mock_gen/cmd")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
