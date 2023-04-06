# Stanza Go SDK

Stanza is a... <one sentence summary here>

"Stanza Go SDK" provides higher-order functions ("decorators") for adding Stanza fault tolerance to your Golang application.

## Installation

Stanza's `sdk-go` can be installed like any other Go library via `go get`:

```shell
$ go get github.com/StanzaSystems/sdk-go
```
NOTE: while this module is under development, you may get `remote: Repository not found` errors when trying to run `go get`. To work around this, you can follow the advice in this [thread](https://stackoverflow.com/questions/27500861/whats-the-proper-way-to-go-get-a-private-repository) which suggests running the command:
* `git config --global url.git@github.com:.insteadOf https://github.com/`
While this works, there is an updated form which is recommended:
* `go env -w GOPRIVATE=github.com/StanzaSystems/*` - [Source](https://stackoverflow.com/questions/58305567/how-to-set-goprivate-environment-variable)
(FWIW the former worked, the latter did not for some reason)
  
## Configuration

To use `sdk-go`, you'll need to import the `sdk-go` package and initialize it with
your local Stanza Hub and any other options.

## Usage

The SDK supports adding flow control, traffic shaping, concurrency limiting, circuit breaking, and adaptive system protection (via [Sentinel](https://github.com/alibaba/sentinel-golang)) to your service and externally managing the configs with the Stanza Control Plane.

See samples to get started. (TODO)

Documentation is available here. You can also find the API documentation here.

## Community

Join Stanza's <something> to get involved and help us improve the SDK!
