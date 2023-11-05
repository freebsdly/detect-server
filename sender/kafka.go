package sender

import (
	"context"
	"detect-server/connector"
	"detect-server/log"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/spf13/viper"
)

type KafkaSenderOptions struct {
	Brokers     []string       `yaml:"brokers"`
	Topic       string         `yaml:"topic"`
	MessageKey  string         `yaml:"message_key"`
	KafkaConfig *sarama.Config `yaml:"kafka_config"`
	Count       int            `yaml:"count"`
}

func NewKafkaSenderOptions() KafkaSenderOptions {
	var options = KafkaSenderOptions{
		Brokers:     viper.GetStringSlice("sender.kafka.brokers"),
		Topic:       viper.GetString("sender.kafka.topic"),
		MessageKey:  viper.GetString("sender.kafka.messageKey"),
		KafkaConfig: sarama.NewConfig(),
		Count:       viper.GetInt("sender.kafka.count"),
	}

	if options.Count <= 0 {
		options.Count = 3
	}

	if options.Topic == "" {
		options.Topic = "detect"
	}

	if options.MessageKey == "" {
		options.MessageKey = "detect"
	}

	return options
}

type KafkaSender struct {
	options    KafkaSenderOptions
	ctx        context.Context
	cancelFunc context.CancelFunc
	receiver   connector.Receiver[any]
}

func NewKafkaSender(options KafkaSenderOptions) Sender[any] {
	var sender = &KafkaSender{
		options: options,
	}

	return sender
}

func (kafka *KafkaSender) AddReceiver(receiver connector.Receiver[any]) {
	kafka.receiver = receiver
}

func (kafka *KafkaSender) Start() error {
	if kafka.receiver == nil {
		return fmt.Errorf("kafka sender message queue is invaild")
	}
	kafka.ctx, kafka.cancelFunc = context.WithCancel(context.Background())
	for i := 1; i <= kafka.options.Count; i++ {
		var name = fmt.Sprintf("sender%d", i)
		var err = kafka.startKafkaClient(name)
		if err != nil {
			log.Logger.Errorf("%s", err)
		}
	}
	return nil
}

func (kafka *KafkaSender) startKafkaClient(name string) error {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = kafka.options.KafkaConfig.Producer.Retry.Max
	config.Producer.Return.Successes = kafka.options.KafkaConfig.Producer.Return.Successes
	config.ClientID = kafka.options.KafkaConfig.ClientID
	config.Producer.Flush.Frequency = kafka.options.KafkaConfig.Producer.Flush.Frequency
	config.Producer.Timeout = kafka.options.KafkaConfig.Producer.Timeout
	var kafkaClient, err = sarama.NewAsyncProducer(kafka.options.Brokers, config)
	if err != nil {
		return fmt.Errorf("create kafka productor failed. %s\n", err)
	}

	go func(ctx context.Context, client sarama.AsyncProducer) {
		log.Logger.Infof("start kafka sender %s", name)
		for {
			select {
			case <-ctx.Done():
				log.Logger.Infof("close kafka client")
				_ = client.Close()
			case msg := <-kafka.receiver.Receive():
				data, err := json.Marshal(msg)
				if err != nil {
					log.Logger.Debugf("send message to kafka failed. %s", err)
					continue
				}
				client.Input() <- &sarama.ProducerMessage{
					Topic: kafka.options.Topic,
					Key:   sarama.StringEncoder(kafka.options.MessageKey),
					Value: sarama.StringEncoder(data),
				}
			}
		}
	}(kafka.ctx, kafkaClient)
	return nil
}

func (kafka *KafkaSender) Stop() error {
	if kafka.cancelFunc == nil {
		return fmt.Errorf("kafka sender already closed")
	}
	kafka.cancelFunc()
	return nil
}
