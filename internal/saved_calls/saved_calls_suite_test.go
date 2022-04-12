package saved_calls_test

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/kevin-secrist/cfactivecallmonitor/internal/saved_calls"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

type DynamoDBMock struct {
	mock.Mock
}

func (dynamoDBMock *DynamoDBMock) PutItem(ctx context.Context, input *dynamodb.PutItemInput, options ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	args := dynamoDBMock.Called(ctx, input, options)
	return args.Get(0).(*dynamodb.PutItemOutput), args.Error(1)
}
func (dynamoDBMock *DynamoDBMock) Query(ctx context.Context, input *dynamodb.QueryInput, options ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	args := dynamoDBMock.Called(ctx, input, options)
	return args.Get(0).(*dynamodb.QueryOutput), args.Error(1)
}
func (dynamoDBMock *DynamoDBMock) UpdateItem(ctx context.Context, input *dynamodb.UpdateItemInput, options ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	args := dynamoDBMock.Called(ctx, input, options)
	return args.Get(0).(*dynamodb.UpdateItemOutput), args.Error(1)
}

var subject *saved_calls.SavedCallDataAccess
var dynamoDBMock *DynamoDBMock
var queryOutput *dynamodb.QueryOutput
var currentTime = time.Date(2030, 1, 1, 6, 30, 0, 0, time.UTC)

var _ = BeforeSuite(func() {

})

var _ = BeforeEach(func() {
	dynamoDBMock = new(DynamoDBMock)
	clock := func() time.Time { return currentTime }

	subject = saved_calls.NewWithClient(dynamoDBMock, clock)
})

var _ = AfterSuite(func() {
})

func TestSavedCalls(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Saved Calls Suite")
}
