// Code generated by capnpc-go. DO NOT EDIT.

package glia

import (
	capnp "capnproto.org/go/capnp/v3"
	text "capnproto.org/go/capnp/v3/encoding/text"
	schemas "capnproto.org/go/capnp/v3/schemas"
)

type CallData capnp.Struct

// CallData_TypeID is the unique identifier for the type CallData.
const CallData_TypeID = 0xb063cb51875ce2d8

func NewCallData(s *capnp.Segment) (CallData, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return CallData(st), err
}

func NewRootCallData(s *capnp.Segment) (CallData, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return CallData(st), err
}

func ReadRootCallData(msg *capnp.Message) (CallData, error) {
	root, err := msg.Root()
	return CallData(root.Struct()), err
}

func (s CallData) String() string {
	str, _ := text.Marshal(0xb063cb51875ce2d8, capnp.Struct(s))
	return str
}

func (s CallData) EncodeAsPtr(seg *capnp.Segment) capnp.Ptr {
	return capnp.Struct(s).EncodeAsPtr(seg)
}

func (CallData) DecodeFromPtr(p capnp.Ptr) CallData {
	return CallData(capnp.Struct{}.DecodeFromPtr(p))
}

func (s CallData) ToPtr() capnp.Ptr {
	return capnp.Struct(s).ToPtr()
}
func (s CallData) IsValid() bool {
	return capnp.Struct(s).IsValid()
}

func (s CallData) Message() *capnp.Message {
	return capnp.Struct(s).Message()
}

func (s CallData) Segment() *capnp.Segment {
	return capnp.Struct(s).Segment()
}
func (s CallData) Stack() (capnp.UInt64List, error) {
	p, err := capnp.Struct(s).Ptr(0)
	return capnp.UInt64List(p.List()), err
}

func (s CallData) HasStack() bool {
	return capnp.Struct(s).HasPtr(0)
}

func (s CallData) SetStack(v capnp.UInt64List) error {
	return capnp.Struct(s).SetPtr(0, v.ToPtr())
}

// NewStack sets the stack field to a newly
// allocated capnp.UInt64List, preferring placement in s's segment.
func (s CallData) NewStack(n int32) (capnp.UInt64List, error) {
	l, err := capnp.NewUInt64List(capnp.Struct(s).Segment(), n)
	if err != nil {
		return capnp.UInt64List{}, err
	}
	err = capnp.Struct(s).SetPtr(0, l.ToPtr())
	return l, err
}
func (s CallData) Method() (string, error) {
	p, err := capnp.Struct(s).Ptr(1)
	return p.Text(), err
}

func (s CallData) HasMethod() bool {
	return capnp.Struct(s).HasPtr(1)
}

func (s CallData) MethodBytes() ([]byte, error) {
	p, err := capnp.Struct(s).Ptr(1)
	return p.TextBytes(), err
}

func (s CallData) SetMethod(v string) error {
	return capnp.Struct(s).SetText(1, v)
}

// CallData_List is a list of CallData.
type CallData_List = capnp.StructList[CallData]

// NewCallData creates a new list of CallData.
func NewCallData_List(s *capnp.Segment, sz int32) (CallData_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2}, sz)
	return capnp.StructList[CallData](l), err
}

// CallData_Future is a wrapper for a CallData promised by a client call.
type CallData_Future struct{ *capnp.Future }

func (f CallData_Future) Struct() (CallData, error) {
	p, err := f.Future.Ptr()
	return CallData(p.Struct()), err
}

type Header capnp.Struct

// Header_TypeID is the unique identifier for the type Header.
const Header_TypeID = 0xb00b0243e9bd824d

func NewHeader(s *capnp.Segment) (Header, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 4})
	return Header(st), err
}

func NewRootHeader(s *capnp.Segment) (Header, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 4})
	return Header(st), err
}

func ReadRootHeader(msg *capnp.Message) (Header, error) {
	root, err := msg.Root()
	return Header(root.Struct()), err
}

func (s Header) String() string {
	str, _ := text.Marshal(0xb00b0243e9bd824d, capnp.Struct(s))
	return str
}

func (s Header) EncodeAsPtr(seg *capnp.Segment) capnp.Ptr {
	return capnp.Struct(s).EncodeAsPtr(seg)
}

