package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/scram"
)

const (
	batchBytes = 1024 * 1024
)

var (
	broker   = flag.String("broker", "", "")
	user     = flag.String("user", "default", "")
	password = flag.String("password", "default", "")
	topic    = flag.String("topic", "demo-events", "")
	messages = flag.Int("message-count", 0, "max message count")
)

func main() {
	flag.Parse()
	authMechanism, err := scram.Mechanism(scram.SHA512, *user, *password)
	if err != nil {
		panic(err)
	}
	if err := createTopic(*broker, *topic, authMechanism); err != nil {
		panic(err)
	}

	writer := &kafka.Writer{
		Addr:       kafka.TCP(*broker),
		Balancer:   &kafka.Hash{},
		BatchBytes: batchBytes,
		Transport: &kafka.Transport{
			TLS:  new(tls.Config), // default SSL connection
			SASL: authMechanism,
		},
	}
	i := 0
	for i < *messages || *messages == 0 {
		i++
		messageBuilder := bytes.Buffer{}
		packSize := rand.Intn(100)
		for j := 0; j < packSize; j++ {
			messageBuilder.WriteString(fmt.Sprintf(
				`{"user_ts": "%s", "id": %v, "message": "test_%v_%v"}`,
				time.Now().UTC().Add(time.Duration(rand.Intn(10000))*time.Nanosecond).Format(time.RFC3339),
				time.Now().UTC().Nanosecond(),
				i,
				j,
			))
			if j < packSize-1 {
				messageBuilder.WriteString("\n")
			}
		}
		if err := writer.WriteMessages(context.Background(), kafka.Message{
			Topic:     *topic,
			Partition: 0,
			Value:     messageBuilder.Bytes(),
			Time:      time.Now(),
		}); err != nil {
			panic(err)
		}
		fmt.Printf("message: %v written\n", i)
		time.Sleep(10 * time.Millisecond)
	}
}

func createTopic(broker string, topic string, authMechanism sasl.Mechanism) error {
	dialer := kafka.Dialer{
		TLS:           new(tls.Config),
		SASLMechanism: authMechanism,
	}
	brokerConn, err := dialer.DialContext(context.Background(), "tcp", broker)
	if err != nil {
		return err
	}
	defer brokerConn.Close()

	controller, err := brokerConn.Controller()
	if err != nil {
		return err
	}

	controllerConn, err := dialer.DialContext(
		context.Background(),
		"tcp",
		fmt.Sprintf("%s:%d", controller.Host, controller.Port),
	)
	if err != nil {
		return err
	}

	_, err = controllerConn.ReadPartitions(topic)
	if err != kafka.UnknownTopicOrPartition {
		fmt.Printf("topic: %s exist\n", topic)
		return nil
	}
	fmt.Printf("topic: %s not exist, will create \n", topic)
	if err := controllerConn.CreateTopics(kafka.TopicConfig{
		Topic:              topic,
		NumPartitions:      -1,
		ReplicationFactor:  -1,
		ReplicaAssignments: nil,
		ConfigEntries: []kafka.ConfigEntry{
			{ConfigName: "retention.ms", ConfigValue: "21600000"},
			{ConfigName: "retention.bytes", ConfigValue: "5368709120"},
		},
	}); err != nil {
		return err
	}
	fmt.Printf("topic: %s created \n", topic)
	return nil
}
