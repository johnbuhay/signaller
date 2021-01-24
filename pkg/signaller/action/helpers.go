package action

import (
	"errors"
	"syscall"

	"github.com/johnbuhay/signaller/pkg/signaller/detect/file"
	"golang.org/x/sys/unix"
)

func ValidatePIDFile(s string) (string, error) {
	if err := file.Exists(s); err != nil {
		return "", err
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
