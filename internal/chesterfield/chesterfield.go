package chesterfield

import (
	"github.com/go-resty/resty/v2"
)

const (
	baseURL = "https://api.chesterfield.gov/api"
)

type ChesterfieldAPIClient struct {
	RestClient *resty.Client
	apiKey     string
}

func New(apiKey string) *ChesterfieldAPIClient {
	restClient := resty.New().
		SetHostURL(baseURL).
		SetAuthToken(apiKey).
		SetHeader("Accept", "application/json").
		SetHeader("Referer", "https://www.chesterfield.gov/").
		SetRetryCount(1)

	return &ChesterfieldAPIClient{
		RestClient: restClient,
		apiKey:     apiKey,
	}
}
