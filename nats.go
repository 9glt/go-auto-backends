package autobackends

import (
	nats "github.com/nats-io/nats.go"
)

func Nats(s string, t string) *NatsClient {
	conn, err := nats.Connect(s)
	if err != nil {
		panic(err)
	}
	return &NatsClient{conn, t}
}

type NatsClient struct {
	nats *nats.Conn
	t    string
}

func (pub *NatsClient) Publish(msg string) {
	pub.nats.Publish(pub.t, []byte(msg))
}

func (sub *NatsClient) Subscribe(fn func(string)) {
	sub.nats.Subscribe(sub.t, func(msg *nats.Msg) {
		fn(string(msg.Data))
	})
}

var _ = (PubSub)(&NatsClient{})
