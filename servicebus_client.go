package main

import (
	"context"

	servicebus "github.com/Azure/azure-service-bus-go"
)

type stats struct {
	queues *[]queueStats
	topics *[]topicStats
}

type messageCounts struct {
	activeMessages             int32
	deadLetterMessages         int32
	scheduledMessages          int32
	transferDeadLetterMessages int32
	transferMessages           int32
}

type sizes struct {
	sizeInBytes    int64
	maxSizeInBytes int64
}

type queueStats struct {
	name string
	messageCounts
	sizes
}

type subscriptionStats struct {
	name string
	messageCounts
}
type topicStats struct {
	name string
	messageCounts
	sizes

	subscriptions *[]subscriptionStats
}

func getServiceBusStats() (*stats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ns, err := servicebus.NewNamespace(servicebus.NamespaceWithConnectionString(connectionString))
	if err != nil {
		return nil, err
	}

	queues, err := getQueueStats(ctx, ns)
	if err != nil {
		return nil, err
	}

	topics, err := getTopicStats(ctx, ns)
	if err != nil {
		return nil, err
	}

	return &stats{
		queues: queues,
		topics: topics,
	}, nil

}

func getQueueStats(ctx context.Context, ns *servicebus.Namespace) (*[]queueStats, error) {
	var result []queueStats

	queueManager := ns.NewQueueManager()

	queues, err := queueManager.List(ctx)
	if err != nil {
		return nil, err
	}

	for _, queue := range queues {
		result = append(result, queueStats{
			name: queue.Name,
			messageCounts: messageCounts{
				activeMessages:             *queue.CountDetails.ActiveMessageCount,
				deadLetterMessages:         *queue.CountDetails.DeadLetterMessageCount,
				scheduledMessages:          *queue.CountDetails.ScheduledMessageCount,
				transferDeadLetterMessages: *queue.CountDetails.TransferDeadLetterMessageCount,
				transferMessages:           *queue.CountDetails.TransferMessageCount,
			},
			sizes: sizes{
				sizeInBytes:    *queue.QueueDescription.SizeInBytes,
				maxSizeInBytes: int64(*queue.QueueDescription.MaxSizeInMegabytes) * 1024 * 1024,
			},
		})
	}

	return &result, nil
}

func getTopicStats(ctx context.Context, ns *servicebus.Namespace) (*[]topicStats, error) {
	var result []topicStats

	topicManager := ns.NewTopicManager()

	topics, err := topicManager.List(ctx)
	if err != nil {
		return nil, err
	}

	for _, topic := range topics {

		subs, err := getSubscriptionStats(ctx, ns, topic.Name)
		if err != nil {
			return nil, err
		}

		result = append(result, topicStats{
			name:          topic.Name,
			messageCounts: countDetailsToMessageCounts(topic.CountDetails),
			sizes: sizes{
				sizeInBytes:    *topic.TopicDescription.SizeInBytes,
				maxSizeInBytes: int64(*topic.TopicDescription.MaxSizeInMegabytes) * 1024 * 1024,
			},
			subscriptions: subs,
		})
	}

	return &result, nil
}

func getSubscriptionStats(ctx context.Context, ns *servicebus.Namespace, topicName string) (*[]subscriptionStats, error) {
	var result []subscriptionStats

	subsManager, err := ns.NewSubscriptionManager(topicName)
	if err != nil {
		return nil, err
	}

	subs, err := subsManager.List(ctx)
	if err != nil {
		return nil, err
	}

	for _, sub := range subs {
		result = append(result, subscriptionStats{
			name:          sub.Name,
			messageCounts: countDetailsToMessageCounts(sub.CountDetails),
		})
	}

	return &result, nil
}

func countDetailsToMessageCounts(countDetails *servicebus.CountDetails) messageCounts {
	return messageCounts{
		activeMessages:             *countDetails.ActiveMessageCount,
		deadLetterMessages:         *countDetails.DeadLetterMessageCount,
		scheduledMessages:          *countDetails.ScheduledMessageCount,
		transferDeadLetterMessages: *countDetails.TransferDeadLetterMessageCount,
		transferMessages:           *countDetails.TransferMessageCount,
	}
}
