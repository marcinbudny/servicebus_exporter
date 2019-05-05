package main

import (
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type messageCountMetrics struct {
	activeMessages             *prometheus.Desc
	deadLetterMessages         *prometheus.Desc
	scheduledMessages          *prometheus.Desc
	transferDeadLetterMessages *prometheus.Desc
	transferMessages           *prometheus.Desc
}

type sizeMetrics struct {
	size    *prometheus.Desc
	maxSize *prometheus.Desc
}

type exporter struct {
	up *prometheus.Desc

	queueMessageCounts messageCountMetrics
	queueSizes         sizeMetrics

	topicMessageCounts messageCountMetrics
	topicSizes         sizeMetrics

	subscriptionMessageCounts messageCountMetrics
}

func newMessageCountMetrics(itemName string, labels ...string) messageCountMetrics {
	return messageCountMetrics{
		activeMessages: prometheus.NewDesc(fmt.Sprintf("servicebus_%s_active_messages", strings.ToLower(itemName)),
			fmt.Sprintf("%s active messages count", itemName), labels, nil),
		deadLetterMessages: prometheus.NewDesc(fmt.Sprintf("servicebus_%s_dead_letter_messages", strings.ToLower(itemName)),
			fmt.Sprintf("%s dead letter messages count", itemName), labels, nil),
		scheduledMessages: prometheus.NewDesc(fmt.Sprintf("servicebus_%s_scheduled_messages", strings.ToLower(itemName)),
			fmt.Sprintf("%s scheduled messages count", itemName), labels, nil),
		transferMessages: prometheus.NewDesc(fmt.Sprintf("servicebus_%s_transfer_messages", strings.ToLower(itemName)),
			fmt.Sprintf("%s transfer messages count", itemName), labels, nil),
		transferDeadLetterMessages: prometheus.NewDesc(fmt.Sprintf("servicebus_%s_transfer_dead_letter_messages", strings.ToLower(itemName)),
			fmt.Sprintf("%s transfer dead letter messages count", itemName), labels, nil),
	}
}

func newSizeMetrics(itemName string) sizeMetrics {
	return sizeMetrics{
		size: prometheus.NewDesc(fmt.Sprintf("servicebus_%s_size_bytes", strings.ToLower(itemName)),
			fmt.Sprintf("%s size in bytes", itemName), []string{"name"}, nil),
		maxSize: prometheus.NewDesc(fmt.Sprintf("servicebus_%s_max_size_bytes", strings.ToLower(itemName)),
			fmt.Sprintf("%s maximum size in bytes", itemName), []string{"name"}, nil),
	}
}

func newExporter() *exporter {
	return &exporter{
		up: prometheus.NewDesc("servicebus_up", "Whether the Azure ServiceBus scrape was successful", nil, nil),

		queueMessageCounts: newMessageCountMetrics("Queue", "name"),
		queueSizes:         newSizeMetrics("Queue"),

		topicMessageCounts: newMessageCountMetrics("Topic", "name"),
		topicSizes:         newSizeMetrics("Topic"),

		subscriptionMessageCounts: newMessageCountMetrics("Subscription", "name", "topicName"),
	}
}

func describeMessageCountMetrics(ch chan<- *prometheus.Desc, metrics *messageCountMetrics) {
	ch <- metrics.activeMessages
	ch <- metrics.deadLetterMessages
	ch <- metrics.scheduledMessages
	ch <- metrics.transferMessages
	ch <- metrics.transferDeadLetterMessages
}

func describeSizeMetrics(ch chan<- *prometheus.Desc, metrics *sizeMetrics) {
	ch <- metrics.size
	ch <- metrics.maxSize
}

func (e *exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.up

	describeMessageCountMetrics(ch, &e.queueMessageCounts)
	describeSizeMetrics(ch, &e.queueSizes)

	describeMessageCountMetrics(ch, &e.topicMessageCounts)
	describeSizeMetrics(ch, &e.topicSizes)

	describeMessageCountMetrics(ch, &e.subscriptionMessageCounts)
}

func (e *exporter) Collect(ch chan<- prometheus.Metric) {
	log.Info("Running scrape")

	if stats, err := getServiceBusStats(); err != nil {
		ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0)

		log.WithError(err).Error("Error during scrape")
	} else {
		ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0)

		collectQueueMetrics(e, ch, stats)
		collectTopicAndSubscriptionMetrics(e, ch, stats)

		log.Info("Scrape completed")
	}

}

func collectQueueMetrics(e *exporter, ch chan<- prometheus.Metric, stats *stats) {
	for _, queue := range *stats.queues {
		collectMessageCounts(ch, &e.queueMessageCounts, &queue.messageCounts, queue.name)
		collectSizes(ch, &e.queueSizes, &queue.sizes, queue.name)
	}
}

func collectTopicAndSubscriptionMetrics(e *exporter, ch chan<- prometheus.Metric, stats *stats) {
	for _, topic := range *stats.topics {
		collectMessageCounts(ch, &e.topicMessageCounts, &topic.messageCounts, topic.name)
		collectSizes(ch, &e.topicSizes, &topic.sizes, topic.name)

		for _, sub := range *topic.subscriptions {
			collectMessageCounts(ch, &e.subscriptionMessageCounts, &sub.messageCounts, sub.name, topic.name)
		}
	}
}

func collectMessageCounts(ch chan<- prometheus.Metric, metrics *messageCountMetrics, counts *messageCounts, labelValues ...string) {
	ch <- prometheus.MustNewConstMetric(metrics.activeMessages, prometheus.GaugeValue, float64(counts.activeMessages), labelValues...)
	ch <- prometheus.MustNewConstMetric(metrics.deadLetterMessages, prometheus.GaugeValue, float64(counts.deadLetterMessages), labelValues...)
	ch <- prometheus.MustNewConstMetric(metrics.scheduledMessages, prometheus.GaugeValue, float64(counts.scheduledMessages), labelValues...)
	ch <- prometheus.MustNewConstMetric(metrics.transferMessages, prometheus.GaugeValue, float64(counts.transferMessages), labelValues...)
	ch <- prometheus.MustNewConstMetric(metrics.transferDeadLetterMessages, prometheus.GaugeValue, float64(counts.transferDeadLetterMessages), labelValues...)
}

func collectSizes(ch chan<- prometheus.Metric, metrics *sizeMetrics, sizes *sizes, labelValues ...string) {
	ch <- prometheus.MustNewConstMetric(metrics.size, prometheus.GaugeValue, float64(sizes.sizeInBytes), labelValues...)
	ch <- prometheus.MustNewConstMetric(metrics.maxSize, prometheus.GaugeValue, float64(sizes.maxSizeInBytes), labelValues...)
}
