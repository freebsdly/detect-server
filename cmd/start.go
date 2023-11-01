/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"detect-server/api"
	"detect-server/detector"
	dispatcher "detect-server/dispatcher"
	"detect-server/log"
	"detect-server/receiver"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	var dispatcherOptions = dispatcher.Options{
		MaxIcmpResultQueueSize: viper.GetInt("dispatcher.icmp.result.queue.maxSize"),
		IcmpReceiverOptions: receiver.Options{
			MaxBufferSize: viper.GetInt("receiver.icmp.buffer.maxSize"),
		},
	}
	var icmpDetectorOptions = detector.IcmpDetectorOptions{
		DefaultTimeout:     viper.GetInt("detector.icmp.detect.timeout"),
		DefaultCount:       viper.GetInt("detector.icmp.detect.count"),
		MaxRunnerCount:     viper.GetInt("detector.icmp.runner.count"),
		MaxTaskBufferSize:  viper.GetInt("detector.icmp.task.bufferSize"),
		MaxResultQueueSize: viper.GetInt("detector.icmp.task.resultQueueSize"),
	}
	var dis = dispatcher.NewDispatcher(dispatcherOptions)
	go func() {
		if err := dis.Start(); err != nil {
			log.Logger.Errorf("start dispatcher failed. %s", err)
			os.Exit(1)
		}
	}()

	var httpApiOptions = api.HttpApiOptions{
		Listen:           "0.0.0.0:8080",
		MaxDetectTargets: 10000,
	}
	var httpApi = api.NewHttpApi(httpApiOptions)
	httpApi.SetIcmpReceiver(dis.GetIcmpReceiver())
	if err := httpApi.Start(); err != nil {
		log.Logger.Errorf("start http api failed. %s", err)
	}
}
