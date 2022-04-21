package harvester_test

import (
	"errors"
	"time"

	"github.com/kevin-secrist/cfactivecallmonitor/internal/chesterfield"
	"github.com/kevin-secrist/cfactivecallmonitor/internal/saved_calls"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Harvester", func() {
	It("does nothing for no calls", func() {
		chesterfieldMock.On("GetFireCalls").Return(chesterfield.CallForService{}, nil)
		chesterfieldMock.On("GetPoliceCalls").Return(chesterfield.CallForService{}, nil)

		daoMock.On("GetActiveCalls", ctx).Return([]saved_calls.SavedCall{}, nil)

		err := subject.Harvest(ctx)

		Expect(len(chesterfieldMock.Calls)).To(Equal(2))
		Expect(len(daoMock.Calls)).To(Equal(1))
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("stores a police call", func() {
		chesterfieldMock.On("GetFireCalls").Return(chesterfield.CallForService{}, nil)
		chesterfieldMock.On("GetPoliceCalls").Return(policeCall, nil)

		daoMock.On("GetActiveCalls", ctx).Return([]saved_calls.SavedCall{}, nil)
		daoMock.On("SaveCall", ctx, mock.MatchedBy(func(activeCall saved_calls.SavedCall) bool {
			Expect(activeCall.ID).To(Equal("0123"))
			Expect(activeCall.CallType).To(Equal("police"))
			Expect(activeCall.CallReason).To(Equal("SUSPICIOUS SITUATION"))
			Expect(activeCall.LastKnownStatus).To(Equal("Dispatched"))
			Expect(activeCall.CallReceived).To(Equal(time.Date(2022, 3, 23, 23, 22, 39, 0, localLocation)))
			Expect(activeCall.Location).To(Equal("22XX FAKE RD"))
			Expect(activeCall.Area).To(Equal("11"))
			Expect(activeCall.Priority).To(Equal("3"))
			Expect(activeCall.HouseNumber).To(Equal("22XX"))
			Expect(activeCall.StreetName).To(Equal("FAKE RD"))
			return true
		})).Return(nil)
		err := subject.Harvest(ctx)

		Expect(len(chesterfieldMock.Calls)).To(Equal(2))
		Expect(len(daoMock.Calls)).To(Equal(2))
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("stores a fire call", func() {
		chesterfieldMock.On("GetFireCalls").Return(fireCall, nil)
		chesterfieldMock.On("GetPoliceCalls").Return(chesterfield.CallForService{}, nil)

		daoMock.On("GetActiveCalls", ctx).Return([]saved_calls.SavedCall{}, nil)
		daoMock.On("SaveCall", ctx, mock.MatchedBy(func(activeCall saved_calls.SavedCall) bool {
			Expect(activeCall.ID).To(Equal("1234"))
			Expect(activeCall.CallType).To(Equal("fire"))
			Expect(activeCall.CallReason).To(Equal("EMS CALL"))
			Expect(activeCall.LastKnownStatus).To(Equal("Dispatched"))
			Expect(activeCall.CallReceived).To(Equal(time.Date(2022, 3, 27, 12, 30, 25, 0, localLocation)))
			Expect(activeCall.Location).To(Equal("123XX DIFFERENT ST"))
			Expect(activeCall.Area).To(Equal("F20"))
			Expect(activeCall.Priority).To(Equal("3"))
			Expect(activeCall.HouseNumber).To(Equal("123XX"))
			Expect(activeCall.StreetName).To(Equal("DIFFERENT ST"))
			return true
		})).Return(nil)
		err := subject.Harvest(ctx)

		Expect(len(chesterfieldMock.Calls)).To(Equal(2))
		Expect(len(daoMock.Calls)).To(Equal(2))
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("updates a call", func() {
		policeCall[0].CurrentStatus = "On Scene"

		chesterfieldMock.On("GetFireCalls").Return(chesterfield.CallForService{}, nil)
		chesterfieldMock.On("GetPoliceCalls").Return(policeCall, nil)

		daoMock.On("GetActiveCalls", ctx).Return([]saved_calls.SavedCall{savedCall}, nil)

		daoMock.On("UpdateStatus", ctx, mock.MatchedBy(func(activeCall saved_calls.SavedCall) bool {
			Expect(activeCall.ID).To(Equal("0123"))
			Expect(activeCall.CallType).To(Equal("police"))
			Expect(activeCall.CallReason).To(Equal("SUSPICIOUS SITUATION"))
			Expect(activeCall.LastKnownStatus).To(Equal("On Scene"))
			Expect(activeCall.CallReceived).To(Equal(time.Date(2022, 3, 23, 23, 22, 39, 0, localLocation)))
			Expect(activeCall.Location).To(Equal("22XX FAKE RD"))
			Expect(activeCall.Area).To(Equal("11"))
			Expect(activeCall.Priority).To(Equal("3"))
			Expect(activeCall.HouseNumber).To(Equal("22XX"))
			Expect(activeCall.StreetName).To(Equal("FAKE RD"))
			return true
		})).Return(nil)
		err := subject.Harvest(ctx)

		Expect(len(chesterfieldMock.Calls)).To(Equal(2))
		Expect(len(daoMock.Calls)).To(Equal(2))
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("skips updates if status did not change", func() {
		chesterfieldMock.On("GetFireCalls").Return(chesterfield.CallForService{}, nil)
		chesterfieldMock.On("GetPoliceCalls").Return(policeCall, nil)

		daoMock.On("GetActiveCalls", ctx).Return([]saved_calls.SavedCall{savedCall}, nil)

		err := subject.Harvest(ctx)

		Expect(len(chesterfieldMock.Calls)).To(Equal(2))
		Expect(len(daoMock.Calls)).To(Equal(1))
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("resolves a call", func() {
		chesterfieldMock.On("GetFireCalls").Return(chesterfield.CallForService{}, nil)
		chesterfieldMock.On("GetPoliceCalls").Return(chesterfield.CallForService{}, nil)

		daoMock.On("GetActiveCalls", ctx).Return([]saved_calls.SavedCall{savedCall}, nil)

		daoMock.On("UpdateStatus", ctx, mock.MatchedBy(func(activeCall saved_calls.SavedCall) bool {
			Expect(activeCall.ID).To(Equal("0123"))
			Expect(activeCall.CallType).To(Equal("police"))
			Expect(activeCall.CallReason).To(Equal("SUSPICIOUS SITUATION"))
			Expect(activeCall.LastKnownStatus).To(Equal("resolved"))
			Expect(activeCall.CallReceived).To(Equal(time.Date(2022, 3, 23, 23, 22, 39, 0, localLocation)))
			Expect(activeCall.Location).To(Equal("22XX FAKE RD"))
			Expect(activeCall.Area).To(Equal("11"))
			Expect(activeCall.Priority).To(Equal("3"))
			Expect(activeCall.HouseNumber).To(Equal("22XX"))
			Expect(activeCall.StreetName).To(Equal("FAKE RD"))
			return true
		})).Return(nil)
		err := subject.Harvest(ctx)

		Expect(len(chesterfieldMock.Calls)).To(Equal(2))
		Expect(len(daoMock.Calls)).To(Equal(2))
		Expect(err).ShouldNot(HaveOccurred())
	})

	Describe("propagates errors", func() {
		unexpectedError := errors.New("error!")

		It("when fetching police calls", func() {
			chesterfieldMock.On("GetFireCalls").Return(chesterfield.CallForService{}, unexpectedError)
			chesterfieldMock.On("GetPoliceCalls").Return(chesterfield.CallForService{}, nil)
			daoMock.On("GetActiveCalls", ctx).Return([]saved_calls.SavedCall{}, nil)

			err := subject.Harvest(ctx)

			Expect(len(chesterfieldMock.Calls)).To(Equal(2))
			Expect(len(daoMock.Calls)).To(Equal(1))
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(Equal("error!"))
		})

		It("when fetching fire calls", func() {
			chesterfieldMock.On("GetFireCalls").Return(chesterfield.CallForService{}, nil)
			chesterfieldMock.On("GetPoliceCalls").Return(chesterfield.CallForService{}, unexpectedError)
			daoMock.On("GetActiveCalls", ctx).Return([]saved_calls.SavedCall{}, nil)

			err := subject.Harvest(ctx)

			Expect(len(chesterfieldMock.Calls)).To(Equal(2))
			Expect(len(daoMock.Calls)).To(Equal(1))
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(Equal("error!"))
		})

		It("when fetching active calls", func() {
			chesterfieldMock.On("GetFireCalls").Return(chesterfield.CallForService{}, nil)
			chesterfieldMock.On("GetPoliceCalls").Return(chesterfield.CallForService{}, nil)
			daoMock.On("GetActiveCalls", ctx).Return([]saved_calls.SavedCall{}, unexpectedError)

			err := subject.Harvest(ctx)

			Expect(len(chesterfieldMock.Calls)).To(Equal(2))
			Expect(len(daoMock.Calls)).To(Equal(1))
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(Equal("error!"))
		})

		It("when saving a call", func() {
			chesterfieldMock.On("GetFireCalls").Return(chesterfield.CallForService{}, nil)
			chesterfieldMock.On("GetPoliceCalls").Return(policeCall, nil)
			daoMock.On("GetActiveCalls", ctx).Return([]saved_calls.SavedCall{}, nil)

			daoMock.On("SaveCall", ctx, mock.Anything).Return(unexpectedError)

			err := subject.Harvest(ctx)

			Expect(len(chesterfieldMock.Calls)).To(Equal(2))
			Expect(len(daoMock.Calls)).To(Equal(2))
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(Equal("error!"))
		})

		It("when updating status", func() {
			policeCall[0].CurrentStatus = "On Scene"

			chesterfieldMock.On("GetFireCalls").Return(chesterfield.CallForService{}, nil)
			chesterfieldMock.On("GetPoliceCalls").Return(policeCall, nil)
			daoMock.On("GetActiveCalls", ctx).Return([]saved_calls.SavedCall{savedCall}, nil)

			daoMock.On("UpdateStatus", ctx, mock.Anything).Return(unexpectedError)

			err := subject.Harvest(ctx)

			Expect(len(chesterfieldMock.Calls)).To(Equal(2))
			Expect(len(daoMock.Calls)).To(Equal(2))
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(Equal("error!"))
		})

		It("when resolving a call", func() {
			chesterfieldMock.On("GetFireCalls").Return(chesterfield.CallForService{}, nil)
			chesterfieldMock.On("GetPoliceCalls").Return(chesterfield.CallForService{}, nil)
			daoMock.On("GetActiveCalls", ctx).Return([]saved_calls.SavedCall{savedCall}, nil)

			daoMock.On("UpdateStatus", ctx, mock.Anything).Return(unexpectedError)

			err := subject.Harvest(ctx)

			Expect(len(chesterfieldMock.Calls)).To(Equal(2))
			Expect(len(daoMock.Calls)).To(Equal(2))
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(Equal("error!"))
		})
	})
})