func (Header) DecodeFromPtr(p capnp.Ptr) Header {
	return Header(capnp.Struct{}.DecodeFromPtr(p))
}

func (s Header) ToPtr() capnp.Ptr {
	return capnp.Struct(s).ToPtr()
}
func (s Header) IsValid() bool {
	return capnp.Struct(s).IsValid()
}

func (s Header) Message() *capnp.Message {
	return capnp.Struct(s).Message()
}

func (s Header) Segment() *capnp.Segment {
	return capnp.Struct(s).Segment()
}
func (s Header) Peer() ([]byte, error) {
	p, err := capnp.Struct(s).Ptr(0)
	return []byte(p.Data()), err
}

func (s Header) HasPeer() bool {
	return capnp.Struct(s).HasPtr(0)
}

func (s Header) SetPeer(v []byte) error {
	return capnp.Struct(s).SetData(0, v)
}

func (s Header) Proc() (string, error) {
	p, err := capnp.Struct(s).Ptr(1)
	return p.Text(), err
}

func (s Header) HasProc() bool {
	return capnp.Struct(s).HasPtr(1)
}

func (s Header) ProcBytes() ([]byte, error) {
	p, err := capnp.Struct(s).Ptr(1)
	return p.TextBytes(), err
}

func (s Header) SetProc(v string) error {
	return capnp.Struct(s).SetText(1, v)
}

func (s Header) Method() (string, error) {
	p, err := capnp.Struct(s).Ptr(2)
	return p.Text(), err
}

func (s Header) HasMethod() bool {
	return capnp.Struct(s).HasPtr(2)
}

func (s Header) MethodBytes() ([]byte, error) {
	p, err := capnp.Struct(s).Ptr(2)
	return p.TextBytes(), err
}

func (s Header) SetMethod(v string) error {
	return capnp.Struct(s).SetText(2, v)
}

func (s Header) Stack() (capnp.UInt64List, error) {
	p, err := capnp.Struct(s).Ptr(3)
	return capnp.UInt64List(p.List()), err
}

func (s Header) HasStack() bool {
	return capnp.Struct(s).HasPtr(3)
}

func (s Header) SetStack(v capnp.UInt64List) error {
	return capnp.Struct(s).SetPtr(3, v.ToPtr())
}

// NewStack sets the stack field to a newly
// allocated capnp.UInt64List, preferring placement in s's segment.
func (s Header) NewStack(n int32) (capnp.UInt64List, error) {
	l, err := capnp.NewUInt64List(capnp.Struct(s).Segment(), n)
	if err != nil {
		return capnp.UInt64List{}, err
	}
	err = capnp.Struct(s).SetPtr(3, l.ToPtr())
	return l, err
}

// Header_List is a list of Header.
type Header_List = capnp.StructList[Header]

// NewHeader creates a new list of Header.
func NewHeader_List(s *capnp.Segment, sz int32) (Header_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 4}, sz)
	return capnp.StructList[Header](l), err
}

// Header_Future is a wrapper for a Header promised by a client call.
type Header_Future struct{ *capnp.Future }

func (f Header_Future) Struct() (Header, error) {
	p, err := f.Future.Ptr()
	return Header(p.Struct()), err
}

type Result capnp.Struct

// Result_TypeID is the unique identifier for the type Result.
const Result_TypeID = 0x85366ce08c8fc52b

func NewResult(s *capnp.Segment) (Result, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 2})
	return Result(st), err
}

func NewRootResult(s *capnp.Segment) (Result, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 2})
	return Result(st), err
}

func ReadRootResult(msg *capnp.Message) (Result, error) {
	root, err := msg.Root()
	return Result(root.Struct()), err
}

func (s Result) String() string {
	str, _ := text.Marshal(0x85366ce08c8fc52b, capnp.Struct(s))
	return str
}

func (s Result) EncodeAsPtr(seg *capnp.Segment) capnp.Ptr {
	return capnp.Struct(s).EncodeAsPtr(seg)
}

func (Result) DecodeFromPtr(p capnp.Ptr) Result {
	return Result(capnp.Struct{}.DecodeFromPtr(p))
}

func (s Result) ToPtr() capnp.Ptr {
	return capnp.Struct(s).ToPtr()
}
func (s Result) IsValid() bool {
	return capnp.Struct(s).IsValid()
}

func (s Result) Message() *capnp.Message {
	return capnp.Struct(s).Message()
}

