package chesterfield

import (
	"fmt"
)

type CallForService []struct {
	ID                    string     `json:"id,omitempty"`
	CallReceived          CustomTime `json:"callReceived,omitempty"`
	Location              string     `json:"location,omitempty"`
	Type                  string     `json:"type,omitempty"`
	CurrentStatus         string     `json:"currentStatus,omitempty"`
	Area                  string     `json:"area,omitempty"`
	Priority              string     `json:"priority,omitempty"`
	CallReceivedFormatted string     `json:"callReceivedFormatted,omitempty"`
}

func (client *ChesterfieldAPIClient) getServiceCalls(service string, version string, authHeaderKey string, authHeaderValue string) (CallForService, error) {
	var result CallForService
	response, err := client.RestClient.R().
		SetResult(&result).
		SetHeader(authHeaderKey, authHeaderValue).
		SetPathParams(map[string]string{
			"Service": service,
			"Version": version,
		}).
		Get("{Service}/{Version}/Calls/CallsForService")

	if err != nil {
		return nil, err
	}

	if response.IsError() {
		return nil, fmt.Errorf("received invalid status code: %d", response.StatusCode())
	}

	slice := response.Result().(*CallForService)
	return *slice, nil
}

// GET https://api.chesterfield.gov/api/Police/V1.1/Calls/CallsForService
func (client *ChesterfieldAPIClient) GetPoliceCalls() (CallForService, error) {
	return client.getServiceCalls("Police", "V1.1", "X-Apikey", client.policeApiKey)
}

// GET https://api.chesterfield.gov/api/Fire/V1.0/Calls/CallsForService
func (client *ChesterfieldAPIClient) GetFireCalls() (CallForService, error) {
	return client.getServiceCalls("Fire", "V1.0", "X-Apikey", client.fireApiKey)
}
