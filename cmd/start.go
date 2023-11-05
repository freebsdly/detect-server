package cmd

import (
	"detect-server/api"
	"detect-server/connector"
	"detect-server/detector"
	dispatcher "detect-server/dispatcher"
	"detect-server/log"
	"detect-server/sender"
	"github.com/go-ping/ping"
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
			MaxBufferSize: viper.GetInt("connector.icmp.buffer.size"),
		}
		msgConnectorOptions = connector.Options{
			MaxBufferSize: viper.GetInt("sender.buffer.size"),
		}
		icmpDetectorOptions = detector.NewIcmpDetectorOptions()
		dispatcherOptions   = dispatcher.NewOptions()
		httpApiOptions      = api.NewHttpApiOptions()
		kafkaSenderOptions  = sender.NewKafkaSenderOptions()
	)

	var (
		icmpConnector = connector.NewChanConnector[dispatcher.Task[detector.IcmpOptions]](icmpConnectorOptions)
		msgConnector  = connector.NewChanConnector[any](msgConnectorOptions)
		dispatch      = dispatcher.NewDispatcher[detector.IcmpOptions, *ping.Statistics, dispatcher.DefaultMessage](dispatcherOptions)
		icmpDetector  = detector.NewIcmpDetector(icmpDetectorOptions)
		httpApi       = api.NewHttpApi(httpApiOptions)
		kafkaSender   = sender.NewKafkaSender(kafkaSenderOptions)
		processor     = dispatcher.NewDefaultProcessor[detector.IcmpOptions, *ping.Statistics, dispatcher.DefaultMessage]()
	)

	// start detector
	var err = icmpDetector.Start()
	if err != nil {
		log.Logger.Errorf("start icmp detector failed. %s", err)
		os.Exit(1)
	}

	// start sender
	kafkaSender.AddReceiver(msgConnector)
	if err = kafkaSender.Start(); err != nil {
		log.Logger.Errorf("start kafka sender failed. %s", err)
		os.Exit(1)
	}

	// start dispatcher
	dispatch.AddReceiver(icmpConnector)
	dispatch.AddDetector(icmpDetector)
	dispatch.AddPublisher(msgConnector)
	dispatch.AddProcessor(processor)

	if err := dispatch.Start(); err != nil {
		log.Logger.Errorf("start dispatcher failed. %s", err)
		os.Exit(1)
	}

	// start api
	httpApi.AddIcmpPublisher(icmpConnector)
	if err := httpApi.Start(); err != nil {
		log.Logger.Errorf("start http api failed. %s", err)
	}
}
