package client

import (
	"context"
	"time"

	azservicebus "github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus/admin"
)

type ServiceBusClient struct {
	connectionString string
	timeout          time.Duration
}

type Stats struct {
	Queues *[]QueueStats
	Topics *[]TopicStats
}

type MessageCounts struct {
	ActiveMessages             int32
	DeadLetterMessages         int32
	ScheduledMessages          int32
	TransferDeadLetterMessages int32
	TransferMessages           int32
}

type Sizes struct {
	SizeInBytes    int64
	MaxSizeInBytes int64
}

type QueueStats struct {
	Name string
	MessageCounts
	Sizes
}

type SubscriptionStats struct {
	Name string
	MessageCounts
}
type TopicStats struct {
	Name string
	MessageCounts
	Sizes

	Subscriptions *[]SubscriptionStats
}

func New(connectionString string, timeout time.Duration) *ServiceBusClient {
	return &ServiceBusClient{
		connectionString: connectionString,
		timeout:          timeout,
	}
}

func (c *ServiceBusClient) GetServiceBusStats() (*Stats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	client, err := azservicebus.NewClientFromConnectionString(c.connectionString, nil)
	//ns, err := servicebus.NewNamespace(servicebus.NamespaceWithConnectionString(c.connectionString))
	if err != nil {
		return nil, err
	}

	queues, err := getQueueStats(ctx, client)
	if err != nil {
		return nil, err
	}

	topics, err := getTopicStats(ctx, client)
	if err != nil {
		return nil, err
	}

	return &Stats{
		Queues: queues,
		Topics: topics,
	}, nil

}

func getQueueStats(ctx context.Context, client *azservicebus.Client) (*[]QueueStats, error) {
	var result []QueueStats

	pager := client.NewListQueuesPager(&azservicebus.ListQueuesOptions{MaxPageSize: 50})

	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, err
		}
		for _, queue := range resp.Queues {
			queueStats, err := client.GetQueueRuntimeProperties(ctx, queue.QueueName, &azservicebus.GetQueueRuntimePropertiesOptions{})
			if err != nil {
				continue
			}
			result = append(result, QueueStats{
				Name: queue.QueueName,
				MessageCounts: MessageCounts{
					ActiveMessages:             queueStats.ActiveMessageCount,
					DeadLetterMessages:         queueStats.DeadLetterMessageCount,
					ScheduledMessages:          queueStats.ScheduledMessageCount,
					TransferDeadLetterMessages: queueStats.TransferDeadLetterMessageCount,
					TransferMessages:           queueStats.TransferMessageCount,
				},
				Sizes: Sizes{
					SizeInBytes:    queueStats.SizeInBytes,
					MaxSizeInBytes: int64(*queue.MaxSizeInMegabytes) * 1024 * 1024,
				},
			})
		}
	}

	return &result, nil
}

func getTopicStats(ctx context.Context, client *azservicebus.Client) (*[]TopicStats, error) {
	var result []TopicStats

	pager := client.NewListTopicsPager(&azservicebus.ListTopicsOptions{MaxPageSize: 20})

	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, err
		}
		for _, topic := range resp.Topics {
			topicStats, err := client.GetTopicRuntimeProperties(ctx, topic.TopicName, &azservicebus.GetTopicRuntimePropertiesOptions{})
			if err != nil {
				continue
			}
			stats, err := getSubscriptionStats(ctx, client, topic.TopicName)
			if err != nil {
				continue
			}
			topicCounts := MessageCounts{
				ActiveMessages:             0,
				DeadLetterMessages:         0,
				ScheduledMessages:          topicStats.ScheduledMessageCount,
				TransferDeadLetterMessages: 0,
				TransferMessages:           0,
			}
			for _, stat := range *stats {
				topicCounts.ActiveMessages += stat.ActiveMessages
				topicCounts.DeadLetterMessages += stat.DeadLetterMessages
				topicCounts.TransferDeadLetterMessages += stat.TransferDeadLetterMessages
				topicCounts.TransferMessages += stat.TransferMessages
			}

			result = append(result, TopicStats{
				Name:          topic.TopicName,
				MessageCounts: topicCounts,
				Sizes: Sizes{
					SizeInBytes:    topicStats.SizeInBytes,
					MaxSizeInBytes: int64(*topic.MaxSizeInMegabytes) * 1024 * 1024,
				},
				Subscriptions: stats,
			})
		}
	}

	return &result, nil
}

func getSubscriptionStats(ctx context.Context, client *azservicebus.Client, topicName string) (*[]SubscriptionStats, error) {
	var result []SubscriptionStats

	pager := client.NewListSubscriptionsPager(topicName, &azservicebus.ListSubscriptionsOptions{MaxPageSize: 50})

	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, err
		}
		for _, sub := range resp.Subscriptions {
			subDetails, err := client.GetSubscriptionRuntimeProperties(ctx, topicName, sub.SubscriptionName, &azservicebus.GetSubscriptionRuntimePropertiesOptions{})
			if err != nil {
				return nil, err
			}
			result = append(result, SubscriptionStats{
				Name:          sub.SubscriptionName,
				MessageCounts: countDetailsToMessageCounts(subDetails),
			})
		}
	}
	return &result, nil
}

func countDetailsToMessageCounts(countDetails *azservicebus.GetSubscriptionRuntimePropertiesResponse) MessageCounts {
	return MessageCounts{
		ActiveMessages:             countDetails.ActiveMessageCount,
		DeadLetterMessages:         countDetails.DeadLetterMessageCount,
		TransferDeadLetterMessages: countDetails.TransferDeadLetterMessageCount,
		TransferMessages:           countDetails.TransferMessageCount,
	}
}
