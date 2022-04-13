package chesterfield_test

import (
	"net/http"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/kevin-secrist/cfactivecallmonitor/internal/chesterfield"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	policeCallUrl = "https://api.chesterfield.gov/api/Police/V1.0/Calls/CallsForService"
	fireCallUrl   = "https://api.chesterfield.gov/api/Fire/V1.0/Calls/CallsForService"
)

var localLocation, _ = time.LoadLocation("America/New_York")

var _ = Describe("Chesterfield API Client", func() {
	It("returns a list of active police calls", func() {
		responder, _ := httpmock.NewJsonResponder(200, httpmock.File("sample_responses/police_calls.json"))
		httpmock.RegisterResponder("GET", policeCallUrl, responder)

		result, err := subject.GetPoliceCalls()

		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).ShouldNot(BeNil())
		Expect(len(result)).To(Equal(2))
		Expect(result).To(Equal(chesterfield.CallForService{
			{
				ID:                    "0123",
				CallReceived:          chesterfield.CustomTime{Time: time.Date(2022, 3, 23, 23, 22, 39, 0, localLocation)},
				Location:              "22XX FAKE RD",
				Type:                  "SUSPICIOUS SITUATION",
				CurrentStatus:         "Dispatched",
				Area:                  "11",
				Priority:              "3",
				CallReceivedFormatted: "3/23/2022 11:22 PM",
			},
			{
				ID:                    "0124",
				CallReceived:          chesterfield.CustomTime{Time: time.Date(2022, 3, 23, 23, 30, 3, 0, localLocation)},
				Location:              "43XX EXAMPLE CT",
				Type:                  "DOMESTIC",
				CurrentStatus:         "Dispatched",
				Area:                  "60",
				Priority:              "2",
				CallReceivedFormatted: "3/23/2022 11:30 PM",
			},
		}))
	})
	It("returns a list of active fire calls", func() {
		responder, _ := httpmock.NewJsonResponder(200, httpmock.File("sample_responses/fire_calls.json"))
		httpmock.RegisterResponder("GET", fireCallUrl, responder)

		result, err := subject.GetFireCalls()

		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).ShouldNot(BeNil())
		Expect(len(result)).To(Equal(1))
		Expect(result).To(Equal(chesterfield.CallForService{
			{
				ID:                    "1234",
				CallReceived:          chesterfield.CustomTime{Time: time.Date(2022, 3, 27, 12, 30, 25, 0, localLocation)},
				Location:              "123XX DIFFERENT ST",
				Type:                  "EMS CALL",
				CurrentStatus:         "Dispatched",
				Area:                  "F20",
				Priority:              "3",
				CallReceivedFormatted: "3/27/2022 12:30 PM",
			},
		}))
	})
	It("passes correct headers", func() {
		httpmock.RegisterResponder("GET", policeCallUrl,
			func(req *http.Request) (*http.Response, error) {
				resp, err := httpmock.NewJsonResponse(200, httpmock.File("sample_responses/police_calls.json"))
				if err != nil {
					return nil, err
				}

				Expect(req.Header["Accept"][0]).To(Equal("application/json"))
				Expect(req.Header["Authorization"][0]).To(Equal("Bearer testApiKey"))
				Expect(req.Header["Referer"][0]).To(Equal("https://www.chesterfield.gov/"))

				return resp, nil
			},
		)

		result, err := subject.GetPoliceCalls()

		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).ShouldNot(BeNil())
		Expect(len(result)).To(Equal(2))
	})
	It("returns error on non-successful status code", func() {
		responder := httpmock.NewStringResponder(500, "")
		httpmock.RegisterResponder("GET", policeCallUrl, responder)

		result, err := subject.GetPoliceCalls()

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("received invalid status code: 500"))
		Expect(result).To(BeNil())
	})
})