func (s Result) Segment() *capnp.Segment {
	return capnp.Struct(s).Segment()
}
func (s Result) Stack() (capnp.UInt64List, error) {
	p, err := capnp.Struct(s).Ptr(0)
	return capnp.UInt64List(p.List()), err
}

func (s Result) HasStack() bool {
	return capnp.Struct(s).HasPtr(0)
}

func (s Result) SetStack(v capnp.UInt64List) error {
	return capnp.Struct(s).SetPtr(0, v.ToPtr())
}

// NewStack sets the stack field to a newly
// allocated capnp.UInt64List, preferring placement in s's segment.
func (s Result) NewStack(n int32) (capnp.UInt64List, error) {
	l, err := capnp.NewUInt64List(capnp.Struct(s).Segment(), n)
	if err != nil {
		return capnp.UInt64List{}, err
	}
	err = capnp.Struct(s).SetPtr(0, l.ToPtr())
	return l, err
}
func (s Result) Status() Result_Status {
	return Result_Status(capnp.Struct(s).Uint16(0))
}

func (s Result) SetStatus(v Result_Status) {
	capnp.Struct(s).SetUint16(0, uint16(v))
}

func (s Result) Info() (string, error) {
	p, err := capnp.Struct(s).Ptr(1)
	return p.Text(), err
}

func (s Result) HasInfo() bool {
	return capnp.Struct(s).HasPtr(1)
}

func (s Result) InfoBytes() ([]byte, error) {
	p, err := capnp.Struct(s).Ptr(1)
	return p.TextBytes(), err
}

func (s Result) SetInfo(v string) error {
	return capnp.Struct(s).SetText(1, v)
}

// Result_List is a list of Result.
type Result_List = capnp.StructList[Result]

// NewResult creates a new list of Result.
func NewResult_List(s *capnp.Segment, sz int32) (Result_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 2}, sz)
	return capnp.StructList[Result](l), err
}

// Result_Future is a wrapper for a Result promised by a client call.
type Result_Future struct{ *capnp.Future }

func (f Result_Future) Struct() (Result, error) {
	p, err := f.Future.Ptr()
	return Result(p.Struct()), err
}

type Result_Status uint16

// Result_Status_TypeID is the unique identifier for the type Result_Status.
const Result_Status_TypeID = 0xe56103a0267e998a

// Values of Result_Status.
const (
	Result_Status_unset          Result_Status = 0
	Result_Status_ok             Result_Status = 1
	Result_Status_invalidRequest Result_Status = 2
	Result_Status_routingError   Result_Status = 3
	Result_Status_procNotFound   Result_Status = 4
	Result_Status_invalidMethod  Result_Status = 5
	Result_Status_methodNotFound Result_Status = 6
	Result_Status_guestError     Result_Status = 7
)

// String returns the enum's constant name.
func (c Result_Status) String() string {
	switch c {
	case Result_Status_unset:
		return "unset"
	case Result_Status_ok:
		return "ok"
	case Result_Status_invalidRequest:
		return "invalidRequest"
	case Result_Status_routingError:
		return "routingError"
	case Result_Status_procNotFound:
		return "procNotFound"
	case Result_Status_invalidMethod:
		return "invalidMethod"
	case Result_Status_methodNotFound:
		return "methodNotFound"
	case Result_Status_guestError:
		return "guestError"

	default:
		return ""
	}
}

// Result_StatusFromString returns the enum value with a name,
// or the zero value if there's no such value.
func Result_StatusFromString(c string) Result_Status {
	switch c {
	case "unset":
		return Result_Status_unset
	case "ok":
		return Result_Status_ok
	case "invalidRequest":
		return Result_Status_invalidRequest
	case "routingError":
		return Result_Status_routingError
	case "procNotFound":
		return Result_Status_procNotFound
	case "invalidMethod":
		return Result_Status_invalidMethod
	case "methodNotFound":
		return Result_Status_methodNotFound
	case "guestError":
		return Result_Status_guestError

	default:
		return 0
	}
}

type Result_Status_List = capnp.EnumList[Result_Status]

func NewResult_Status_List(s *capnp.Segment, sz int32) (Result_Status_List, error) {
	return capnp.NewEnumList[Result_Status](s, sz)
}

