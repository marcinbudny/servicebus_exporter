# Azure Service Bus Prometheus exporter - Helm Chart

This is a fork from https://github.com/marcinbudny/servicebus_exporter, just for the Helm Chart.

See [the chart readme](https://github.com/giggio/servicebus_exporter/blob/helmchart/charts/servicebusexporter/README.md).

# Azure Service Bus Prometheus exporter
Azure Service Bus metrics Prometheus exporter. Does not use Azure Monitor, connects to the Service Bus directly and scrapes all queues, topics and subscriptions.

## Installation

### From source

You need to have a Go 1.10+ environment configured. Clone the repo (outside your `GOPATH`) and then:

```bash
go build -o servicebus_exporter
./servicebus_exporter --connection-string=[YOUR CONNECTION STRING]
```

### Using Docker

```bash
docker run -d -p 9580:9580 -e CONNECTION_STRING=[YOUR CONNECTION STRING] marcinbudny/servicebus_exporter
```

## Configuration

The exporter can be configured with commandline arguments, environment variables and a configuration file. For the details on how to format the configuration file, visit [namsral/flag](https://github.com/namsral/flag) repo.

|Flag|ENV variable|Default|Meaning|
|---|---|---|---|
|--connection-string|CONNECTION_STRING|_no default_|Connection string for Azure Service Bus. The exporter requires `Manage` SAS policy.|
|--port|PORT|9580|Port to expose scrape endpoint on|
|--timeout|TIMEOUT|30s|Timeout when scraping the Service Bus|
|--verbose|VERBOSE|false|Enable verbose logging|


## Exported metrics

See [here](https://docs.microsoft.com/en-us/azure/service-bus-messaging/message-counters#message-count-details) for explanation of different message count types.

```
# HELP servicebus_queue_active_messages Queue active messages count
# TYPE servicebus_queue_active_messages gauge
servicebus_queue_active_messages{name="somequeue"} 0
# HELP servicebus_queue_dead_letter_messages Queue dead letter messages count
# TYPE servicebus_queue_dead_letter_messages gauge
servicebus_queue_dead_letter_messages{name="somequeue"} 0
# HELP servicebus_queue_max_size_bytes Queue maximum size in bytes
# TYPE servicebus_queue_max_size_bytes gauge
servicebus_queue_max_size_bytes{name="somequeue"} 1.073741824e+09
# HELP servicebus_queue_scheduled_messages Queue scheduled messages count
# TYPE servicebus_queue_scheduled_messages gauge
servicebus_queue_scheduled_messages{name="somequeue"} 0
# HELP servicebus_queue_size_bytes Queue size in bytes
# TYPE servicebus_queue_size_bytes gauge
servicebus_queue_size_bytes{name="somequeue"} 0
# HELP servicebus_queue_transfer_dead_letter_messages Queue transfer dead letter messages count
# TYPE servicebus_queue_transfer_dead_letter_messages gauge
servicebus_queue_transfer_dead_letter_messages{name="somequeue"} 0
# HELP servicebus_queue_transfer_messages Queue transfer messages count
# TYPE servicebus_queue_transfer_messages gauge
servicebus_queue_transfer_messages{name="somequeue"} 0

# HELP servicebus_subscription_active_messages Subscription active messages count
# TYPE servicebus_subscription_active_messages gauge
servicebus_subscription_active_messages{name="somesubscription",topic_name="sometopic"} 0
# HELP servicebus_subscription_dead_letter_messages Subscription dead letter messages count
# TYPE servicebus_subscription_dead_letter_messages gauge
servicebus_subscription_dead_letter_messages{name="somesubscription",topic_name="sometopic"} 0
# HELP servicebus_subscription_scheduled_messages Subscription scheduled messages count
# TYPE servicebus_subscription_scheduled_messages gauge
servicebus_subscription_scheduled_messages{name="somesubscription",topic_name="sometopic"} 0
# HELP servicebus_subscription_transfer_dead_letter_messages Subscription transfer dead letter messages count
# TYPE servicebus_subscription_transfer_dead_letter_messages gauge
servicebus_subscription_transfer_dead_letter_messages{name="somesubscription",topic_name="sometopic"} 0
# HELP servicebus_subscription_transfer_messages Subscription transfer messages count
# TYPE servicebus_subscription_transfer_messages gauge
servicebus_subscription_transfer_messages{name="somesubscription",topic_name="sometopic"} 0

# HELP servicebus_topic_active_messages Topic active messages count
# TYPE servicebus_topic_active_messages gauge
servicebus_topic_active_messages{name="sometopic"} 0
# HELP servicebus_topic_dead_letter_messages Topic dead letter messages count
# TYPE servicebus_topic_dead_letter_messages gauge
servicebus_topic_dead_letter_messages{name="sometopic"} 0
# HELP servicebus_topic_max_size_bytes Topic maximum size in bytes
# TYPE servicebus_topic_max_size_bytes gauge
servicebus_topic_max_size_bytes{name="sometopic"} 1.073741824e+09
# HELP servicebus_topic_scheduled_messages Topic scheduled messages count
# TYPE servicebus_topic_scheduled_messages gauge
servicebus_topic_scheduled_messages{name="sometopic"} 0
# HELP servicebus_topic_size_bytes Topic size in bytes
# TYPE servicebus_topic_size_bytes gauge
servicebus_topic_size_bytes{name="sometopic"} 0
# HELP servicebus_topic_transfer_dead_letter_messages Topic transfer dead letter messages count
# TYPE servicebus_topic_transfer_dead_letter_messages gauge
servicebus_topic_transfer_dead_letter_messages{name="sometopic"} 0
# HELP servicebus_topic_transfer_messages Topic transfer messages count
# TYPE servicebus_topic_transfer_messages gauge
servicebus_topic_transfer_messages{name="sometopic"} 0

# HELP servicebus_up Whether the Azure ServiceBus scrape was successful
# TYPE servicebus_up gauge
servicebus_up 1

```
