package action

import (
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

type Action struct {
	signal  syscall.Signal
	pidfile string
}

func New(i interface{}) (*Action, error) {
	si, err := ValidateSignal(i.(map[string]interface{})["signal"].(string))
	if err != nil {
		return &Action{}, err
	}
	pf, err := ValidatePIDFile(i.(map[string]interface{})["pidfile"].(string))
	if err != nil {
		return &Action{}, err
	}

	return &Action{signal: si, pidfile: pf}, nil
}

func (a *Action) SendSignal() error {
	data, err := ioutil.ReadFile(a.pidfile)
	if err != nil {
		return err
	}
	// chomp newline just in case
	pid, err := strconv.Atoi(strings.TrimSuffix(string(data), "\n"))
	if err != nil {
		return err
	}

	syscall.Kill(pid, a.signal)
	log.Println("Sent signal:", unix.SignalName(a.signal))
	return nil
}
