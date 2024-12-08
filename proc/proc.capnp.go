// Code generated by capnpc-go. DO NOT EDIT.

package proc

import (
	capnp "capnproto.org/go/capnp/v3"
	text "capnproto.org/go/capnp/v3/encoding/text"
	schemas "capnproto.org/go/capnp/v3/schemas"
)

type MethodCall capnp.Struct

// MethodCall_TypeID is the unique identifier for the type MethodCall.
const MethodCall_TypeID = 0xbf851cafa4aebfbb

func NewMethodCall(s *capnp.Segment) (MethodCall, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 3})
	return MethodCall(st), err
}

func NewRootMethodCall(s *capnp.Segment) (MethodCall, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 3})
	return MethodCall(st), err
}

func ReadRootMethodCall(msg *capnp.Message) (MethodCall, error) {
	root, err := msg.Root()
	return MethodCall(root.Struct()), err
}

func (s MethodCall) String() string {
	str, _ := text.Marshal(0xbf851cafa4aebfbb, capnp.Struct(s))
	return str
}

func (s MethodCall) EncodeAsPtr(seg *capnp.Segment) capnp.Ptr {
	return capnp.Struct(s).EncodeAsPtr(seg)
}

func (MethodCall) DecodeFromPtr(p capnp.Ptr) MethodCall {
	return MethodCall(capnp.Struct{}.DecodeFromPtr(p))
}

func (s MethodCall) ToPtr() capnp.Ptr {
	return capnp.Struct(s).ToPtr()
}
func (s MethodCall) IsValid() bool {
	return capnp.Struct(s).IsValid()
}

func (s MethodCall) Message() *capnp.Message {
	return capnp.Struct(s).Message()
}

func (s MethodCall) Segment() *capnp.Segment {
	return capnp.Struct(s).Segment()
}
func (s MethodCall) Name() (string, error) {
	p, err := capnp.Struct(s).Ptr(0)
	return p.Text(), err
}

func (s MethodCall) HasName() bool {
	return capnp.Struct(s).HasPtr(0)
}

func (s MethodCall) NameBytes() ([]byte, error) {
	p, err := capnp.Struct(s).Ptr(0)
	return p.TextBytes(), err
}

func (s MethodCall) SetName(v string) error {
	return capnp.Struct(s).SetText(0, v)
}

func (s MethodCall) Stack() (capnp.UInt64List, error) {
	p, err := capnp.Struct(s).Ptr(1)
	return capnp.UInt64List(p.List()), err
}

func (s MethodCall) HasStack() bool {
	return capnp.Struct(s).HasPtr(1)
}

func (s MethodCall) SetStack(v capnp.UInt64List) error {
	return capnp.Struct(s).SetPtr(1, v.ToPtr())
}

// NewStack sets the stack field to a newly
// allocated capnp.UInt64List, preferring placement in s's segment.
func (s MethodCall) NewStack(n int32) (capnp.UInt64List, error) {
	l, err := capnp.NewUInt64List(capnp.Struct(s).Segment(), n)
	if err != nil {
		return capnp.UInt64List{}, err
	}
	err = capnp.Struct(s).SetPtr(1, l.ToPtr())
	return l, err
}
func (s MethodCall) CallData() ([]byte, error) {
	p, err := capnp.Struct(s).Ptr(2)
	return []byte(p.Data()), err
}

func (s MethodCall) HasCallData() bool {
	return capnp.Struct(s).HasPtr(2)
}

func (s MethodCall) SetCallData(v []byte) error {
	return capnp.Struct(s).SetData(2, v)
}

// MethodCall_List is a list of MethodCall.
type MethodCall_List = capnp.StructList[MethodCall]

// NewMethodCall creates a new list of MethodCall.
func NewMethodCall_List(s *capnp.Segment, sz int32) (MethodCall_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 3}, sz)
	return capnp.StructList[MethodCall](l), err
}

// MethodCall_Future is a wrapper for a MethodCall promised by a client call.
type MethodCall_Future struct{ *capnp.Future }

func (f MethodCall_Future) Struct() (MethodCall, error) {
	p, err := f.Future.Ptr()
	return MethodCall(p.Struct()), err
}

const schema_c13229b64d08e68a = "x\xda\x14\xc8?J\xc4@\x18\x86\xf1\xf7\xfd&\xeb\x1f" +
	"L\xd4\x81\x94\xc2\xd6\x0a\x0ani\xb5\xe0Z(.\xec" +
	"\xd7+8\x8c\x81\x80c\x12t\xae`\xe39\xac\x0dX" +
	"\xd8X\x04\x0b\x0b\x0b\xaf\xe15fI\xf5<\xfc\xf6\xe3" +
	"<;-\x06B\xb4\x9cl\xa4\xaf\xe1\xfd\xad?x\x19" +
	"`w\x98^\xff\xb7\x96\x9f\x87\xb3oL\xcc&`\x7f" +
	">\xec\xdf\xd8\xdf\x1e\xc7\xa9{j\xfd\x89w\x1d\x9b\xee" +
	"lY\xc5\xba\x9d\xde\x9f\xbb\x10V\xa4\xe6&\x032\x02" +
	"\xf6\xe2\x08\xd0\xb9\xa1^\x0b-Yr\xc4\xcb\x19\xa0\x0b" +
	"C\xbd\x13Z\x91\x92\x02\xd8\xdb+@o\x0c\xb5\x16\xee" +
	"5\xee\xb1b\x0ea\x0eN\x9f\xa3\xf3\x0f\xdc\x05W\x86" +
	"\xdc\x86\x8c\x9b\xbc\x0ba\xe1\xa2\x03\xc0\x02\xc2\x02\\\x07" +
	"\x00\x00\xff\xff\xa9\xe1)\x9b"

func RegisterSchema(reg *schemas.Registry) {
	reg.Register(&schemas.Schema{
		String: schema_c13229b64d08e68a,
		Nodes: []uint64{
			0xbf851cafa4aebfbb,
		},
		Compressed: true,
	})
}