const schema_f381800d6f8057ad = "x\xdat\x91Ok\x14A\x10\xc5\xdf\xeb\xdeM\x0c\xc9" +
	":\xb6\x13\x10\xbc$\x07\x11T\x8c\x1aP4\x106\x92" +
	"l0\x87\x84t\x14t%B\x9a\xdd1\x193\x99Y" +
	"gz\x14<H\x14E\x10A\xc8\xcd\x9bx\x17\xfd\x02" +
	"\x1e%\xa0_\xc0\xa3\x08\"9x\x10?\xc0HO\xd8" +
	"\xfc\x11=uu\xf5\xebz\xbf\xaa:{\x86\x13\x95s" +
	"\xb5M\x01\xa1\x87\xab=\xc5\xa9\x8f/_|\x8d.<" +
	"\x85\xee'\x8b\xb7\xd7\xd7\x93\xda\xfa\xa3\xdf\xa8\x8a^\xc0" +
	"\x9f\xe1\x86\xafy\x04\xf0\x9b|\x07\x16\xb3\x8f?lM" +
	"\x8a\xfe\xf7P\xfd{\xb5\x15\xa7\xa5\xd8\xf0\xfb\xca_U" +
	"\xf1\x03,\xbe|[|\xa6?\xb7\xfe\xd6\x96\x8a-\xf1" +
	"\xc6\xffUF?\xc5}\xb0x\xfe\xea\xe1\xf1\xd7\xd2|" +
	"\x87:,v\x89@\xbf)?\xf9\x81tB#7q" +
	"\xbaX\x8eB3\xd22\x1d\xc6\x9d\xb1\x85 \xcb#Z" +
	"]\xe1\xde\x02\x1c\xab_\xb5\xc6\xe6\x99\x1e\x90\x15\xa0B" +
	"@5F\x01=!\xa9\x17\x05\xc9A\xba\\s\x0c\xd0" +
	"\xd7$\xf5\x92\xa0\x12\x1c\xa4\x00\xd4\xad\x93\x80\xbe!\xa9" +
	"\xdb\x82C\x995\xadU\x1e\x04\xe7%\xd9\x07\xe1\xc2z" +
	"V\xd6\xa6\xb7\xeb\x08\xd2\x03\xbd0\xbe\x9dp\x00\x82\x03" +
	"\xe0>\xcc+\x81i\x07L\xe7I}h\x87\xc88\xa3" +
	"EI\xbd\"\xa8\xbaH\x81K.I\xea\xc8!\x89m" +
	"\xa4\xd0q\xb6%uGPI9H\x09\xa85\xd7\xd0" +
	"\x8a\xa4~\"\xe8u\x82 e\x0d\x825\xd0\xeb\xa4I" +
	"\xab\x8bQ_\x0b\xecJ\xd2\xee^\xff\xd9\xd0>\xd4I" +
	"\x13ES\xbd\xc6\x1a\x07{`\x07\xf6\x84s;&\xa9" +
	"'\xf6\xc0\x8e;\xae\x8b\x92z\xea\x7f\xa3\xda\xef\xbec" +
	"$\xba\xab\xb3#\xdb\x8b\x02\x9c\xddp\xd9ms\xd4\xcd" +
	"S\xe9\xa3\x00\x85\x9ay\x00P\xaa\xc6\x1d\x80\x15u\xd9" +
	"\x1dU5\x9e\x02\xecQ\x97\xdc[\xaf:\x7f\x13\x18\xca" +
	"\xe3,\xb02Y-\xc2\xf8\x9e\x89\xc2\xf6\x02\xea\xc1\xdd" +
	"<\xc8l\x91&\xb9\x0d\xe3\xe5\x06\xbc4M\xd2\xc2M" +
	"g.\xb1\xd3\xf0\x92<nw\xe5\xb3\x18*Q\x8bm" +
	"\xe2\xb9\x04u;]\x0a\x96]\x91F\x9aB&\xe9\x9f" +
	"\x00\x00\x00\xff\xff\xf9\x8e\xbe\xc1"

func RegisterSchema(reg *schemas.Registry) {
	reg.Register(&schemas.Schema{
		String: schema_f381800d6f8057ad,
		Nodes: []uint64{
			0x85366ce08c8fc52b,
			0xb00b0243e9bd824d,
			0xb063cb51875ce2d8,
			0xe56103a0267e998a,
		},
		Compressed: true,
	})
}
