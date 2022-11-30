package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"net/http"
)

// Consumer is the type for receiving events from queue.
type Consumer struct {
	conn      *amqp.Connection
	queueName string
}

func NewCustomer(conn *amqp.Connection) (Consumer, error) {
	consumer := Consumer{
		conn: conn,
	}

	err := consumer.setup()

	if err != nil {
		return Consumer{}, err
	}

	return consumer, nil
}

func (consumer *Consumer) setup() error {
	channel, err := consumer.conn.Channel()
	if err != nil {
		return err
	}

	// return the result of declaring an exchange
	return declareExchange(channel)
}

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (consumer *Consumer) Listen(topics []string) error {
	ch, err := consumer.conn.Channel()
	if err != nil {
		return err
	}

	defer ch.Close()
	q, err := declareRandomQueue(ch)
	if err != nil {
		return err
	}

	for _, curr := range topics {
		err = ch.QueueBind(
			q.Name,
			curr,
			"logs_topic",
			false,
			nil,
		)

		if err != nil {
			return err
		}
	}
	messages, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	forever := make(chan bool)

	go func() {
		for d := range messages {
			var payload Payload
			_ = json.Unmarshal(d.Body, &payload)

			go handlePayload(payload)
		}
	}()

	fmt.Printf("Waiting for message [Exchange, Queue] [logs_topic, %s]\n", q.Name)
	<-forever

	return nil
}

func handlePayload(p Payload) {
	switch p.Name {
	case "log", "event":
		// log whatever we get
		err := logEvent(p)

		if err != nil {
			log.Println(err)
		}
	case "auth":
	//authenticate
	default:
		err := logEvent(p)

		if err != nil {
			log.Println(err)
		}
	}
}

func logEvent(p Payload) error {
	jsonData, _ := json.MarshalIndent(p, "", "\t")

	logServiceURL := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		return err
	}

	return nil
}
