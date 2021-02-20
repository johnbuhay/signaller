package detect

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/johnbuhay/signaller/pkg/signaller/detect/file"
)

type Detect struct {
	File *file.File
}

func New(i interface{}) (*Detect, error) {
	f, err := file.New(i.(map[string]interface{})["file"].(string))
	if err != nil {
		return &Detect{}, err
	}
	return &Detect{
		File: f,
	}, nil
}

func (d *Detect) Watch(ctx context.Context, watcher *fsnotify.Watcher, actionChan chan bool) {
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case <-ctx.Done():
				// maybe close these instead?
				actionChan <- false
				done <- true
				break
			case event, ok := <-watcher.Events:
				if !ok {
					done <- true
					return
				}

				if event.Op&fsnotify.Write == fsnotify.Write {
					c, _ := file.Checksum(d.File.Path())
					log.Println("event:", event)

					changed, err := d.File.CompareChecksum()
					if err != nil {
						log.Println(err)
						break
					}
					if changed {
						log.Println("modified file:", event.Name, c)
						actionChan <- true
						// remeasure checksum
						// be aware of "Why am I receiving multiple events for the same file on OS X?"
						// from https://github.com/fsnotify/fsnotify#faq
						d.File, err = file.New(d.File.Path())
						if err != nil {
							log.Println(err)
							break
						}
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					done <- true
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	log.Printf("Watching %v with checksum %v\n", d.File.Path(), d.File.Checksum())
	fileToWatch, err := filepath.EvalSymlinks(d.File.Path())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		done <- true
	} else {
		if err := watcher.Add(fileToWatch); err != nil {
			fmt.Fprintln(os.Stderr, err)
			done <- true
		}
	}

	<-done
	log.Println("The watch has ended...")
}

func (d *Detect) Poll(ctx context.Context, actionChan chan bool, interval int) {
	done := make(chan bool)
	go func() {
	loop:
		for {
			select {
			case <-ctx.Done():
				actionChan <- false
				done <- true
				break loop
			default:
				changed, err := d.File.CompareChecksum()
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					break loop
				}
				if changed {
					c, _ := file.Checksum(d.File.Path())
					log.Println("file modified:", d.File.Path(), c)
					actionChan <- true

					d.File, err = file.New(d.File.Path())
					if err != nil {
						fmt.Fprintln(os.Stderr, err)
						break loop
					}
				}
				time.Sleep(time.Duration(interval) * time.Second)
			}
		}
		done <- true
	}()

	log.Printf("Polling %v with checksum %v every %v seconds\n", d.File.Path(), d.File.Checksum(), interval)
	<-done
	log.Println("Polling has ended...")
}
