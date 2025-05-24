package mq

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"gitlab.com/Hamed1984/stats/pkg/conf"
)

func createAMQPConn(config *conf.Configuration, cNotify chan *amqp.Error) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(config.GetMQConnString())
	if err != nil {
		return nil, nil, err
	}
	conn.NotifyClose(cNotify)
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, err
	}
	return conn, ch, nil
}

type AMQPConsumer interface {
	StartConsume(onReceive func(amqp.Delivery, <-chan struct{}) error)
}