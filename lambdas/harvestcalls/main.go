package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"

	"github.com/kevin-secrist/cfactivecallmonitor/internal/harvester"
)

var harvesterInstance *harvester.Harvester

func init() {
	policeApiKey := os.Getenv("CPD_API_KEY")
	fireApiKey := os.Getenv("CFD_API_KEY")

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("unable to load aws config")
	}
	harvesterInstance = harvester.New(policeApiKey, fireApiKey, cfg)
}

func HandleRequest(ctx context.Context) error {
	return harvesterInstance.Harvest(ctx)
}

func main() {
	lambda.Start(HandleRequest)
}
