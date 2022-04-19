package chesterfield

import (
	"github.com/go-resty/resty/v2"
)

const (
	baseURL = "https://api.chesterfield.gov/api"
)

type ChesterfieldAPIClient struct {
	RestClient   *resty.Client
	policeApiKey string
	fireApiKey   string
}

type Client interface {
	GetPoliceCalls() (CallForService, error)
	GetFireCalls() (CallForService, error)
}

func New(policeApiKey string, fireApiKey string) *ChesterfieldAPIClient {
	restClient := resty.New().
		SetBaseURL(baseURL).
		SetHeader("Referer", "https://www.chesterfield.gov/").
		SetRetryCount(1)

	return &ChesterfieldAPIClient{
		RestClient:   restClient,
		policeApiKey: policeApiKey,
		fireApiKey:   fireApiKey,
	}
}
