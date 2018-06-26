package main

import (
	"errors"
	"fmt"
)

// PubSub represents a redis pubsub pattern
// A client subscribed to one or more channels should not issue commands, although it can subscribe and unsubscribe to and from other channels.
// The replies to subscription and unsubscription operations are sent in the form of messages,
// so that the client can just read a coherent stream of messages where the first element indicates the type of message.
//The commands that are allowed in the context of a subscribed client are SUBSCRIBE, PSUBSCRIBE, UNSUBSCRIBE, PUNSUBSCRIBE, PING and QUIT.
type PubSub struct {
	conn Conn
	pool *ConnPool
}

// Message is a wrapper for messages received from server
type Message struct {
	Channel string
	Kind    string
	Payload string
}

func (m *Message) String() string {
	return fmt.Sprintf("Channel: %s\nKind: %s\nPayload: %s\n", m.Channel, m.Kind, m.Payload)
}

// Publish a message to channel
func (ps *PubSub) Publish(channel, message string) error {
	return ps.conn.SendCommand("PUBLISH", channel, message)
}

// Subscribe a channel
func (ps *PubSub) Subscribe(channel ...string) error {
	cmdStr := append([]string{"SUBSCRIBE"}, channel...)
	return ps.conn.SendCommand(cmdStr...)
}

// Unsubscribe a channel
func (ps *PubSub) Unsubscribe(channel ...string) error {
	cmdStr := append([]string{"UNSUBSCRIBE"}, channel...)
	return ps.conn.SendCommand(cmdStr...)
}

// PSubscribe subscribes the client to the given patterns
func (ps *PubSub) PSubscribe(pattern ...string) error {
	cmdStr := append([]string{"PSUBSCRIBE"}, pattern...)
	return ps.conn.SendCommand(cmdStr...)
}

// PUnsubscribe unsubscribes the client from the given patterns, or from all of them if none is given.
func (ps *PubSub) PUnsubscribe(pattern ...string) error {
	cmdStr := append([]string{"PUNSUBSCRIBE"}, pattern...)
	return ps.conn.SendCommand(cmdStr...)
}

// Receive message from server
// A message is a Array reply with three elements.
// The first element is the kind of message:
// 1. subscribe: means that we successfully subscribed to the channel given as the second element in the reply.
// The third argument represents the number of channels we are currently subscribed to.
// 2. unsubscribe: means that we successfully unsubscribed from the channel given as second element in the reply.
// The third argument represents the number of channels we are currently subscribed to.
// When the last argument is zero, we are no longer subscribed to any channel, and the client can issue any kind of Redis command as we are outside the Pub/Sub state.
// 3. message: it is a message received as result of a PUBLISH command issued by another client.
// The second element is the name of the originating channel, and the third argument is the actual message payload.
func (ps *PubSub) Receive() (*Message, error) {
	reply, err := ps.conn.ReadResp()
	if err != nil {
		return nil, err
	}
	arrayReply := reply.arrayVal
	messageType := string(arrayReply[0].stringVal)
	switch messageType {
	case "subscribe", "unsubscribe":
		return &Message{
			Kind:    messageType,
			Channel: string(arrayReply[1].stringVal),
		}, nil
	case "message":
		return &Message{
			Kind:    messageType,
			Channel: string(arrayReply[1].stringVal),
			Payload: string(arrayReply[2].stringVal),
		}, nil
	default:
		return nil, errors.New("not supported message type")
	}
}

// Close the pubsub connection
func (ps *PubSub) Close() {
	ps.pool.ReleaseConn(ps.conn)
}
