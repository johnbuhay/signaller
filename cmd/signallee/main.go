package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/johnbuhay/signaller/pkg/signallee"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	rootCmd *cobra.Command
	// these are overriden by ldflags at build time
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func init() {
	rootCmd = &cobra.Command{
		Use:   "signallee",
		Short: "Display signals recieved",
		Long:  `Used to test signaller`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancelFunc := context.WithCancel(context.Background())

			signalChan := make(chan os.Signal, 1)
			go signalHandler(ctx, cancelFunc, signalChan)

			c := make(map[string]interface{})
			err := viper.Unmarshal(&c)
			if err != nil {
				fmt.Fprintln(os.Stderr, "unable to decode into struct,", err)
				os.Exit(1)
			}

			if err = signallee.Run(ctx, c); err != nil {
				return err
			}

			cancelFunc()
			return nil
		},
	}

	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $PWD/.signaller.yaml)")
	viper.BindEnv("PIDFILE")
	rootCmd.Version = version
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		pwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, "unable to get working directory,", err)
			os.Exit(1)
		}
		// Search config in current working directory with name ".signallee" (without extension).
		viper.AddConfigPath(pwd)
		viper.SetConfigName(".signallee")
	}

	viper.SetEnvPrefix("SIGNALLEE")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		log.Println("Loaded config from", viper.ConfigFileUsed())
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func signalHandler(ctx context.Context, cancelFunc context.CancelFunc, signalChan chan os.Signal) {
	defer close(signalChan)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		break
	case s := <-signalChan:
		if s != nil {
			log.Println("Caught signal:", s)
		}
	}

	cancelFunc()
}
