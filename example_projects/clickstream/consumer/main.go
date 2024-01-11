package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/segmentio/kafka-go/sasl/scram"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	batchBytes     = 1024 * 1024
	defaultTimeout = 30 * time.Second
)

var (
	broker   = flag.String("broker", "", "")
	user     = flag.String("user", "default", "")
	password = flag.String("password", "default", "")
	topic    = flag.String("topic", "demo-events", "")
	groupID  = flag.String("group-id", "cli", "")
)

func main() {
	flag.Parse()
	sasl, err := scram.Mechanism(scram.SHA512, *user, *password)
	if err != nil {
		panic(err)
	}

	readerConfig := kafka.ReaderConfig{
		Brokers:  []string{*broker},
		Topic:    *topic,
		MaxBytes: batchBytes,
		Dialer: &kafka.Dialer{
			Timeout:       defaultTimeout,
			DualStack:     true,
			TLS:           new(tls.Config),
			SASLMechanism: sasl,
		},
		GroupID: *groupID,
	}
	if err := readerConfig.Validate(); err != nil {
		panic(err)
	}

	reader := kafka.NewReader(readerConfig)
	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			panic(err)
		}
		fmt.Printf("message: \n%s\n", string(msg.Value))
		if err := reader.CommitMessages(context.Background(), msg); err != nil {
			panic(err)
		}
	}
}
