package meta

import (
	pb "github.com/proximax-storage/go-bitswap/meta/pb"
)

func NewBitSwapMeta() BitSwapMeta {
	m := make(plainMeta)
	return &m
}

type PlainUnit struct {
	K string
	V []byte
}

func (ref *PlainUnit) Key() string {
	return ref.K
}

func (ref *PlainUnit) Value() []byte {
	return ref.V
}

type plainMeta map[string][]byte

func (m *plainMeta) Set(unit BitSwapMetaUnit) {
	(*m)[unit.Key()] = unit.Value()
}

func (m plainMeta) Get(key string) BitSwapMetaUnit {
	val, ok := m[key]
	if !ok {
		return nil
	}

	return &PlainUnit{K: key, V: val}
}

func (m *plainMeta) Delete(key string) bool {
	if _, ok := (*m)[key]; ok {
		delete(*m, key)

		return true
	}

	return false
}

func (m plainMeta) All() []BitSwapMetaUnit {
	if len(m) == 0 {
		return nil
	}

	units := make([]BitSwapMetaUnit, 0, len(m))

	for key, val := range m {
		units = append(units, &PlainUnit{K: key, V: val})
	}

	return units
}

type ToProtoConverterFn func(meta BitSwapMeta) (*pb.Meta, error)

func (ref ToProtoConverterFn) ToProto(meta BitSwapMeta) (*pb.Meta, error) {
	return ref(meta)
}

func NewToProtoConverter() ToProtoConverter {
	return ToProtoConverterFn(toProtoMeta)
}

type FromProtoConverterFn func(metaProto *pb.Meta) (BitSwapMeta, error)

func (ref FromProtoConverterFn) FromProto(metaProto *pb.Meta) (BitSwapMeta, error) {
	return ref(metaProto)
}

func NewFromProtoConverter() FromProtoConverter {
	return FromProtoConverterFn(fromProtoMeta)
}

func toProtoMeta(meta BitSwapMeta) (*pb.Meta, error) {
	if meta == nil {
		return nil, ErrNilMeta
	}

	values := meta.All()

	protoMeta := &pb.Meta{
		Units: make([]*pb.MetaUnit, len(values)),
	}

	for idx, unit := range meta.All() {
		protoMeta.Units[idx] = &pb.MetaUnit{
			Key:   unit.Key(),
			Value: unit.Value(),
		}
	}

	return protoMeta, nil
}

func fromProtoMeta(metaProto *pb.Meta) (BitSwapMeta, error) {
	if metaProto == nil {
		return nil, ErrNilMeta
	}

	meta := NewBitSwapMeta()

	for _, unit := range metaProto.Units {
		meta.Set(&PlainUnit{
			K: unit.Key,
			V: unit.Value,
		})
	}

	return meta, nil
}
