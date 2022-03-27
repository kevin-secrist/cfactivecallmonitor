package chesterfield_test

import (
	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/kevin-secrist/cfactivecallmonitor/internal/chesterfield"
)

var _ = Describe("Chesterfield API Client", func() {
	It("returns a list of active police calls", func() {
		responder, _ := httpmock.NewJsonResponder(200, httpmock.File("sample_responses/service_calls.json"))
		url := "https://api.chesterfield.gov/api/Police/V1.0/Calls/CallsForService"
		httpmock.RegisterResponder("GET", url, responder)

		result, err := subject.GetPoliceCalls()

		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).ShouldNot(BeNil())
		Expect(len(result)).To(Equal(2))
		Expect(result).To(Equal(chesterfield.CallForService{
			{
				ID:                    "0123",
				CallReceived:          "3/23/2022 11:22:39 PM",
				Location:              "22XX FAKE RD",
				Type:                  "SUSPICIOUS SITUATION",
				CurrentStatus:         "Dispatched",
				Area:                  "11",
				Priority:              "3",
				CallReceivedFormatted: "3/23/2022 11:22 PM",
			},
			{
				ID:                    "0124",
				CallReceived:          "3/23/2022 11:30:03 PM",
				Location:              "43XX EXAMPLE CT",
				Type:                  "DOMESTIC",
				CurrentStatus:         "Dispatched",
				Area:                  "60",
				Priority:              "2",
				CallReceivedFormatted: "3/23/2022 11:30 PM",
			},
		}))
	})
})
