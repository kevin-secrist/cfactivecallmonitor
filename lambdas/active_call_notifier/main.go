package main

import (
	"context"
	"fmt"
	"os"

	"github.com/kevin-secrist/cfactivecallmonitor/internal/saved_calls"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

var twilioClient *twilio.RestClient
var toNumber string

func init() {
	toNumber = os.Getenv("SMS_TO")
	accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
	apiKey := os.Getenv("TWILIO_API_KEY")
	apiSecret := os.Getenv("TWILIO_API_SECRET")

	twilioClient = twilio.NewRestClientWithParams(twilio.ClientParams{
		Username:   apiKey,
		Password:   apiSecret,
		AccountSid: accountSid,
	})
}

type DynamoEventChange struct {
	OldImage map[string]types.AttributeValue `json:"OldImage"`
	NewImage map[string]types.AttributeValue `json:"NewImage"`
}

type DynamoEventRecord struct {
	Change    DynamoEventChange `json:"dynamodb"`
	EventName string            `json:"eventName"`
	EventID   string            `json:"eventID"`
}

type DynamoEvent struct {
	Records []DynamoEventRecord `json:"records"`
}

func SendSms(message string) error {
	params := &openapi.CreateMessageParams{}
	params.SetTo(toNumber)
	params.SetBody(message)

	_, err := twilioClient.ApiV2010.CreateMessage(params)
	return err
}

func HandleRequest(ctx context.Context, event DynamoEvent) error {
	for _, record := range event.Records {
		newImage := saved_calls.SavedCall{}
		err := attributevalue.UnmarshalMap(record.Change.NewImage, &newImage)
		if err != nil {
			return err
		}

		oldImage := saved_calls.SavedCall{}
		err = attributevalue.UnmarshalMap(record.Change.OldImage, &oldImage)
		if err != nil {
			return err
		}

		if oldImage.LastKnownStatus != newImage.LastKnownStatus {
			message := fmt.Sprintf(
				"Active call alert at %v: %v, %v",
				newImage.Location,
				newImage.CallReason,
				newImage.LastKnownStatus,
			)
			SendSms(message)
		}
	}

	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
