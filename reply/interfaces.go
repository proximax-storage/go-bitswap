package reply

import (
	"github.com/proximax-storage/go-bitswap/meta"
	pb "github.com/proximax-storage/go-bitswap/reply/pb"
)

type Interface interface {
	meta.Interface
}

type ToProtoConverter interface {
	ToProto(reply Interface) (*pb.Reply, error)
}

type FromProtoConverter interface {
	FromProto(reply *pb.Reply) (Interface, error)
}

type Writer interface {
	WriteReply(reply Interface) error
}

type Reader interface {
	ReadReply() (Interface, error)
}
