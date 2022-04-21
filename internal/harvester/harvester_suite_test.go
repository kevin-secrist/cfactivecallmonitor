package harvester_test

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"

	"github.com/kevin-secrist/cfactivecallmonitor/internal/chesterfield"
	"github.com/kevin-secrist/cfactivecallmonitor/internal/harvester"
	"github.com/kevin-secrist/cfactivecallmonitor/internal/saved_calls"
)

type ChesterfieldMock struct {
	mock.Mock
}

type DataAccessObjectMock struct {
	mock.Mock
}

func (dao *DataAccessObjectMock) GetActiveCalls(ctx context.Context) ([]saved_calls.SavedCall, error) {
	args := dao.Called(ctx)
	return args.Get(0).([]saved_calls.SavedCall), args.Error(1)
}
func (dao *DataAccessObjectMock) SaveCall(ctx context.Context, activeCall saved_calls.SavedCall) error {
	args := dao.Called(ctx, activeCall)
	return args.Error(0)

}
func (dao *DataAccessObjectMock) UpdateStatus(ctx context.Context, activeCall saved_calls.SavedCall) error {
	args := dao.Called(ctx, activeCall)
	return args.Error(0)
}

func (apiClient *ChesterfieldMock) GetPoliceCalls() (chesterfield.CallForService, error) {
	args := apiClient.Called()
	return args.Get(0).(chesterfield.CallForService), args.Error(1)
}
func (apiClient *ChesterfieldMock) GetFireCalls() (chesterfield.CallForService, error) {
	args := apiClient.Called()
	return args.Get(0).(chesterfield.CallForService), args.Error(1)
}

var chesterfieldMock *ChesterfieldMock
var daoMock *DataAccessObjectMock
var subject *harvester.Harvester
var ctx = context.TODO()
var localLocation, _ = time.LoadLocation("America/New_York")

var policeCall chesterfield.CallForService
var fireCall chesterfield.CallForService
var savedCall saved_calls.SavedCall

var _ = BeforeSuite(func() {
	log.SetOutput(ioutil.Discard)
})

var _ = BeforeEach(func() {
	chesterfieldMock = &ChesterfieldMock{}
	daoMock = &DataAccessObjectMock{}

	subject = harvester.NewWithClients(chesterfieldMock, daoMock)

	policeCall = chesterfield.CallForService{
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
	}

	fireCall = chesterfield.CallForService{
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
	}

	savedCall = saved_calls.SavedCall{
		ID:              "0123",
		CallType:        "police",
		CallReason:      "SUSPICIOUS SITUATION",
		LastKnownStatus: "Dispatched",
		CallReceived:    time.Date(2022, 3, 23, 23, 22, 39, 0, localLocation),
		IsActive:        "-",
		Location:        "22XX FAKE RD",
		Area:            "11",
		Priority:        "3",
		HouseNumber:     "22XX",
		StreetName:      "FAKE RD",
	}
})

var _ = AfterSuite(func() {
	log.SetOutput(os.Stdout)
})

func TestSavedCalls(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Harvester Suite")
}
