package message

import (
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	pb "github.com/proximax-storage/go-bitswap/message/pb"
	"github.com/proximax-storage/go-bitswap/meta"
)

type Interface interface {
	// WantList returns a slice of unique keys that represent data wanted by
	// the sender.
	WantList() []cid.Cid

	// Blocks returns a slice of unique blocks.
	Blocks() []blocks.Block

	// AddWantedEntry adds an entry to the WantList.
	AddWantedEntry(cid.Cid)

	AddBlock(blocks.Block)
}

type MetaExtended interface {
	Interface
	meta.Interface
}

type ToProtoConverter interface {
	ToProto(message MetaExtended) (*pb.Message, error)
}

type FromProtoConverter interface {
	FromProto(messageProto *pb.Message) (MetaExtended, error)
}

type Writer interface {
	WriteMessage(message MetaExtended) error
}

type Reader interface {
	ReadMessage() (MetaExtended, error)
}
