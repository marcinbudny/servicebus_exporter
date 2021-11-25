package collector

import (
	"fmt"
	"strings"

	sb "github.com/marcinbudny/servicebus_exporter/client"
	klog "k8s.io/klog/v2"

	"github.com/prometheus/client_golang/prometheus"
)

type messageCountMetrics struct {
	activeMessages             *prometheus.Desc
	deadLetterMessages         *prometheus.Desc
	scheduledMessages          *prometheus.Desc
	transferDeadLetterMessages *prometheus.Desc
	transferMessages           *prometheus.Desc
	totalMessageCount          *prometheus.Desc
}

type sizeMetrics struct {
	size    *prometheus.Desc
	maxSize *prometheus.Desc
}

type Collector struct {
	client *sb.ServiceBusClient

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
		totalMessageCount: prometheus.NewDesc(fmt.Sprintf("servicebus_%s_transfer_total_messages", strings.ToLower(itemName)),
			fmt.Sprintf("%s total messages count", itemName), labels, nil),
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

func New(client *sb.ServiceBusClient) *Collector {
	return &Collector{
		client: client,

		up: prometheus.NewDesc("servicebus_up", "Whether the Azure ServiceBus scrape was successful", nil, nil),

		queueMessageCounts: newMessageCountMetrics("Queue", "name"),
		queueSizes:         newSizeMetrics("Queue"),

		topicMessageCounts: newMessageCountMetrics("Topic", "name"),
		topicSizes:         newSizeMetrics("Topic"),

		subscriptionMessageCounts: newMessageCountMetrics("Subscription", "name", "topic_name"),
	}
}

func describeMessageCountMetrics(ch chan<- *prometheus.Desc, metrics *messageCountMetrics) {
	ch <- metrics.activeMessages
	ch <- metrics.deadLetterMessages
	ch <- metrics.scheduledMessages
	ch <- metrics.transferMessages
	ch <- metrics.transferDeadLetterMessages
	ch <- metrics.totalMessageCount
}

func describeSizeMetrics(ch chan<- *prometheus.Desc, metrics *sizeMetrics) {
	ch <- metrics.size
	ch <- metrics.maxSize
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up

	describeMessageCountMetrics(ch, &c.queueMessageCounts)
	describeSizeMetrics(ch, &c.queueSizes)

	describeMessageCountMetrics(ch, &c.topicMessageCounts)
	describeSizeMetrics(ch, &c.topicSizes)

	describeMessageCountMetrics(ch, &c.subscriptionMessageCounts)
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	klog.Info("running scrape")

	if stats, err := c.client.GetServiceBusStats(); err != nil {
		ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 0)

		klog.ErrorS(err, "error during scrape")
	} else {
		ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 1)

		collectQueueMetrics(c, ch, stats)
		collectTopicAndSubscriptionMetrics(c, ch, stats)

		klog.Info("scrape completed")
	}

}

func collectQueueMetrics(c *Collector, ch chan<- prometheus.Metric, stats *sb.Stats) {
	for _, queue := range *stats.Queues {
		collectMessageCounts(ch, &c.queueMessageCounts, &queue.MessageCounts, queue.Name)
		collectSizes(ch, &c.queueSizes, &queue.Sizes, queue.Name)
	}
}

func collectTopicAndSubscriptionMetrics(c *Collector, ch chan<- prometheus.Metric, stats *sb.Stats) {
	for _, topic := range *stats.Topics {
		collectMessageCounts(ch, &c.topicMessageCounts, &topic.MessageCounts, topic.Name)
		collectSizes(ch, &c.topicSizes, &topic.Sizes, topic.Name)

		for _, sub := range *topic.Subscriptions {
			collectMessageCounts(ch, &c.subscriptionMessageCounts, &sub.MessageCounts, sub.Name, topic.Name)
		}
	}
}

func collectMessageCounts(ch chan<- prometheus.Metric, metrics *messageCountMetrics, counts *sb.MessageCounts, labelValues ...string) {
	ch <- prometheus.MustNewConstMetric(metrics.activeMessages, prometheus.GaugeValue, float64(counts.ActiveMessages), labelValues...)
	ch <- prometheus.MustNewConstMetric(metrics.deadLetterMessages, prometheus.GaugeValue, float64(counts.DeadLetterMessages), labelValues...)
	ch <- prometheus.MustNewConstMetric(metrics.scheduledMessages, prometheus.GaugeValue, float64(counts.ScheduledMessages), labelValues...)
	ch <- prometheus.MustNewConstMetric(metrics.transferMessages, prometheus.GaugeValue, float64(counts.TransferMessages), labelValues...)
	ch <- prometheus.MustNewConstMetric(metrics.transferDeadLetterMessages, prometheus.GaugeValue, float64(counts.TransferDeadLetterMessages), labelValues...)
}

func collectSizes(ch chan<- prometheus.Metric, metrics *sizeMetrics, sizes *sb.Sizes, labelValues ...string) {
	ch <- prometheus.MustNewConstMetric(metrics.size, prometheus.GaugeValue, float64(sizes.SizeInBytes), labelValues...)
	ch <- prometheus.MustNewConstMetric(metrics.maxSize, prometheus.GaugeValue, float64(sizes.MaxSizeInBytes), labelValues...)
}
