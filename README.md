# Cmd-Notify

A super simple command for wrapping long running tasks. Was developed out of frustration while configuring long running build and release processes locally on my own computer. It only integrates with pushover but open to other send method integrations. Future work to make a send interface that makes sense so it can be plugged in.

## Configuration

Simple configuration
```
export PUSHOVER_API_TOKEN=<api token from pushover.net>
export PUSHOVER_USER_TOKEN=<user or group token from pushover.net>
# Optional debug for outputting any non-critical information from the wrapping command
export GO_NOTIFY_DEBUG=true
```

## Installation
```
go get github.com/kwyn/cmd-notify
```

## Usage
```sh
cmd-notify docker build .
```

Will run docker build and notify you once the command is complete.

Once the build is done it'll send you a notification via pushover.

## Development
Special env variable can be set to skip the send step when debugging logic:
```
GO_NOTIFY_SKIP_SEND=true
```
