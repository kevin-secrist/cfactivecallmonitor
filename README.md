# Active Calls Application

## History

This project is a rewrite of an existing application I wrote in 2019. The existing application is an Azure Function which polls the API for [this site](https://www.chesterfield.gov/3999/Active-Police-Calls) for active police calls, i.e. events that police are dispatched to. When an event that matches some criteria is found, it will send me a text message with the details, e.g. an active call on my street. The application has been running solidly for 3 years and for the vast majority of that time only cost $0.01/mo. Within the last few months (in 2022) the activity on the storage account has increased and costs a few pennies more per month.

The purpose of this rewrite is to try out Go for the first time, move my infrastructure over to AWS, and create some example code that I can publicly share. While the function of the app is very simple, it has an enormous potential for new features and additions.

## Architecture

```mermaid
sequenceDiagram
    participant Harvester as Lambda Harvester
    participant DB as DynamoDB
    participant Notifier as Lambda Notifier
    participant Twilio
    participant CPD as Chesterfield Service Calls API
    Harvester->>+CPD: GET https://api.chesterfield.gov/api/Police/V1.0/Calls/CallsForService
    Harvester->>+DB: Query for Stored Active Calls
    DB-->>-Harvester: Stored Service Calls
    CPD->>-Harvester: Currently Active Service Calls
    Harvester->>Harvester: Compare Active Calls
    note right of Harvester: Any Service Calls not returned by <br/> the API are marked resolved
    Harvester->>+DB: Store Updates
    DB-->>-Harvester: Updates Processed
    DB-)Notifier: DynamoDB Stream Trigger
    note right of Notifier: Events filtered based <br /> on message fields
    Notifier-)Twilio: Send SMS
```

## To Do

* Monitoring
* GraphQL API and UI for visualizing service calls
* More flexible subscription model, via SNS
  * Multi-user support
  * Backend store for subscriptions and users
* End-user configurable subscriptions, with UI/API
* Additional unit test coverage for lambda notifier
  * This part was thrown together in a rush