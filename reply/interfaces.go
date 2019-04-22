package reply

import (
	"github.com/proximax-storage/go-bitswap/meta"
	pb "github.com/proximax-storage/go-bitswap/reply/pb"
)

type BitSwapMetaReply interface {
	meta.BitSwapMeta
}

type ToProtoConverter interface {
	ToProto(reply BitSwapMetaReply) (*pb.Reply, error)
}

type FromProtoConverter interface {
	FromProto(reply *pb.Reply) (BitSwapMetaReply, error)
}

type Writer interface {
	WriteReply(reply BitSwapMetaReply) error
}

type Reader interface {
	ReadReply() (BitSwapMetaReply, error)
}
