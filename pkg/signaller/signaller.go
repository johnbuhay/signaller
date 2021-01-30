package signaller

import (
	"context"
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/johnbuhay/signaller/pkg/signaller/action"
	"github.com/johnbuhay/signaller/pkg/signaller/detect"
)

type Config struct {
	action *action.Action
	detect *detect.Detect
}

func New(i interface{}) (*Config, error) {
	a, err := action.New(i.(map[string]interface{})["action"])
	if err != nil {
		return &Config{}, err
	}
	d, err := detect.New(i.(map[string]interface{})["detect"])
	if err != nil {
		return &Config{}, err
	}

	return &Config{
		action: a,
		detect: d,
	}, nil
}

func (c *Config) Poll(ctx context.Context, interval int) error {
	changed := make(chan bool)
	go c.detect.Poll(ctx, changed, interval) // producer

	action := func() error {
		if err := c.action.SendSignal(); err != nil {
			return err
		}
		return nil
	}
	if err := c.Repeat(ctx, action, changed); err != nil { // consumer
		return err
	}

	log.Println("Closing Poll")
	return nil
}

func (c *Config) Watch(ctx context.Context) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil
	}
	changed := make(chan bool)
	go c.detect.Watch(ctx, watcher, changed) // producer

	action := func() error {
		if err := c.action.SendSignal(); err != nil {
			return err
		}

		return nil
	}
	if err = c.Repeat(ctx, action, changed); err != nil { // consumer
		return err
	}

	log.Println("Closing Watch")
	return nil
}

func (c *Config) Repeat(ctx context.Context, f func() error, b chan bool) error {
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case repeat, ok := <-b:
			if !ok {
				break loop
			}

			if repeat {
				if err := f(); err != nil {
					return err
				}
			}
		}
	}
	log.Println("Closing Repeat")
	return nil
}
