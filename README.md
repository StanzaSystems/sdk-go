# Stanza Go SDK

Stanza is a developer-oriented system for increasing reliability based on traffic management and prioritisation, quota allocation, and rate-limiting. On the back-end, it helps prevent problems related to overload and excessive use of third-party APIs. On the front-end, it helps you to segment and prioritise your traffic, react automatically to overload or other conditions, and inspect the state of your critical user journeys.

"Stanza Go SDK" provides higher-order functions ("decorators") for adding Stanza fault tolerance to your Golang application.

## Installation

Stanza's `sdk-go` can be installed like any other Go library via `go get`:

```shell
$ go get github.com/StanzaSystems/sdk-go
```
  
## Configuration

To use `sdk-go`, you'll need to import the `sdk-go` package and initialize it with
your local Stanza Hub and any other options.

## Usage

The SDK supports adding flow control, traffic shaping, concurrency limiting, circuit breaking, and adaptive system protection (via [Sentinel](https://github.com/alibaba/sentinel-golang)) to your service and externally managing the configs with the Stanza Control Plane.

See samples to get started. (TODO)

Documentation is available here. You can also find the API documentation here.

## Community

Join Stanza's <something> to get involved and help us improve the SDK!
