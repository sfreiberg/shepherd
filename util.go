package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
)

func FileHash(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hash := md5.New()
	_, err = io.Copy(hash, f)
	if err != nil {
		return "", err
	}
	h := fmt.Sprintf("%x", hash.Sum(nil))
	return h, nil
}
