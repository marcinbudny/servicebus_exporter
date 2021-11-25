package client

import (
	"context"
	"sync"
	"time"

	azservicebus "github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus/admin"
	"k8s.io/klog/v2"
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
	TotalMessageCount          int64
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

	client, err := azservicebus.NewClientFromConnectionString(c.connectionString, &azservicebus.ClientOptions{})
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

	queues := client.ListQueues(&azservicebus.ListQueuesOptions{MaxPageSize: 50})
	for {
		gotPage := queues.NextPage(ctx)
		if !gotPage {
			break
		}
		klog.V(3).InfoS("loading queue page", "itemCount", len(queues.PageResponse().Items))

		for _, queue := range queues.PageResponse().Items {
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

	topics := client.ListTopics(&azservicebus.ListTopicsOptions{MaxPageSize: 25})

	for {
		gotPage := topics.NextPage(ctx)
		if !gotPage {
			break
		}
		klog.V(3).InfoS("loading topic page", "itemCount", len(topics.PageResponse().Items))

		var wg sync.WaitGroup
		var mutex sync.Mutex
		s := make([][]SubscriptionStats, len(topics.PageResponse().Items))
		for i, topic := range topics.PageResponse().Items {
			wg.Add(1)

			go func(topicName string, output [][]SubscriptionStats, i int) {
				defer wg.Done()
				subStats, err := getSubscriptionStats(ctx, client, topicName)
				if err != nil {
					return
				}

				mutex.Lock()
				defer mutex.Unlock()
				output[i] = *subStats
			}(topic.TopicName, s, i)
		}

		wg.Wait()

		for i, topic := range topics.PageResponse().Items {
			topicStats, err := client.GetTopicRuntimeProperties(ctx, topic.TopicName, &azservicebus.GetTopicRuntimePropertiesOptions{})
			if err != nil {
				continue
			}

			sub := &s[i]

			result = append(result, TopicStats{
				Name:          topic.TopicName,
				MessageCounts: MessageCounts{},
				Sizes: Sizes{
					SizeInBytes:    topicStats.SizeInBytes,
					MaxSizeInBytes: int64(*topic.MaxSizeInMegabytes) * 1024 * 1024,
				},
				Subscriptions: sub,
			})
		}
	}

	return &result, nil
}

func getSubscriptionStats(ctx context.Context, client *azservicebus.Client, topicName string) (*[]SubscriptionStats, error) {
	var result []SubscriptionStats

	subs := client.ListSubscriptions(topicName, &azservicebus.ListSubscriptionsOptions{MaxPageSize: 50})

	for {
		gotPage := subs.NextPage(ctx)
		if !gotPage {
			break
		}
		klog.V(3).InfoS("loading topic subscription page", "itemCount", len(subs.PageResponse().Items))

		for _, sub := range subs.PageResponse().Items {
			runtimeProps, err := client.GetSubscriptionRuntimeProperties(ctx, topicName, sub.SubscriptionName, &azservicebus.GetSubscriptionRuntimePropertiesOptions{})
			if err != nil {
				continue
			}

			result = append(result, SubscriptionStats{
				Name:          sub.SubscriptionName,
				MessageCounts: countDetailsToMessageCounts(runtimeProps),
			})
		}
	}

	return &result, nil
}

func countDetailsToMessageCounts(countDetails *azservicebus.GetSubscriptionRuntimePropertiesResponse) MessageCounts {
	return MessageCounts{
		ActiveMessages:             countDetails.ActiveMessageCount,
		DeadLetterMessages:         countDetails.DeadLetterMessageCount,
		TotalMessageCount:          countDetails.TotalMessageCount,
		TransferDeadLetterMessages: countDetails.TransferDeadLetterMessageCount,
		TransferMessages:           countDetails.TransferMessageCount,
	}
}
