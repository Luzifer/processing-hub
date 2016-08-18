package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const (
	maxSQSMessageBodySize = 256 * 1024 // 256kB
)

type message map[string]string

func createMessage(t string) *message {
	return &message{
		"_type": t,
		"_date": time.Now().Format(time.RFC3339),
	}
}

func (m message) Validate() error {
	if d, ok := m["_type"]; ok {
		if len(strings.Split(d, ".")) < 3 {
			return errors.New("Message type must be in reverse uri format")
		}
	} else {
		return errors.New("Message contains invalid or missing type")
	}

	if d, ok := m["_date"]; ok {
		if _, err := time.Parse(time.RFC3339, d); err != nil {
			return errors.New("Message contains invalid or missing date")
		}
	} else {
		return errors.New("Message contains invalid or missing date")
	}

	if d, err := m.Encode(); err != nil || len(*d) > maxSQSMessageBodySize {
		return errors.New("Message encoding failed or message is bigger than max SQS body size")
	}

	return nil
}

func (m message) Type() string { return m["_type"] }

func (m message) Date() time.Time {
	d, _ := time.Parse(time.RFC3339, m["_date"])
	return d
}

func (m message) Set(key, value string) {
	m[key] = value
}

func (m message) Encode() (*string, error) {
	buf := bytes.NewBuffer([]byte{})
	gzipBuffer := gzip.NewWriter(buf)

	if err := json.NewEncoder(gzipBuffer).Encode(m); err != nil {
		return nil, err
	}

	gzipBuffer.Flush()

	msg := base64.StdEncoding.EncodeToString(buf.Bytes())
	return &msg, nil
}

func DecodeMessage(msg *string) (message, error) {
	raw, err := base64.StdEncoding.DecodeString(*msg)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(raw)
	gzipBuffer, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}

	result := message{}
	if err := json.NewDecoder(gzipBuffer).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}
