package meta

import (
	"github.com/pkg/errors"
	pb "github.com/proximax-storage/go-bitswap/meta/pb"
)

var ErrNilMeta = errors.New("nil meta")

type Unit interface {
	Key() string
	Value() []byte
}

type Interface interface {
	Set(Unit)
	Get(key string) Unit
	Delete(key string) bool
	All() []Unit
}

type ToProtoConverter interface {
	ToProto(meta Interface) (*pb.Meta, error)
}

type FromProtoConverter interface {
	FromProto(metaProto *pb.Meta) (Interface, error)
}
