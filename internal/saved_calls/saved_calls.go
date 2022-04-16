package saved_calls

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	savedCallsTableName = "SavedCalls"
	secondaryIndexName  = "ActiveIndex"
	isActiveString      = "-"
)

type DynamoDB interface {
	PutItem(ctx context.Context,
		params *dynamodb.PutItemInput,
		optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	Query(ctx context.Context,
		params *dynamodb.QueryInput,
		optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	UpdateItem(ctx context.Context,
		params *dynamodb.UpdateItemInput,
		optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
}

type SavedCallDataAccess struct {
	Service DynamoDB
	clock   func() time.Time
}

type SavedCall struct {
	SortKey         string    `dynamodbav:"sortKey,omitempty"`
	ID              string    `dynamodbav:"id,omitempty"`
	CallType        string    `dynamodbav:"callType,omitempty"`
	CallReason      string    `dynamodbav:"callReason,omitempty"`
	LastKnownStatus string    `dynamodbav:"lastKnownStatus,omitempty"`
	CallReceived    time.Time `dynamodbav:"callReceived,omitempty"`
	CallArrival     time.Time `dynamodbav:"callArrival,omitempty"`
	CallResolved    time.Time `dynamodbav:"callResolved,omitempty"`
	IsActive        string    `dynamodbav:"isActive,omitempty"`
	Location        string    `dynamodbav:"location,omitempty"`
	Area            string    `dynamodbav:"area,omitempty"`
	Priority        string    `dynamodbav:"priority,omitempty"`
	HouseNumber     string    `dynamodbav:"houseNumber,omitempty"`
	StreetName      string    `dynamodbav:"streetName,omitempty"`
}

func normalizeCall(savedCall *SavedCall) {
	savedCall.LastKnownStatus = strings.ToLower(savedCall.LastKnownStatus)
	savedCall.SortKey = strings.Join(
		[]string{
			savedCall.CallReceived.Format("2006/01/02"),
			savedCall.ID,
			savedCall.CallType,
		}, "#",
	)
	savedCall.CallReceived = savedCall.CallReceived.UTC()
	savedCall.CallArrival = savedCall.CallArrival.UTC()
	savedCall.CallResolved = savedCall.CallResolved.UTC()
}

func New(config aws.Config) *SavedCallDataAccess {
	service := dynamodb.NewFromConfig(config)
	clock := func() time.Time { return time.Now().UTC() }

	return &SavedCallDataAccess{
		Service: service,
		clock:   clock,
	}
}

func NewWithClient(dynamoDB DynamoDB, clock func() time.Time) *SavedCallDataAccess {
	return &SavedCallDataAccess{
		Service: dynamoDB,
		clock:   clock,
	}
}

func (dao *SavedCallDataAccess) GetActiveCalls(ctx context.Context) ([]SavedCall, error) {
	keyExpression := expression.Key("isActive").Equal(expression.Value(isActiveString))
	expr, err := expression.
		NewBuilder().
		WithKeyCondition(keyExpression).
		Build()

	if err != nil {
		return nil, err
	}

	params := &dynamodb.QueryInput{
		TableName:                 aws.String(savedCallsTableName),
		IndexName:                 aws.String(secondaryIndexName),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	var result []SavedCall

	paginator := dynamodb.NewQueryPaginator(dao.Service, params, func(qpo *dynamodb.QueryPaginatorOptions) {})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		records := []SavedCall{}
		err = attributevalue.UnmarshalListOfMaps(page.Items, &records)
		if err != nil {
			return nil, err
		}
		result = append(result, records...)
	}

	return result, err
}

func (dao *SavedCallDataAccess) SaveCall(ctx context.Context, activeCall SavedCall) error {
	normalizeCall(&activeCall)

	item, err := attributevalue.MarshalMap(activeCall)

	if err != nil {
		return err
	}

	_, err = dao.Service.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(savedCallsTableName),
		Item:      item,
	})

	return err
}

func (dao *SavedCallDataAccess) UpdateStatus(ctx context.Context, activeCall SavedCall) error {
	normalizeCall(&activeCall)

	var timestampColumnName string
	var isActive string

	switch activeCall.LastKnownStatus {
	case "dispatched":
		isActive = isActiveString
	case "on scene":
		isActive = isActiveString
		timestampColumnName = "callArrival"
	case "resolved":
		isActive = ""
		timestampColumnName = "callResolved"
	default:
		return fmt.Errorf("unknown status: %s", activeCall.LastKnownStatus)
	}

	setExpression := expression.
		Set(expression.Name("lastKnownStatus"), expression.Value(activeCall.LastKnownStatus))

	if timestampColumnName != "" {
		setExpression = setExpression.Set(expression.Name(timestampColumnName), expression.Value(dao.clock()))
	}

	if isActive == "" {
		setExpression = setExpression.Remove(expression.Name("isActive"))
	}

	expr, err := expression.
		NewBuilder().
		WithUpdate(setExpression).
		Build()

	if err != nil {
		return err
	}

	sortKey, err := attributevalue.Marshal(activeCall.SortKey)
	if err != nil {
		return err
	}
	streetName, err := attributevalue.Marshal(activeCall.StreetName)
	if err != nil {
		return err
	}

	_, err = dao.Service.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(savedCallsTableName),
		Key: map[string]types.AttributeValue{
			"streetName": streetName,
			"sortKey":    sortKey,
		},
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
	})

	return err
}
