package message

import (
	"fmt"
	"github.com/pkg/errors"
	"io"

	blocks "github.com/ipfs/go-block-format"
	pb "github.com/proximax-storage/go-bitswap/message/pb"

	ggio "github.com/gogo/protobuf/io"
	"github.com/ipfs/go-cid"
	inet "github.com/libp2p/go-libp2p-net"
	"github.com/proximax-storage/go-bitswap/meta"
)

var ErrNilMessage = errors.New("nil message")

type protoWriter struct {
	ggio.Writer
	toProtoConverter ToProtoConverter
}

func NewProtoWriter(writer io.Writer) *protoWriter {
	return &protoWriter{
		Writer:           ggio.NewDelimitedWriter(writer),
		toProtoConverter: NewToProtoConverter(),
	}
}

func (ref *protoWriter) WriteMessage(message MetaExtended) error {
	messageProto, err := ref.toProtoConverter.ToProto(message)
	if err != nil {
		return err
	}

	return ref.WriteMsg(messageProto)
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

func (ref *protoReader) ReadMessage() (MetaExtended, error) {
	messageProto := new(pb.Message)
	if err := ref.ReadMsg(messageProto); err != nil {
		return nil, err
	}

	return ref.fromProtoConverter.FromProto(messageProto)
}

type toProtoConverter struct {
	metaConverter meta.ToProtoConverter
}

func NewToProtoConverter() *toProtoConverter {
	return &toProtoConverter{
		metaConverter: meta.NewToProtoConverter(),
	}
}

func (ref *toProtoConverter) ToProto(message MetaExtended) (*pb.Message, error) {
	if message == nil {
		return nil, ErrNilMessage
	}

	metaProto, err := ref.metaConverter.ToProto(message)
	if err != nil {
		return nil, ErrNilMessage
	}

	var (
		wantList = message.WantList()
		blockArr = message.Blocks()
	)

	messageProto := &pb.Message{
		Meta:     metaProto,
		Wantlist: &pb.Message_Wantlist{},
	}

	messageProto.Wantlist.Entries = make([]*pb.Message_Wantlist_Entry, len(wantList))
	for idx, want := range wantList {
		messageProto.Wantlist.Entries[idx] = &pb.Message_Wantlist_Entry{
			Cid: want.Bytes(),
		}
	}

	messageProto.Payload = make([]*pb.Message_Block, len(blockArr))
	for idx, b := range blockArr {
		messageProto.Payload[idx] = &pb.Message_Block{
			Data:   b.RawData(),
			Prefix: b.Cid().Prefix().Bytes(),
		}
	}

	return messageProto, nil
}

type fromProtoConverter struct {
	metaConverter meta.FromProtoConverter
}

func NewFromProtoConverter() *fromProtoConverter {
	return &fromProtoConverter{
		metaConverter: meta.NewFromProtoConverter(),
	}
}

func (ref *fromProtoConverter) FromProto(messageProto *pb.Message) (MetaExtended, error) {
	if messageProto == nil {
		return nil, ErrNilMessage
	}

	m, err := ref.metaConverter.FromProto(messageProto.Meta)
	if err != nil {
		return nil, err
	}

	message := &message{
		Interface: m,
	}

	for _, entry := range messageProto.Wantlist.Entries {
		c, err := cid.Cast([]byte(entry.Cid))
		if err != nil {
			return nil, fmt.Errorf("incorrectly formatted cid in wantlist: %s", err)
		}

		message.AddWantedEntry(c)
	}

	for _, b := range messageProto.GetPayload() {
		pref, err := cid.PrefixFromBytes(b.GetPrefix())
		if err != nil {
			return nil, err
		}

		c, err := pref.Sum(b.GetData())
		if err != nil {
			return nil, err
		}

		blk, err := blocks.NewBlockWithCid(b.GetData(), c)
		if err != nil {
			return nil, err
		}

		message.AddBlock(blk)
	}

	return message, nil
}

type message struct {
	meta.Interface
	full     bool
	wantList map[cid.Cid]struct{}
	blocks   map[cid.Cid]blocks.Block
}

func New() *message {
	return &message{
		Interface: meta.New(),
		blocks:    make(map[cid.Cid]blocks.Block),
		wantList:  make(map[cid.Cid]struct{}),
	}
}
func (m *message) Empty() bool {
	return len(m.blocks) == 0 && len(m.wantList) == 0
}

func (m *message) WantList() []cid.Cid {
	out := make([]cid.Cid, 0, len(m.wantList))
	for id := range m.wantList {
		out = append(out, id)
	}
	return out
}

func (m *message) Blocks() []blocks.Block {
	bs := make([]blocks.Block, 0, len(m.blocks))
	for _, block := range m.blocks {
		bs = append(bs, block)
	}
	return bs
}

func (m *message) AddWantedEntry(k cid.Cid) {
	m.wantList[k] = struct{}{}
}

func (m *message) AddBlock(b blocks.Block) {
	m.blocks[b.Cid()] = b
}
