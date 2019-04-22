package message

import (
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	pb "github.com/proximax-storage/go-bitswap/message/pb"
	"github.com/proximax-storage/go-bitswap/meta"
)

type BitSwapMessage interface {
	// WantList returns a slice of unique keys that represent data wanted by
	// the sender.
	WantList() []cid.Cid

	// Blocks returns a slice of unique blocks.
	Blocks() []blocks.Block

	// AddWantedEntry adds an entry to the WantList.
	AddWantedEntry(cid.Cid)

	Empty() bool

	AddBlock(blocks.Block)
}

type BitSwapMetaMessage interface {
	BitSwapMessage
	meta.BitSwapMeta
}

type ToProtoConverter interface {
	ToProto(message BitSwapMetaMessage) (*pb.Message, error)
}

type FromProtoConverter interface {
	FromProto(messageProto *pb.Message) (BitSwapMetaMessage, error)
}

type Writer interface {
	WriteMessage(message BitSwapMetaMessage) error
}

type Reader interface {
	ReadMessage() (BitSwapMetaMessage, error)
}
