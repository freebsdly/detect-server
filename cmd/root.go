package cmd

import (
	"fmt"
	"github.com/spf13/viper"
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "detect-server",
	Short: "detect server supply some protocol detect",
	Long:  `detect server supply some protocol detect`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(os.Stderr, cmd.UsageString())

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// global flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
// if --config or -c is not set, ping server will auto search file ping-server.yaml in config paths
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("etc")
		viper.AddConfigPath("/etc")
		viper.SetConfigName("detect-server.yaml")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "read configuration file failed. %s\n", err)
		os.Exit(1)
	}
}
