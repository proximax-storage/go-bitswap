package network

import (
	"context"
	inet "github.com/libp2p/go-libp2p-net"
	"github.com/proximax-storage/go-bitswap/message"
	"time"
)

type asyncMessageSender struct {
	s                    inet.Stream
	messageWriter        message.Writer
	writeDeadlineDefiner DeadlineDefiner
}

func newAsyncMessageSender(s inet.Stream, messageWriter message.Writer, writeDeadlineDefiner DeadlineDefiner) *asyncMessageSender {
	return &asyncMessageSender{
		s:                    s,
		messageWriter:        messageWriter,
		writeDeadlineDefiner: writeDeadlineDefiner,
	}
}

func (ref *asyncMessageSender) Send(ctx context.Context, msg message.MetaExtended) error {
	var err error

	if err = ref.s.SetWriteDeadline(ref.writeDeadlineDefiner.Define(ctx)); err != nil {
		return err
	}

	if err = ref.messageWriter.WriteMessage(msg); err != nil {
		return err
	}

	if err = ref.s.SetWriteDeadline(time.Time{}); err != nil {
		return err
	}

	return nil
}

func createAsyncMessageSenderFactory(msgWriterFactory MessageWriterFactory, writeDeadlineDefiner DeadlineDefiner) MessageSenderFactory {
	return MessageSenderFactoryFn(func(stream inet.Stream) (MessageSender, error) {
		writer, err := msgWriterFactory.Create(stream)
		if err != nil {
			return nil, err
		}

		return newAsyncMessageSender(stream, writer, writeDeadlineDefiner), nil
	})
}
