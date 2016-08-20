package main

import (
	"errors"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type configStoreEnv struct{}

func (c configStoreEnv) Get(path string) (interface{}, error) {
	return os.Getenv(c.pathToEnv(path)), nil
}

func (c configStoreEnv) GetString(path string) (string, error) {
	v, _ := c.Get(path)
	if r, ok := v.(string); ok {
		return r, nil
	}
	return "", errors.New("Unable to convert to string")
}

func (c configStoreEnv) GetInt(path string) (int64, error) {
	v, _ := c.Get(path)
	if r, ok := v.(string); ok {
		return strconv.ParseInt(r, 10, 64)
	}
	return 0, errors.New("Unable to convert to int64")
}

func (c configStoreEnv) pathToEnv(path string) string {
	path = strings.ToUpper(path)
	path = regexp.MustCompile(`[^A-Z]`).ReplaceAllString(path, "_")
	path = regexp.MustCompile(`__+`).ReplaceAllString(path, "_")
	return path
}
