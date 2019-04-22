package meta

import (
	"github.com/pkg/errors"
	pb "github.com/proximax-storage/go-bitswap/meta/pb"
)

var ErrNilMeta = errors.New("nil meta")

type BitSwapMetaUnit interface {
	Key() string
	Value() []byte
}

type BitSwapMeta interface {
	Set(BitSwapMetaUnit)
	Get(key string) BitSwapMetaUnit
	Delete(key string) bool
	All() []BitSwapMetaUnit
}

type ToProtoConverter interface {
	ToProto(meta BitSwapMeta) (*pb.Meta, error)
}

type FromProtoConverter interface {
	FromProto(metaProto *pb.Meta) (BitSwapMeta, error)
}
