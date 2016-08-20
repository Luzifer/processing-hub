# Luzifer / processing-hub

The `processing-hub` is a hub for sending messages between different services. It supports several input webhooks and has some built in outputs. For everything else you can add own script written in Javascript. Currently its not supported to create own inputs but you can send data to every URL reachable over HTTP(s).

## Facts

- Messages are simple maps / objects specified as `map[string]string`
- Messages are stored inside Amazon SQS and retried as often as you configure the queue to do so
- The content of the message depends on the input hook, the only required fields are `_type` and `_date` which are set when creating the message
- Message types must have the format of a reverse FQDN
- The scripts can derive a new message after parsing the inbound message. That way you can create intermediate scripts to pre-process messages before they are processed by an output

## Available inputs

- `/generic/<type>` - Creates a message using the defined type with the passed form values as fields inside the message. Useful for testing own message processors.

## Available outputs

- `io.luzifer.outputs.log` - Dumps the whole message into the log of the processing-hub
- `io.luzifer.outputs.pushover` - Sends a Pushover.net message
  - Parameters: Supports all parameters of the Pushover.net API. Required parameters are `message` and `user`.
  - Configuration values:
    - `outputs/pushover/token` - Application token for the Pushover.net API

## Available configuration backends

- `env` - Reads configurations from the environment.  
Paths are translated as following: `outputs/pushover/token => OUTPUTS_PUSHOVER_TOKEN`
