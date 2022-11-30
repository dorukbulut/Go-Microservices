package event

import (
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

type Emitter struct {
	connection *amqp.Connection
}

func NewEventEmiterr(connection *amqp.Connection) (Emitter, error) {
	emitter := Emitter{
		connection: connection,
	}
	err := emitter.setup()
	if err != nil {
		return Emitter{}, err
	}
	return emitter, nil
}

func (e *Emitter) setup() error {
	ch, err := e.connection.Channel()
	if err != nil {
		return err
	}

	defer ch.Close()

	return declareExchange(ch)

}

func (e *Emitter) Push(event, severity string) error {
	channel, err := e.connection.Channel()

	if err != nil {
		return err
	}
	defer channel.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()

	err = channel.PublishWithContext(
		ctx,
		"logs_topic",
		severity,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(event),
		})
	if err != nil {
		return err
	}

	return nil
}
