package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

var twilioClient *twilio.RestClient
var toNumber string
var fromNumber string

func init() {
	toNumber = os.Getenv("SMS_TO")
	fromNumber = os.Getenv("SMS_FROM")
	accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
	apiKey := os.Getenv("TWILIO_API_KEY")
	apiSecret := os.Getenv("TWILIO_API_SECRET")

	twilioClient = twilio.NewRestClientWithParams(twilio.ClientParams{
		Username:   apiKey,
		Password:   apiSecret,
		AccountSid: accountSid,
	})
}

type StreamRecord struct {
	Records []struct {
		EventID      string `json:"eventID"`
		EventName    string `json:"eventName"`
		EventVersion string `json:"eventVersion"`
		EventSource  string `json:"eventSource"`
		AwsRegion    string `json:"awsRegion"`
		Dynamodb     struct {
			OldImage struct {
				LastKnownStatus struct {
					S string `json:"S"`
				} `json:"LastKnownStatus"`
			} `json:"OldImage"`
			NewImage struct {
				LastKnownStatus struct {
					S string `json:"S"`
				} `json:"LastKnownStatus"`
				Location struct {
					S string `json:"S"`
				} `json:"Location"`
				CallReason struct {
					S string `json:"S"`
				} `json:"CallReason"`
			} `json:"NewImage"`
		} `json:"dynamodb"`
	} `json:"Records"`
}

func SendSms(message string) error {
	params := &openapi.CreateMessageParams{}
	params.SetTo(toNumber)
	params.SetFrom(fromNumber)
	params.SetBody(message)

	_, err := twilioClient.ApiV2010.CreateMessage(params)
	return err
}

func HandleRequest(ctx context.Context, event StreamRecord) error {
	for _, record := range event.Records {
		newImage := record.Dynamodb.NewImage
		oldImage := record.Dynamodb.OldImage

		if oldImage.LastKnownStatus.S != newImage.LastKnownStatus.S {
			message := fmt.Sprintf(
				"Active call alert at %v: %v, %v",
				newImage.Location.S,
				newImage.CallReason.S,
				newImage.LastKnownStatus.S,
			)
			err := SendSms(message)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
