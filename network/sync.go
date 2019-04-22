package network

import (
	"context"
	inet "github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/pkg/errors"
	"github.com/proximax-storage/go-bitswap/message"
	"github.com/proximax-storage/go-bitswap/reply"
	"time"
)

type syncSender struct {
	*asyncMessageSender
	replyReader         reply.Reader
	replyReceiver       ReplyReceiver
	readDeadlineDefiner DeadlineDefiner
}

func newSyncSender(
	sender *asyncMessageSender,
	replyReader reply.Reader,
	replyReceiver ReplyReceiver,
	readDeadlineDefiner DeadlineDefiner,
) *syncSender {
	return &syncSender{
		asyncMessageSender:  sender,
		replyReader:         replyReader,
		replyReceiver:       replyReceiver,
		readDeadlineDefiner: readDeadlineDefiner,
	}
}

func (ref *syncSender) Send(ctx context.Context, message message.BitSwapMetaMessage) error {
	var err error

	if err = ref.asyncMessageSender.Send(ctx, message); err != nil {
		return err
	}

	if err = ref.s.SetReadDeadline(ref.writeDeadlineDefiner.Define(ctx)); err != nil {
		return err
	}

	bitSwapReply, err := ref.replyReader.ReadReply()
	if err != nil {
		return err
	}

	ref.replyReceiver.Receive(ref.s.Conn().RemotePeer(), bitSwapReply)

	return nil
}

type syncMsgSenderFactory struct {
	replyReceiverFactory ReplyReceiverFactory
	replyReaderFactory   ReplyReaderFactory
	writeDeadlineDef     DeadlineDefiner
	readDeadlineDef      DeadlineDefiner
}

func NewSyncMsgSenderFactory(
	replyReceiverFactory ReplyReceiverFactory,
	replyReaderFactory ReplyReaderFactory,
	writeDeadlineDef DeadlineDefiner,
	readDeadlineDef DeadlineDefiner,
) *syncMsgSenderFactory {
	return &syncMsgSenderFactory{
		replyReceiverFactory: replyReceiverFactory,
		replyReaderFactory:   replyReaderFactory,
		writeDeadlineDef:     writeDeadlineDef,
		readDeadlineDef:      readDeadlineDef,
	}
}

func (ref *syncMsgSenderFactory) Create(stream inet.Stream) (MessageSender, error) {
	replyReceiver, err := ref.replyReceiverFactory.Create(stream)
	if err != nil {
		return nil, err
	}

	replyReader, err := ref.replyReaderFactory.Create(stream)
	if err != nil {
		return nil, err
	}

	return newSyncSender(
		&asyncMessageSender{
			s:                    stream,
			messageWriter:        message.NewProtoWriter(stream),
			writeDeadlineDefiner: ref.writeDeadlineDef,
		},
		replyReader,
		replyReceiver,
		ref.readDeadlineDef,
	), nil
}

type syncMessageReceiver struct {
	s inet.Stream

	replyWriter     reply.Writer
	messageReplier  MessageReplier
	deadlineDefiner DeadlineDefiner

	errHandler ErrorHandler
}

func (ref *syncMessageReceiver) Receive(ctx context.Context, peerId peer.ID, message message.BitSwapMetaMessage) {
	r := ref.messageReplier.Reply(ctx, peerId, message)

	if err := ref.s.SetWriteDeadline(ref.deadlineDefiner.Define(ctx)); err != nil {
		ref.errHandler.Handle(err)
		return
	}

	if err := ref.replyWriter.WriteReply(r); err != nil {
		ref.errHandler.Handle(err)
		return
	}

	if err := ref.s.SetWriteDeadline(time.Time{}); err != nil {
		ref.errHandler.Handle(err)
		return
	}
}

type syncMessageReceiverFactory struct {
	replyWriterFactory ReplyWriterFactory
	messageReplier     MessageReplier
	deadlineDefiner    DeadlineDefiner
}

func NewSyncMessageReceiverFactory(replyWriterFactory ReplyWriterFactory, messageReplier MessageReplier, writeDeadlineDefiner DeadlineDefiner) MessageReceiverFactory {
	return &syncMessageReceiverFactory{
		replyWriterFactory: replyWriterFactory,
		messageReplier:     messageReplier,
		deadlineDefiner:    writeDeadlineDefiner,
	}
}

func (ref *syncMessageReceiverFactory) Create(stream inet.Stream) (MessageReceiver, error) {
	if stream == nil {
		return nil, errors.New("stream should not be nil")
	}

	replyWriter, err := ref.replyWriterFactory.Create(stream)
	if err != nil {
		return nil, err
	}

	return &syncMessageReceiver{
		s:               stream,
		replyWriter:     replyWriter,
		messageReplier:  ref.messageReplier,
		deadlineDefiner: ref.deadlineDefiner,
	}, nil
}
