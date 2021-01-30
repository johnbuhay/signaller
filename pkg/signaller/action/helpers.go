package action

import (
	"errors"
	"syscall"
	"time"

	"github.com/johnbuhay/signaller/pkg/signaller/detect/file"
	"golang.org/x/sys/unix"
)

func ValidatePIDFile(s string) (string, error) {
	retries := 3 // maybe in the future i will parameterize this
	for i := 0; i <= retries; i++ {
		timeout := time.Duration(i) * time.Second
		time.Sleep(timeout)
		if err := file.Exists(s); err != nil {
			if i == retries {
				return "", err
			}
		} else {
			break
		}
	}
	return s, nil
}

// ValidateSignal returns 0, error when invalid
// https://pkg.go.dev/syscall#pkg-constants
// https://pkg.go.dev/golang.org/x/sys/unix#SignalNum
func ValidateSignal(s string) (syscall.Signal, error) {
	i := unix.SignalNum(s)
	if i == 0 {
		return i, errors.New("Signal with such name is not found. The signal name should start with \"SIG\"")
	}
	return i, nil
}
