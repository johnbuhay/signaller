package signaller

import (
	"context"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/johnbuhay/signaller/pkg/signaller/action"
	"github.com/johnbuhay/signaller/pkg/signaller/detect"
)

type Config struct {
	files []File
}

type File struct {
	action *action.Action
	detect *detect.Detect
}

func New(i map[string]interface{}) (*Config, error) {
	files := []File{}
	listOfItems := i["files"]

	for _, item := range listOfItems.([]interface{}) {
		f, err := GetFile(item.(map[interface{}]interface{}))
		if err != nil {
			return &Config{}, err
		}
		files = append(files, *f)
	}

	return &Config{files: files}, nil
}

func GetFile(item interface{}) (*File, error) {
	actionType := map[string]interface{}{
		"signal":  item.(map[interface{}]interface{})["signal"].(string),
		"pidfile": item.(map[interface{}]interface{})["pidfile"].(string),
	}
	a, err := action.New(actionType)
	if err != nil {
		return &File{}, err
	}
	detectType := map[string]interface{}{
		"file": item.(map[interface{}]interface{})["path"].(string),
	}
	d, err := detect.New(detectType)
	if err != nil {
		return &File{}, err
	}

	return &File{
		action: a,
		detect: d,
	}, nil
}

func (c *Config) Poll(ctx context.Context, interval int) error {
	done := make(chan bool)
	for _, file := range c.files {
		go PollFile(ctx, file, interval)
	}

	<-done
	log.Println("Closing Poll")
	return nil
}

func PollFile(ctx context.Context, file File, interval int) {
	changed := make(chan bool)
	go file.detect.Poll(ctx, changed, interval) // producer

	action := func() error {
		if err := file.action.SendSignal(); err != nil {
			return err
		}
		return nil
	}
	if err := Repeat(ctx, action, changed); err != nil { // consumer
		log.Println(err)
	}

	log.Println("Closing Poll")
}

func (c *Config) Watch(ctx context.Context) error {
	done := make(chan bool)
	for _, file := range c.files {
		go WatchFile(ctx, file, done)
	}

	<-done
	log.Println("Closing Watch")
	return nil
}

func WatchFile(ctx context.Context, file File, channel chan bool) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println("failed to allocate new watcher")
		os.Exit(1)
	}
	changed := make(chan bool)
	go file.detect.Watch(ctx, watcher, changed) // producer

	action := func() error {
		if err := file.action.SendSignal(); err != nil {
			return err
		}

		return nil
	}
	if err = Repeat(ctx, action, changed); err != nil { // consumer
		log.Println(err)
	}

	log.Println("Closing Watch")
}

func Repeat(ctx context.Context, f func() error, b chan bool) error {
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
