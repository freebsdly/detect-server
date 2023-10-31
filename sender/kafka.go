package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/spf13/viper"
)

type KafkaSenderOptions struct {
	Brokers        []string `yaml:"brokers,omitempty"`
	Topic          string   `yaml:"topic,omitempty"`
	Partition      int      `yaml:"partition,omitempty"`
	ClientId       string   `yaml:"client_id"`
	Timeout        int      `yaml:"timeout"`
	FlushFrequency int      `yaml:"flush_frequency"`
	MessageKey     string   `yaml:"message_key"`
}

type KafkaSender struct {
	kafkaClient sarama.AsyncProducer
	options     KafkaSenderOptions
	ctx         context.Context
	cancelFunc  context.CancelFunc
}

func NewKafkaSender(options KafkaSenderOptions) *KafkaSender {
	viper.SetDefault("kafkaSender.brokers", []string{"0.0.0.0:9092"})
	viper.SetDefault("kafkaSender.producer.retry.max", 1)
	viper.SetDefault("kafkaSender.producer.return.successes", false)
	viper.SetDefault("kafkaSender.producer.flush.frequency", 500)
	viper.SetDefault("kafkaSender.clientId", "detect-server")
	viper.SetDefault("kafkaSender.producer.timeout", 3000)
	if len(options.Brokers) == 0 {
		options.Brokers = []string{
			"0.0.0.0:9092",
		}
	}
	if options.Timeout <= 0 {
		options.Timeout = 3
	}

	if options.FlushFrequency <= 0 {
		options.FlushFrequency = 500
	}

	var sender = &KafkaSender{
		options: options,
	}

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 1
	config.Producer.Return.Successes = false
	config.ClientID = options.ClientId
	config.Producer.Flush.Frequency = time.Duration(options.FlushFrequency) * time.Millisecond
	config.Producer.Timeout = time.Duration(options.Timeout) * time.Second

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
		Value: sarama.StringEncoder(string(msg)),
	}
	return nil
}
