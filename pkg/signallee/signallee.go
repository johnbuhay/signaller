package signallee

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
)

// signalPrinter prints signals received to stdout
func signalPrinter(ctx context.Context, signalChan chan os.Signal) {
	defer close(signalChan)
	signal.Notify(signalChan)

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case s := <-signalChan:
			if s != nil {
				// s != syscall.SIGINT && s != syscall.SIGTERM
				log.Println("Caught signal:", s)
			}
		}
	}
}

// Run manages the pidfile and prints signals notified to stdout
func Run(ctx context.Context, config interface{}) error {
	pidfile := config.(map[string]interface{})["pidfile"].(string)

	// listen for signals
	signalChan := make(chan os.Signal)
	go signalPrinter(ctx, signalChan)

	// write pid to pidfile
	if err := WritePID(pidfile); err != nil {
		return nil
	}

	// remove pidfile upon exiting
	<-ctx.Done()
	if err := RemovePID(pidfile); err != nil {
		return err
	}

	log.Println("Finished Run")
	return nil
}

// RemovePID deletes the pidfile
func RemovePID(path string) error {
	if err := os.Remove(path); err != nil {
		return err
	}
	return nil
}

// WritePID gets the current process id and writes it to a file
func WritePID(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	_, err = f.WriteString(strconv.Itoa(os.Getpid()))
	if err != nil {
		f.Close()
		return err
	}
	fmt.Println("pidfile written successfully")
	err = f.Close()
	if err != nil {
		return err
	}

	return nil
}
