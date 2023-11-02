package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"log"
)

type KafkaSenderOptions struct {
	Brokers     []string       `yaml:"brokers"`
	Topic       string         `yaml:"topic"`
	MessageKey  string         `yaml:"message_key"`
	KafkaConfig *sarama.Config `yaml:"kafka_config"`
}

type KafkaSender struct {
	kafkaClient sarama.AsyncProducer
	options     KafkaSenderOptions
	ctx         context.Context
	cancelFunc  context.CancelFunc
}

func NewKafkaSender(options KafkaSenderOptions) *KafkaSender {
	var sender = &KafkaSender{
		options: options,
	}

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = options.KafkaConfig.Producer.Retry.Max
	config.Producer.Return.Successes = options.KafkaConfig.Producer.Return.Successes
	config.ClientID = options.KafkaConfig.ClientID
	config.Producer.Flush.Frequency = options.KafkaConfig.Producer.Flush.Frequency
	config.Producer.Timeout = options.KafkaConfig.Producer.Timeout

	var err error
	sender.kafkaClient, err = sarama.NewAsyncProducer(options.Brokers, config)
	if err != nil {
		log.Printf("create kafka productor failed. %s\n", err)
	}

	return sender
}

func (sender *KafkaSender) SendMessage(result any) error {
	msg, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("send ping message to kafka failed. %s", err)
	}
	sender.kafkaClient.Input() <- &sarama.ProducerMessage{
		Topic: sender.options.Topic,
		Key:   sarama.StringEncoder(sender.options.MessageKey),
		Value: sarama.StringEncoder(msg),
	}
	return nil
}
