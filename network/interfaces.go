package network

import (
	"context"
	inet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	"github.com/proximax-storage/go-bitswap/message"
	"github.com/proximax-storage/go-bitswap/reply"
	"io"
	"time"
)

type BlockExchange interface {
	SendMessage(context.Context, peer.ID, message.MetaExtended) error

	HandleNewStream(inet.Stream)
}

type MessageSender interface {
	Send(context.Context, message.MetaExtended) error
}

type MessageReceiver interface {
	Receive(context.Context, peer.ID, message.MetaExtended)
}

type MessageReplier interface {
	Reply(context.Context, peer.ID, message.MetaExtended) reply.Interface
}

type ReplyReceiver interface {
	Receive(peer.ID, reply.Interface)
}

type ErrorHandler interface {
	Handle(err error)
}

type DeadlineDefiner interface {
	Define(ctx context.Context) time.Time
}

type MessageSenderFactory interface {
	Create(inet.Stream) (MessageSender, error)
}

type ReplyReceiverFactory interface {
	Create(inet.Stream) (ReplyReceiver, error)
}

type MessageReceiverFactory interface {
	Create(inet.Stream) (MessageReceiver, error)
}

type ReplyWriterFactory interface {
	Create(inet.Stream) (reply.Writer, error)
}

type MessageReaderFactory interface {
	Create(io.Reader) (message.Reader, error)
}

type MessageWriterFactory interface {
	Create(io.Writer) (message.Writer, error)
}

type ReplyReaderFactory interface {
	Create(io.Reader) (reply.Reader, error)
}

type MessageSenderFactoryFn func(stream inet.Stream) (MessageSender, error)

func (ref MessageSenderFactoryFn) Create(stream inet.Stream) (MessageSender, error) {
	return ref(stream)
}

type ReplyWriterFactoryFn func(stream inet.Stream) (reply.Writer, error)

func (ref ReplyWriterFactoryFn) Create(stream inet.Stream) (reply.Writer, error) {
	return ref(stream)
}

type MessageReceiverFactoryFn func(stream inet.Stream) (MessageReceiver, error)

func (ref MessageReceiverFactoryFn) Create(stream inet.Stream) (MessageReceiver, error) {
	return ref(stream)
}

type MessageReaderFactoryFn func(r io.Reader) (message.Reader, error)

func (ref MessageReaderFactoryFn) Create(r io.Reader) (message.Reader, error) {
	return ref(r)
}

type ReplyReaderFactoryFn func(r io.Reader) (reply.Reader, error)

func (ref ReplyReaderFactoryFn) Create(r io.Reader) (reply.Reader, error) {
	return ref(r)
}

type MessageWriterFactoryFn func(r io.Writer) (message.Writer, error)

func (ref MessageWriterFactoryFn) Create(r io.Writer) (message.Writer, error) {
	return ref(r)
}

type ErrorHandlerFn func(err error)

func (ref ErrorHandlerFn) Handle(err error) {
	ref(err)
}
