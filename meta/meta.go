package meta

import (
	pb "github.com/proximax-storage/go-bitswap/meta/pb"
)

func New() *plainMeta {
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

func (m *plainMeta) Set(unit Unit) {
	(*m)[unit.Key()] = unit.Value()
}

func (m plainMeta) Get(key string) Unit {
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

func (m plainMeta) All() []Unit {
	if len(m) == 0 {
		return nil
	}

	units := make([]Unit, 0, len(m))

	for key, val := range m {
		units = append(units, &PlainUnit{K: key, V: val})
	}

	return units
}

type ToProtoConverterFn func(meta Interface) (*pb.Meta, error)

func (ref ToProtoConverterFn) ToProto(meta Interface) (*pb.Meta, error) {
	return ref(meta)
}

func NewToProtoConverter() ToProtoConverterFn {
	return ToProtoConverterFn(toProtoMeta)
}

type FromProtoConverterFn func(metaProto *pb.Meta) (Interface, error)

func (ref FromProtoConverterFn) FromProto(metaProto *pb.Meta) (Interface, error) {
	return ref(metaProto)
}

func NewFromProtoConverter() FromProtoConverterFn {
	return FromProtoConverterFn(fromProtoMeta)
}

func toProtoMeta(meta Interface) (*pb.Meta, error) {
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

func fromProtoMeta(metaProto *pb.Meta) (Interface, error) {
	if metaProto == nil {
		return nil, ErrNilMeta
	}

	meta := New()

	for _, unit := range metaProto.Units {
		meta.Set(&PlainUnit{
			K: unit.Key,
			V: unit.Value,
		})
	}

	return meta, nil
}
