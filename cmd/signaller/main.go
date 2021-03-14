package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/johnbuhay/signaller/pkg/signaller"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	poll    int
	rootCmd *cobra.Command

	// these are set` by ldflags at build time
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func exitIfError(e error) {
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}

func init() {
	rootCmd = &cobra.Command{
		Use:   "signaller",
		Short: "Detect a change and act",
		Long:  ``,
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
			you, err := signaller.New(c)
			if err != nil {
				return err
			}

			if poll > 0 {
				if err := you.Poll(ctx, poll); err != nil {
					return err
				}
			} else {
				if err := you.Watch(ctx); err != nil {
					return err
				}
			}

			cancelFunc()
			return nil
		},
	}

	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.signaller.yaml)")
	rootCmd.PersistentFlags().IntVar(&poll, "poll", 0, "to enable polling specify an interval greater than 0, in seconds")
	rootCmd.Version = version

}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// Search config in home directory with name ".signaller" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".signaller")
	}

	viper.SetEnvPrefix("SIGNALLER")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		log.Println("Loaded config from", viper.ConfigFileUsed())
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		exitIfError(err)
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
