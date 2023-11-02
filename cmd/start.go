package cmd

import (
	"detect-server/api"
	"detect-server/connector"
	"detect-server/detector"
	dispatcher "detect-server/dispatcher"
	"detect-server/log"
	"detect-server/sender"
	"github.com/IBM/sarama"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "gopkg.in/yaml.v3"
	"os"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start detect server",
	Long:  `start detect server daemon`,
	Run: func(cmd *cobra.Command, args []string) {
		log.InitLogger()
		startDetectServer()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func startDetectServer() {
	var (
		icmpConnectorOptions = connector.Options{
			MaxBufferSize: viper.GetInt("connector.icmp.buffer.maxSize"),
		}
		icmpDetectorOptions = detector.IcmpDetectorOptions{
			DefaultTimeout:     viper.GetInt("detector.icmp.detect.timeout"),
			DefaultCount:       viper.GetInt("detector.icmp.detect.count"),
			MaxRunnerCount:     viper.GetInt("detector.icmp.runner.count"),
			MaxTaskBufferSize:  viper.GetInt("detector.icmp.task.bufferSize"),
			MaxResultQueueSize: viper.GetInt("detector.icmp.task.resultQueueSize"),
		}
		dispatcherOptions = dispatcher.Options{
			MaxIcmpResultQueueSize: viper.GetInt("dispatcher.icmp.result.queue.maxSize"),
		}
		httpApiOptions = api.HttpApiOptions{
			Listen:           viper.GetString("api.http.listen"),
			MaxDetectTargets: viper.GetInt("api.http.maxReceiveSize"),
		}
		kafkaSenderOptions = sender.KafkaSenderOptions{
			Brokers:     viper.GetStringSlice("sender.kafka.brokers"),
			Topic:       viper.GetString("sender.kafka.topic"),
			MessageKey:  viper.GetString("sender.kafka.messageKey"),
			KafkaConfig: sarama.NewConfig(),
		}
	)

	var (
		icmpConnector = connector.NewConnector[detector.Task[detector.IcmpDetect]](icmpConnectorOptions)
		dispatch      = dispatcher.NewDispatcher(dispatcherOptions)
		icmpDetector  = detector.NewIcmpDetector(icmpDetectorOptions)
		httpApi       = api.NewHttpApi(httpApiOptions)
		kafkaSender   = sender.NewKafkaSender(kafkaSenderOptions)
	)

	var err = icmpDetector.Start()
	if err != nil {
		log.Logger.Errorf("start icmp detector failed. %s", err)
	}

	dispatch.AddIcmpReceiver(icmpConnector)
	dispatch.AddIcmpDetector(icmpDetector)
	dispatch.AddSender(kafkaSender)

	if err := dispatch.Start(); err != nil {
		log.Logger.Errorf("start dispatcher failed. %s", err)
		os.Exit(1)
	}

	httpApi.AddIcmpPublisher(icmpConnector)
	if err := httpApi.Start(); err != nil {
		log.Logger.Errorf("start http api failed. %s", err)
	}
}
