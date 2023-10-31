/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"detect-server/detector"
	"detect-server/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"time"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start detect server",
	Long:  `start detect server daemon`,
	Run: func(cmd *cobra.Command, args []string) {
		log.InitLogger()
		startDetector()
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

func startDetector() {
	var icmpDetectorOptions = detector.IcmpDetectorOptions{
		DefaultTimeout: viper.GetInt("icmpDetector.timeout"),
		DefaultCount:   viper.GetInt("icmpDetector.count"),
		MaxRunnerCount: viper.GetInt("icmpDetector.maxRunnerCount"),
		MaxTargetQueue: viper.GetInt("icmpDetector.maxTargetQueue"),
	}
	var icmpDetector = detector.NewIcmpDetector(icmpDetectorOptions)
	var err = icmpDetector.Start()
	if err != nil {
		log.Logger.Errorf("%s", err)
		return
	}
	var targets = []detector.DetectOptions{
		{
			Ip: "192.168.1.1",
		},
		{
			Ip: "www.baidu.com",
		},
		{
			Ip: "www.google.com",
		},
		{
			Ip: "192.168.254.1",
		},
	}
	icmpDetector.Detects(targets)
	icmpDetector.Stop()
	time.Sleep(time.Second * 5)
}
