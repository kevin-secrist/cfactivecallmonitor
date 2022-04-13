package saved_calls_test

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/kevin-secrist/cfactivecallmonitor/internal/saved_calls"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var localLocation, _ = time.LoadLocation("America/New_York")

var _ = Describe("Saved Calls DAO", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.TODO()
	})

	Describe("GetActiveCalls()", func() {
		It("returns a list of active calls", func() {
			sampleCallItems := []map[string]types.AttributeValue{
				{
					"partitionKey":    &types.AttributeValueMemberS{Value: "2022/03/23#0123#police"},
					"id":              &types.AttributeValueMemberS{Value: "0123"},
					"callType":        &types.AttributeValueMemberS{Value: "police"},
					"callReason":      &types.AttributeValueMemberS{Value: "SUSPICIOUS SITUATION"},
					"lastKnownStatus": &types.AttributeValueMemberS{Value: "dispatched"},
					"callReceived":    &types.AttributeValueMemberS{Value: "2022-03-23T23:22:39-04:00"},
					"callArrival":     &types.AttributeValueMemberS{Value: "2022-03-23T23:27:39-04:00"},
					"callResolved":    &types.AttributeValueMemberS{Value: "2022-03-23T23:32:39-04:00"},
					"isActive":        &types.AttributeValueMemberBOOL{Value: true},
					"location":        &types.AttributeValueMemberS{Value: "22XX FAKE RD"},
					"area":            &types.AttributeValueMemberS{Value: "11"},
					"priority":        &types.AttributeValueMemberS{Value: "3"},
					"houseNumber":     &types.AttributeValueMemberS{Value: "22XX"},
					"streetName":      &types.AttributeValueMemberS{Value: "FAKE RD"},
				},
				{
					"partitionKey":    &types.AttributeValueMemberS{Value: "2022/03/23#0124#police"},
					"id":              &types.AttributeValueMemberS{Value: "0124"},
					"callType":        &types.AttributeValueMemberS{Value: "police"},
					"callReason":      &types.AttributeValueMemberS{Value: "DOMESTIC"},
					"lastKnownStatus": &types.AttributeValueMemberS{Value: "dispatched"},
					"callReceived":    &types.AttributeValueMemberS{Value: "2022-03-23T23:30:03-04:00"},
					"callArrival":     &types.AttributeValueMemberS{Value: "2022-03-23T23:35:03-04:00"},
					"callResolved":    &types.AttributeValueMemberS{Value: "2022-03-23T23:40:03-04:00"},
					"isActive":        &types.AttributeValueMemberBOOL{Value: true},
					"location":        &types.AttributeValueMemberS{Value: "43XX EXAMPLE CT"},
					"area":            &types.AttributeValueMemberS{Value: "60"},
					"priority":        &types.AttributeValueMemberS{Value: "2"},
					"houseNumber":     &types.AttributeValueMemberS{Value: "43XX"},
					"streetName":      &types.AttributeValueMemberS{Value: "EXAMPLE CT"},
				},
			}

			queryOutput = &dynamodb.QueryOutput{
				Items: sampleCallItems,
				Count: int32(len(sampleCallItems)),
			}

			dynamoDBMock.On("Query", ctx, mock.MatchedBy(func(queryInput *dynamodb.QueryInput) bool {
				input := *queryInput
				Expect(*input.TableName).To(Equal("SavedCalls"))
				Expect(*input.IndexName).To(Equal("ActiveIndex"))
				Expect(*input.KeyConditionExpression).To(Equal("#0 = :0"))
				Expect(input.ExpressionAttributeNames).To(Equal(map[string]string{
					"#0": "isActive",
				}))
				Expect(input.ExpressionAttributeValues).To(Equal(map[string]types.AttributeValue{
					":0": &types.AttributeValueMemberBOOL{Value: true},
				}))

				return true
			}), mock.Anything).Return(queryOutput, nil)

			result, err := subject.GetActiveCalls(ctx)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).ShouldNot(BeNil())
			Expect(len(result)).To(Equal(2))

			Expect(result[0].PartitionKey).To(Equal("2022/03/23#0123#police"))
			Expect(result[0].ID).To(Equal("0123"))
			Expect(result[0].CallType).To(Equal("police"))
			Expect(result[0].CallReason).To(Equal("SUSPICIOUS SITUATION"))
			Expect(result[0].LastKnownStatus).To(Equal("dispatched"))
			Expect(result[0].CallReceived.Equal(time.Date(2022, 3, 24, 03, 22, 39, 0, time.UTC))).To(BeTrue())
			Expect(result[0].CallArrival.Equal(time.Date(2022, 3, 24, 03, 27, 39, 0, time.UTC))).To(BeTrue())
			Expect(result[0].CallResolved.Equal(time.Date(2022, 3, 24, 03, 32, 39, 0, time.UTC))).To(BeTrue())
			Expect(result[0].IsActive).To(Equal(true))
			Expect(result[0].Location).To(Equal("22XX FAKE RD"))
			Expect(result[0].Area).To(Equal("11"))
			Expect(result[0].Priority).To(Equal("3"))
			Expect(result[0].HouseNumber).To(Equal("22XX"))
			Expect(result[0].StreetName).To(Equal("FAKE RD"))

			Expect(result[1].PartitionKey).To(Equal("2022/03/23#0124#police"))
			Expect(result[1].ID).To(Equal("0124"))
			Expect(result[1].CallType).To(Equal("police"))
			Expect(result[1].CallReason).To(Equal("DOMESTIC"))
			Expect(result[1].LastKnownStatus).To(Equal("dispatched"))
			Expect(result[1].CallReceived.Equal(time.Date(2022, 3, 24, 03, 30, 03, 0, time.UTC))).To(BeTrue())
			Expect(result[1].CallArrival.Equal(time.Date(2022, 3, 24, 03, 35, 03, 0, time.UTC))).To(BeTrue())
			Expect(result[1].CallResolved.Equal(time.Date(2022, 3, 24, 03, 40, 03, 0, time.UTC))).To(BeTrue())
			Expect(result[1].IsActive).To(Equal(true))
			Expect(result[1].Location).To(Equal("43XX EXAMPLE CT"))
			Expect(result[1].Area).To(Equal("60"))
			Expect(result[1].Priority).To(Equal("2"))
			Expect(result[1].HouseNumber).To(Equal("43XX"))
			Expect(result[1].StreetName).To(Equal("EXAMPLE CT"))
		})
	})

	Describe("SaveCall()", func() {
		It("stores an object in dynamo", func() {
			callToSave := saved_calls.SavedCall{
				ID:              "0123",
				CallType:        "police",
				CallReason:      "SUSPICIOUS SITUATION",
				LastKnownStatus: "Dispatched",
				CallReceived:    time.Date(2022, 3, 23, 23, 22, 39, 0, localLocation),
				CallArrival:     time.Date(2022, 3, 23, 23, 27, 39, 0, localLocation),
				CallResolved:    time.Date(2022, 3, 23, 23, 32, 39, 0, localLocation),
				IsActive:        true,
				Location:        "22XX FAKE RD",
				Area:            "11",
				Priority:        "3",
				HouseNumber:     "22XX",
				StreetName:      "FAKE RD",
			}

			putOutput := &dynamodb.PutItemOutput{}
			dynamoDBMock.On("PutItem", ctx, mock.MatchedBy(func(putInput *dynamodb.PutItemInput) bool {
				input := *putInput
				Expect(*input.TableName).To(Equal("SavedCalls"))

				Expect(input.Item["partitionKey"]).To(Equal(&types.AttributeValueMemberS{Value: "2022/03/23#0123#police"}))
				Expect(input.Item["id"]).To(Equal(&types.AttributeValueMemberS{Value: "0123"}))
				Expect(input.Item["callType"]).To(Equal(&types.AttributeValueMemberS{Value: "police"}))
				Expect(input.Item["callReason"]).To(Equal(&types.AttributeValueMemberS{Value: "SUSPICIOUS SITUATION"}))
				Expect(input.Item["lastKnownStatus"]).To(Equal(&types.AttributeValueMemberS{Value: "dispatched"}))
				Expect(input.Item["callReceived"]).To(Equal(&types.AttributeValueMemberS{Value: "2022-03-24T03:22:39Z"}))
				Expect(input.Item["callArrival"]).To(Equal(&types.AttributeValueMemberS{Value: "2022-03-24T03:27:39Z"}))
				Expect(input.Item["callResolved"]).To(Equal(&types.AttributeValueMemberS{Value: "2022-03-24T03:32:39Z"}))
				Expect(input.Item["isActive"]).To(Equal(&types.AttributeValueMemberBOOL{Value: true}))
				Expect(input.Item["location"]).To(Equal(&types.AttributeValueMemberS{Value: "22XX FAKE RD"}))
				Expect(input.Item["area"]).To(Equal(&types.AttributeValueMemberS{Value: "11"}))
				Expect(input.Item["priority"]).To(Equal(&types.AttributeValueMemberS{Value: "3"}))
				Expect(input.Item["houseNumber"]).To(Equal(&types.AttributeValueMemberS{Value: "22XX"}))
				Expect(input.Item["streetName"]).To(Equal(&types.AttributeValueMemberS{Value: "FAKE RD"}))

				return true
			}), mock.Anything).Return(putOutput, nil)

			err := subject.SaveCall(ctx, callToSave)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(callToSave.PartitionKey).To(Equal(""))
		})
	})

	Describe("UpdateStatus()", func() {
		var callToSave saved_calls.SavedCall

		BeforeEach(func() {
			callToSave = saved_calls.SavedCall{
				ID:              "0123",
				CallType:        "police",
				CallReason:      "SUSPICIOUS SITUATION",
				LastKnownStatus: "On Scene",
				CallReceived:    time.Date(2022, 3, 23, 23, 22, 39, 0, localLocation),
				IsActive:        true,
				Location:        "22XX FAKE RD",
				Area:            "11",
				Priority:        "3",
				HouseNumber:     "22XX",
				StreetName:      "FAKE RD",
			}
		})
		It("sets arrival time", func() {
			updateOutput := &dynamodb.UpdateItemOutput{}
			dynamoDBMock.On("UpdateItem", ctx, mock.MatchedBy(func(updateInput *dynamodb.UpdateItemInput) bool {
				input := *updateInput
				Expect(*input.TableName).To(Equal("SavedCalls"))
				Expect(len(input.Key)).To(Equal(1))
				Expect(*input.UpdateExpression).To(Equal("SET #0 = :0, #1 = :1\n"))
				Expect(input.Key["partitionKey"]).To(Equal(&types.AttributeValueMemberS{Value: "2022/03/23#0123#police"}))
				Expect(input.ExpressionAttributeNames).To(Equal(map[string]string{
					"#0": "lastKnownStatus",
					"#1": "callArrival",
				}))
				Expect(input.ExpressionAttributeValues).To(Equal(map[string]types.AttributeValue{
					":0": &types.AttributeValueMemberS{Value: "on scene"},
					":1": &types.AttributeValueMemberS{Value: "2030-01-01T06:30:00Z"},
				}))

				return true
			}), mock.Anything).Return(updateOutput, nil)

			err := subject.UpdateStatus(ctx, callToSave)

			Expect(err).ShouldNot(HaveOccurred())
		})
		It("sets resolve time", func() {
			callToSave.CallArrival = time.Date(2022, 3, 23, 23, 27, 39, 0, localLocation)
			callToSave.LastKnownStatus = "Resolved"

			updateOutput := &dynamodb.UpdateItemOutput{}
			dynamoDBMock.On("UpdateItem", ctx, mock.MatchedBy(func(updateInput *dynamodb.UpdateItemInput) bool {
				input := *updateInput
				Expect(*input.TableName).To(Equal("SavedCalls"))
				Expect(len(input.Key)).To(Equal(1))
				Expect(*input.UpdateExpression).To(Equal("REMOVE #0\nSET #1 = :0, #2 = :1\n"))
				Expect(input.Key["partitionKey"]).To(Equal(&types.AttributeValueMemberS{Value: "2022/03/23#0123#police"}))
				Expect(input.ExpressionAttributeNames).To(Equal(map[string]string{
					"#0": "isActive",
					"#1": "lastKnownStatus",
					"#2": "callResolved",
				}))
				Expect(input.ExpressionAttributeValues).To(Equal(map[string]types.AttributeValue{
					":0": &types.AttributeValueMemberS{Value: "resolved"},
					":1": &types.AttributeValueMemberS{Value: "2030-01-01T06:30:00Z"},
				}))

				return true
			}), mock.Anything).Return(updateOutput, nil)

			err := subject.UpdateStatus(ctx, callToSave)

			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})
