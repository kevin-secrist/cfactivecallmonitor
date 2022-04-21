package main

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"

	"github.com/kevin-secrist/cfactivecallmonitor/internal/harvester"
)

func main() {
	policeApiKey := os.Getenv("CPD_API_KEY")
	fireApiKey := os.Getenv("CFD_API_KEY")

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("unable to load aws config")
	}
	harvesterInstance := harvester.New(policeApiKey, fireApiKey, cfg)
	err = harvesterInstance.Harvest(context.TODO())
	if err != nil {
		panic(err)
	}
}
