package harvester

import (
	"context"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/kevin-secrist/cfactivecallmonitor/internal/chesterfield"
	"github.com/kevin-secrist/cfactivecallmonitor/internal/saved_calls"
)

type Harvester struct {
	apiClient chesterfield.Client
	dao       saved_calls.Client
}

func New(policeApiKey string, fireApiKey string, cfg aws.Config) *Harvester {
	return &Harvester{
		apiClient: chesterfield.New(policeApiKey, fireApiKey),
		dao:       saved_calls.New(cfg),
	}
}

func NewWithClients(apiClient chesterfield.Client, dao saved_calls.Client) *Harvester {
	return &Harvester{
		apiClient: apiClient,
		dao:       dao,
	}
}

type CallResult struct {
	Calls chesterfield.CallForService
	Err   error
}

type SavedCallResult struct {
	Calls []saved_calls.SavedCall
	Err   error
}

var streetNameRegex = regexp.MustCompile(`(?:(\d+XX) )?(.*)`)

func (harvester *Harvester) updateCalls(ctx context.Context, callType string, activeCalls chesterfield.CallForService, savedCalls []saved_calls.SavedCall) error {
	callMap := map[string]saved_calls.SavedCall{}
	for _, call := range savedCalls {
		if call.CallType == callType {
			callMap[call.ID] = call
		}
	}

	for _, activeCall := range activeCalls {
		match := streetNameRegex.FindStringSubmatch(activeCall.Location)
		savedCall := saved_calls.SavedCall{
			ID:              activeCall.ID,
			CallType:        callType,
			CallReason:      activeCall.Type,
			LastKnownStatus: activeCall.CurrentStatus,
			CallReceived:    activeCall.CallReceived.Time,
			Location:        activeCall.Location,
			Area:            activeCall.Area,
			Priority:        activeCall.Priority,
			HouseNumber:     match[1],
			StreetName:      match[2],
		}
		if existingCall, ok := callMap[savedCall.ID]; ok {
			delete(callMap, savedCall.ID)
			if existingCall.LastKnownStatus != savedCall.LastKnownStatus {
				err := harvester.dao.UpdateStatus(ctx, savedCall)
				if err != nil {
					return err
				}
			}
		} else {
			err := harvester.dao.SaveCall(ctx, savedCall)
			if err != nil {
				return err
			}
		}
	}

	for _, resolvedCall := range callMap {
		resolvedCall.LastKnownStatus = "resolved"
		err := harvester.dao.UpdateStatus(ctx, resolvedCall)
		if err != nil {
			return err
		}
	}

	return nil
}

func (harvester *Harvester) Harvest(ctx context.Context) error {
	policeCallsCh := make(chan CallResult)
	fireCallsCh := make(chan CallResult)
	savedCallsCh := make(chan SavedCallResult)

	go func() {
		defer close(policeCallsCh)
		log.Println("Retrieving Police Calls")
		calls, err := harvester.apiClient.GetPoliceCalls()
		policeCallsCh <- CallResult{
			Calls: calls,
			Err:   err,
		}
		log.Printf("Found %d Police Calls, %+v\n", len(calls), err)
	}()

	go func() {
		defer close(fireCallsCh)
		log.Println("Retrieving Fire Calls")
		calls, err := harvester.apiClient.GetFireCalls()
		fireCallsCh <- CallResult{
			Calls: calls,
			Err:   err,
		}
		log.Printf("Found %d Fire Calls, %+v\n", len(calls), err)
	}()

	go func() {
		defer close(savedCallsCh)
		log.Println("Retrieving Saved Calls")
		calls, err := harvester.dao.GetActiveCalls(ctx)
		savedCallsCh <- SavedCallResult{
			Calls: calls,
			Err:   err,
		}
		log.Printf("Found %d Saved Calls, %+v\n", len(calls), err)
	}()

	policeCalls, fireCalls, savedCalls := <-policeCallsCh, <-fireCallsCh, <-savedCallsCh

	if policeCalls.Err != nil {
		return policeCalls.Err
	}
	if fireCalls.Err != nil {
		return fireCalls.Err
	}
	if savedCalls.Err != nil {
		return savedCalls.Err
	}

	log.Println("Updating Police Calls")
	err := harvester.updateCalls(ctx, "police", policeCalls.Calls, savedCalls.Calls)
	if err != nil {
		log.Printf("Encountered error while updating police calls, %+v\n", err)
		return err
	}
	log.Println("Updating Fire Calls")
	err = harvester.updateCalls(ctx, "fire", fireCalls.Calls, savedCalls.Calls)
	if err != nil {
		log.Printf("Encountered error while updating fire calls, %+v\n", err)
		return err
	}

	log.Println("Completed Harvest")
	return nil
}
