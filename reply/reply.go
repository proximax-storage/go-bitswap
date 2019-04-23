package reply

import (
	ggio "github.com/gogo/protobuf/io"
	inet "github.com/libp2p/go-libp2p-net"
	"github.com/pkg/errors"
	"github.com/proximax-storage/go-bitswap/meta"
	pb "github.com/proximax-storage/go-bitswap/reply/pb"
	"io"
)

var ErrNilReply = errors.New("nil reply")

type protoWriter struct {
	ggio.Writer
	toProtoConverter ToProtoConverter
}

func NewProtoWriter(w io.Writer) *protoWriter {
	return &protoWriter{
		Writer:           ggio.NewDelimitedWriter(w),
		toProtoConverter: NewToProtoConverter(),
	}
}

func (ref *protoWriter) WriteReply(reply Interface) error {
	replyProto, err := ref.toProtoConverter.ToProto(reply)
	if err != nil {
		return err
	}

	return ref.WriteMsg(replyProto)
}

type protoReader struct {
	ggio.Reader
	fromProtoConverter FromProtoConverter
}

func NewProtoReader(r io.Reader) *protoReader {
	return &protoReader{
		Reader:             ggio.NewDelimitedReader(r, inet.MessageSizeMax),
		fromProtoConverter: NewFromProtoConverter(),
	}
}

func (ref *protoReader) ReadReply() (Interface, error) {
	replyProto := new(pb.Reply)

	if err := ref.ReadMsg(replyProto); err != nil {
		return nil, err
	}

	reply, err := ref.fromProtoConverter.FromProto(replyProto)
	if err != nil {
		return nil, err
	}

	return reply, nil
}

type toProtoConverter struct {
	metaConverter meta.ToProtoConverter
}

func NewToProtoConverter() *toProtoConverter {
	return &toProtoConverter{
		metaConverter: meta.NewToProtoConverter(),
	}
}

func (ref *toProtoConverter) ToProto(reply Interface) (*pb.Reply, error) {
	if reply == nil {
		return nil, ErrNilReply
	}

	metaProto, err := ref.metaConverter.ToProto(reply)
	if err != nil {
		return nil, err
	}

	return &pb.Reply{
		Meta: metaProto,
	}, nil
}

type fromProtoConverter struct {
	metaConverter meta.FromProtoConverter
}

func NewFromProtoConverter() *fromProtoConverter {
	return &fromProtoConverter{
		metaConverter: meta.NewFromProtoConverter(),
	}
}

func (ref *fromProtoConverter) FromProto(replyProto *pb.Reply) (Interface, error) {
	if replyProto == nil {
		return nil, ErrNilReply
	}

	m, err := ref.metaConverter.FromProto(replyProto.Meta)
	if err != nil {
		return nil, err
	}

	return &reply{
		Interface: m,
	}, nil
}

type reply struct {
	meta.Interface
}

func New() *reply {
	return &reply{
		Interface: meta.New(),
	}
}
