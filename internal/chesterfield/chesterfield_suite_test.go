package chesterfield_test

import (
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/kevin-secrist/cfactivecallmonitor/internal/chesterfield"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var subject *chesterfield.ChesterfieldAPIClient

var _ = BeforeSuite(func() {
	subject = chesterfield.New("testApiKey")
	httpmock.ActivateNonDefault(subject.RestClient.GetClient())
})

var _ = BeforeEach(func() {
	httpmock.Reset()
})

var _ = AfterSuite(func() {
	httpmock.DeactivateAndReset()
})

func TestChesterfield(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Chesterfield Suite")
}
