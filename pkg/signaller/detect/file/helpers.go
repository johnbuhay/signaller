package file

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

// Exists returns an error if the path provided
// does not exist or is not a file
func Exists(p string) error {
	info, err := os.Stat(p)
	if os.IsNotExist(err) {
		return errors.New("file does not exist")
	}
	if info.IsDir() {
		return fmt.Errorf("%v is a directory", p)
	}
	return nil
}

func Checksum(p string) (string, error) {
	data, err := ioutil.ReadFile(p)
	if err != nil {
		return "", err
	}

	checksum := fmt.Sprintf("%x", sha1.Sum(data))
	return checksum, nil
}
