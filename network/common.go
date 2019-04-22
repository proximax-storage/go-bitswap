package network

import (
	"context"
	"github.com/libp2p/go-libp2p-protocol"
	"github.com/proximax-storage/go-bitswap/reply"
	"io"
	"time"

	bsmsg "github.com/proximax-storage/go-bitswap/message"

	host "github.com/libp2p/go-libp2p-host"
	inet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	routing "github.com/libp2p/go-libp2p-routing"
)

var writeMessageTimeout = 30 * time.Second
var readMessageTimeout = 30 * time.Second

type Options struct {
	writeDeadlineDefiner DeadlineDefiner
	readDeadlineDefiner  DeadlineDefiner

	msgReaderFactory MessageReaderFactory
	msgWriterFactory MessageWriterFactory

	replyReaderFactory ReplyReaderFactory
	replyWriterFactory ReplyWriterFactory

	errHandler ErrorHandler
}

type Option interface {
	apply(options *Options)
}

type OptionFn func(options *Options)

func (ref OptionFn) apply(options *Options) {
	ref(options)
}

func WriteTimeOut(timeout time.Duration) Option {
	return OptionFn(func(options *Options) {
		options.writeDeadlineDefiner = newDefiner(timeout)
	})
}

func ReadTimeOut(timeout time.Duration) Option {
	return OptionFn(func(options *Options) {
		options.readDeadlineDefiner = newDefiner(timeout)
	})
}

func WriteDeadlineDefiner(definer DeadlineDefiner) Option {
	return OptionFn(func(options *Options) {
		options.writeDeadlineDefiner = definer
	})
}

func ReadDeadlineDefiner(definer DeadlineDefiner) Option {
	return OptionFn(func(options *Options) {
		options.readDeadlineDefiner = definer
	})
}

func MsgReaderFactory(readerFactory MessageReaderFactory) Option {
	return OptionFn(func(options *Options) {
		options.msgReaderFactory = readerFactory
	})
}

func MsgWriterFactory(writerFactory MessageWriterFactory) Option {
	return OptionFn(func(options *Options) {
		options.msgWriterFactory = writerFactory
	})
}

func ErrHandler(handler ErrorHandler) Option {
	return OptionFn(func(options *Options) {
		options.errHandler = handler
	})
}

func NewAsync(
	h host.Host,
	msgReceiverFactory MessageReceiverFactory,
	protocols []protocol.ID,
	opts ...Option,
) BlockExchange {
	options := newOptions(opts...)

	network := newNetwork(h, msgReceiverFactory, protocols, options)
	network.msgSenderFactory = createAsyncMessageSenderFactory(options.msgWriterFactory, options.writeDeadlineDefiner)

	return network
}

func NewSync(
	h host.Host,
	messageReplier MessageReplier,
	replyReceiverFactory ReplyReceiverFactory,
	protocols []protocol.ID,
	opts ...Option,
) BlockExchange {
	options := newOptions(opts...)

	network := newNetwork(
		h,
		NewSyncMessageReceiverFactory(
			options.replyWriterFactory,
			messageReplier,
			options.writeDeadlineDefiner,
		),
		protocols,
		options,
	)

	network.msgSenderFactory = NewSyncMsgSenderFactory(
		replyReceiverFactory,
		options.replyReaderFactory,
		options.writeDeadlineDefiner,
		options.readDeadlineDefiner,
	)

	return network
}

func newOptions(opts ...Option) *Options {
	options := &Options{
		writeDeadlineDefiner: newDefiner(writeMessageTimeout),
		readDeadlineDefiner:  newDefiner(readMessageTimeout),
		msgWriterFactory: MessageWriterFactoryFn(func(w io.Writer) (bsmsg.Writer, error) {
			return bsmsg.NewProtoWriter(w), nil
		}),
		msgReaderFactory: MessageReaderFactoryFn(func(r io.Reader) (bsmsg.Reader, error) {
			return bsmsg.NewProtoReader(r), nil
		}),
		replyWriterFactory: ReplyWriterFactoryFn(func(stream inet.Stream) (reply.Writer, error) {
			return reply.NewProtoWriter(stream), nil
		}),
		replyReaderFactory: ReplyReaderFactoryFn(func(r io.Reader) (reply.Reader, error) {
			return reply.NewProtoReader(r), nil
		}),
		errHandler: ErrorHandlerFn(func(err error) {}),
	}

	for _, opt := range opts {
		opt.apply(options)
	}

	return options
}

func newNetwork(host host.Host, msgReceiverFactory MessageReceiverFactory, protocols []protocol.ID, options *Options) *impl {
	return &impl{
		host:               host,
		protocols:          protocols,
		msgReaderFactory:   options.msgReaderFactory,
		msgReceiverFactory: msgReceiverFactory,
	}
}

type impl struct {
	host    host.Host
	routing routing.ContentRouting

	msgSenderFactory   MessageSenderFactory
	msgReaderFactory   MessageReaderFactory
	msgReceiverFactory MessageReceiverFactory

	errHandler ErrorHandler

	protocols []protocol.ID
}

func (ref *impl) SendMessage(ctx context.Context, p peer.ID, outgoing bsmsg.BitSwapMetaMessage) error {
	s, err := ref.host.NewStream(ctx, p, ref.protocols...)
	if err != nil {
		return err
	}

	messageSender, err := ref.msgSenderFactory.Create(s)
	if err != nil {
		return err
	}

	if err = messageSender.Send(ctx, outgoing); err != nil {
		_ = s.Reset()
		return err
	}

	return inet.FullClose(s)
}

func (ref *impl) HandleNewStream(s inet.Stream) {
	defer func() {
		_ = s.Close()
	}()

	messageReader, err := ref.msgReaderFactory.Create(s)
	if err != nil {
		_ = s.Reset()
		return
	}

	messageReceiver, err := ref.msgReceiverFactory.Create(s)
	if err != nil {
		_ = s.Reset()
		return
	}

	for {
		message, err := messageReader.ReadMessage()
		if err != nil {
			if err != io.EOF {
				_ = s.Reset()
				ref.errHandler.Handle(err)
			}

			return
		}

		messageReceiver.Receive(context.Background(), s.Conn().RemotePeer(), message)
	}
}

type definer struct {
	duration time.Duration
}

func newDefiner(duration time.Duration) *definer {
	return &definer{
		duration: duration,
	}
}

func (ref *definer) Define(ctx context.Context) time.Time {
	deadline := time.Now().Add(ref.duration)
	if dl, ok := ctx.Deadline(); ok {
		deadline = dl
	}

	return deadline
}
