// Code generated by capnpc-go. DO NOT EDIT.

package system

import (
	capnp "capnproto.org/go/capnp/v3"
	text "capnproto.org/go/capnp/v3/encoding/text"
	fc "capnproto.org/go/capnp/v3/flowcontrol"
	schemas "capnproto.org/go/capnp/v3/schemas"
	server "capnproto.org/go/capnp/v3/server"
	context "context"
)

type Proc capnp.Client

// Proc_TypeID is the unique identifier for the type Proc.
const Proc_TypeID = 0xd272f1ad383445bb

func (c Proc) Handle(ctx context.Context, params func(Proc_handle_Params) error) (Proc_handle_Results_Future, capnp.ReleaseFunc) {

	s := capnp.Send{
		Method: capnp.Method{
			InterfaceID:   0xd272f1ad383445bb,
			MethodID:      0,
			InterfaceName: "system.capnp:Proc",
			MethodName:    "handle",
		},
	}
	if params != nil {
		s.ArgsSize = capnp.ObjectSize{DataSize: 0, PointerCount: 1}
		s.PlaceArgs = func(s capnp.Struct) error { return params(Proc_handle_Params(s)) }
	}

	ans, release := capnp.Client(c).SendCall(ctx, s)
	return Proc_handle_Results_Future{Future: ans.Future()}, release

}

func (c Proc) WaitStreaming() error {
	return capnp.Client(c).WaitStreaming()
}

// String returns a string that identifies this capability for debugging
// purposes.  Its format should not be depended on: in particular, it
// should not be used to compare clients.  Use IsSame to compare clients
// for equality.
func (c Proc) String() string {
	return "Proc(" + capnp.Client(c).String() + ")"
}

// AddRef creates a new Client that refers to the same capability as c.
// If c is nil or has resolved to null, then AddRef returns nil.
func (c Proc) AddRef() Proc {
	return Proc(capnp.Client(c).AddRef())
}

// Release releases a capability reference.  If this is the last
// reference to the capability, then the underlying resources associated
// with the capability will be released.
//
// Release will panic if c has already been released, but not if c is
// nil or resolved to null.
func (c Proc) Release() {
	capnp.Client(c).Release()
}

// Resolve blocks until the capability is fully resolved or the Context
// expires.
func (c Proc) Resolve(ctx context.Context) error {
	return capnp.Client(c).Resolve(ctx)
}

func (c Proc) EncodeAsPtr(seg *capnp.Segment) capnp.Ptr {
	return capnp.Client(c).EncodeAsPtr(seg)
}

func (Proc) DecodeFromPtr(p capnp.Ptr) Proc {
	return Proc(capnp.Client{}.DecodeFromPtr(p))
}

// IsValid reports whether c is a valid reference to a capability.
// A reference is invalid if it is nil, has resolved to null, or has
// been released.
func (c Proc) IsValid() bool {
	return capnp.Client(c).IsValid()
}

// IsSame reports whether c and other refer to a capability created by the
// same call to NewClient.  This can return false negatives if c or other
// are not fully resolved: use Resolve if this is an issue.  If either
// c or other are released, then IsSame panics.
func (c Proc) IsSame(other Proc) bool {
	return capnp.Client(c).IsSame(capnp.Client(other))
}

// Update the flowcontrol.FlowLimiter used to manage flow control for
// this client. This affects all future calls, but not calls already
// waiting to send. Passing nil sets the value to flowcontrol.NopLimiter,
// which is also the default.
func (c Proc) SetFlowLimiter(lim fc.FlowLimiter) {
	capnp.Client(c).SetFlowLimiter(lim)
}

// Get the current flowcontrol.FlowLimiter used to manage flow control
// for this client.
func (c Proc) GetFlowLimiter() fc.FlowLimiter {
	return capnp.Client(c).GetFlowLimiter()
}

// A Proc_Server is a Proc with a local implementation.
type Proc_Server interface {
	Handle(context.Context, Proc_handle) error
}

// Proc_NewServer creates a new Server from an implementation of Proc_Server.
func Proc_NewServer(s Proc_Server) *server.Server {
	c, _ := s.(server.Shutdowner)
	return server.New(Proc_Methods(nil, s), s, c)
}

// Proc_ServerToClient creates a new Client from an implementation of Proc_Server.
// The caller is responsible for calling Release on the returned Client.
func Proc_ServerToClient(s Proc_Server) Proc {
	return Proc(capnp.NewClient(Proc_NewServer(s)))
}

// Proc_Methods appends Methods to a slice that invoke the methods on s.
// This can be used to create a more complicated Server.
func Proc_Methods(methods []server.Method, s Proc_Server) []server.Method {
	if cap(methods) == 0 {
		methods = make([]server.Method, 0, 1)
	}

	methods = append(methods, server.Method{
		Method: capnp.Method{
			InterfaceID:   0xd272f1ad383445bb,
			MethodID:      0,
			InterfaceName: "system.capnp:Proc",
			MethodName:    "handle",
		},
		Impl: func(ctx context.Context, call *server.Call) error {
			return s.Handle(ctx, Proc_handle{call})
		},
	})

	return methods
}

// Proc_handle holds the state for a server call to Proc.handle.
// See server.Call for documentation.
type Proc_handle struct {
	*server.Call
}

// Args returns the call's arguments.
func (c Proc_handle) Args() Proc_handle_Params {
	return Proc_handle_Params(c.Call.Args())
}

// AllocResults allocates the results struct.
func (c Proc_handle) AllocResults() (Proc_handle_Results, error) {
	r, err := c.Call.AllocResults(capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return Proc_handle_Results(r), err
}

// Proc_List is a list of Proc.
type Proc_List = capnp.CapList[Proc]

// NewProc creates a new list of Proc.
func NewProc_List(s *capnp.Segment, sz int32) (Proc_List, error) {
	l, err := capnp.NewPointerList(s, sz)
	return capnp.CapList[Proc](l), err
}

type Proc_handle_Params capnp.Struct

// Proc_handle_Params_TypeID is the unique identifier for the type Proc_handle_Params.
const Proc_handle_Params_TypeID = 0xb57a402998306cb9

func NewProc_handle_Params(s *capnp.Segment) (Proc_handle_Params, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Proc_handle_Params(st), err
}

func NewRootProc_handle_Params(s *capnp.Segment) (Proc_handle_Params, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Proc_handle_Params(st), err
}

func ReadRootProc_handle_Params(msg *capnp.Message) (Proc_handle_Params, error) {
	root, err := msg.Root()
	return Proc_handle_Params(root.Struct()), err
}

func (s Proc_handle_Params) String() string {
	str, _ := text.Marshal(0xb57a402998306cb9, capnp.Struct(s))
	return str
}

func (s Proc_handle_Params) EncodeAsPtr(seg *capnp.Segment) capnp.Ptr {
	return capnp.Struct(s).EncodeAsPtr(seg)
}

func (Proc_handle_Params) DecodeFromPtr(p capnp.Ptr) Proc_handle_Params {
	return Proc_handle_Params(capnp.Struct{}.DecodeFromPtr(p))
}

func (s Proc_handle_Params) ToPtr() capnp.Ptr {
	return capnp.Struct(s).ToPtr()
}
func (s Proc_handle_Params) IsValid() bool {
	return capnp.Struct(s).IsValid()
}

func (s Proc_handle_Params) Message() *capnp.Message {
	return capnp.Struct(s).Message()
}

func (s Proc_handle_Params) Segment() *capnp.Segment {
	return capnp.Struct(s).Segment()
}
func (s Proc_handle_Params) Event() ([]byte, error) {
	p, err := capnp.Struct(s).Ptr(0)
	return []byte(p.Data()), err
}

func (s Proc_handle_Params) HasEvent() bool {
	return capnp.Struct(s).HasPtr(0)
}

func (s Proc_handle_Params) SetEvent(v []byte) error {
	return capnp.Struct(s).SetData(0, v)
}

// Proc_handle_Params_List is a list of Proc_handle_Params.
type Proc_handle_Params_List = capnp.StructList[Proc_handle_Params]

// NewProc_handle_Params creates a new list of Proc_handle_Params.
func NewProc_handle_Params_List(s *capnp.Segment, sz int32) (Proc_handle_Params_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return capnp.StructList[Proc_handle_Params](l), err
}

// Proc_handle_Params_Future is a wrapper for a Proc_handle_Params promised by a client call.
type Proc_handle_Params_Future struct{ *capnp.Future }

func (f Proc_handle_Params_Future) Struct() (Proc_handle_Params, error) {
	p, err := f.Future.Ptr()
	return Proc_handle_Params(p.Struct()), err
}

type Proc_handle_Results capnp.Struct

// Proc_handle_Results_TypeID is the unique identifier for the type Proc_handle_Results.
const Proc_handle_Results_TypeID = 0xb65a10541fbdb6e7

func NewProc_handle_Results(s *capnp.Segment) (Proc_handle_Results, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return Proc_handle_Results(st), err
}

func NewRootProc_handle_Results(s *capnp.Segment) (Proc_handle_Results, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return Proc_handle_Results(st), err
}

func ReadRootProc_handle_Results(msg *capnp.Message) (Proc_handle_Results, error) {
	root, err := msg.Root()
	return Proc_handle_Results(root.Struct()), err
}

func (s Proc_handle_Results) String() string {
	str, _ := text.Marshal(0xb65a10541fbdb6e7, capnp.Struct(s))
	return str
}

func (s Proc_handle_Results) EncodeAsPtr(seg *capnp.Segment) capnp.Ptr {
	return capnp.Struct(s).EncodeAsPtr(seg)
}

func (Proc_handle_Results) DecodeFromPtr(p capnp.Ptr) Proc_handle_Results {
	return Proc_handle_Results(capnp.Struct{}.DecodeFromPtr(p))
}

func (s Proc_handle_Results) ToPtr() capnp.Ptr {
	return capnp.Struct(s).ToPtr()
}
func (s Proc_handle_Results) IsValid() bool {
	return capnp.Struct(s).IsValid()
}

func (s Proc_handle_Results) Message() *capnp.Message {
	return capnp.Struct(s).Message()
}

func (s Proc_handle_Results) Segment() *capnp.Segment {
	return capnp.Struct(s).Segment()
}

// Proc_handle_Results_List is a list of Proc_handle_Results.
type Proc_handle_Results_List = capnp.StructList[Proc_handle_Results]

// NewProc_handle_Results creates a new list of Proc_handle_Results.
func NewProc_handle_Results_List(s *capnp.Segment, sz int32) (Proc_handle_Results_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0}, sz)
	return capnp.StructList[Proc_handle_Results](l), err
}

// Proc_handle_Results_Future is a wrapper for a Proc_handle_Results promised by a client call.
type Proc_handle_Results_Future struct{ *capnp.Future }

func (f Proc_handle_Results_Future) Struct() (Proc_handle_Results, error) {
	p, err := f.Future.Ptr()
	return Proc_handle_Results(p.Struct()), err
}

const schema_910958683ac350d8 = "x\xda|\xcf\xb1J\xc3@\x00\xc6\xf1\xef\xbb\\\x8c\x0a" +
	"\x81\x9eqT\x8a\xe0R\xd0b\xd5A\xb24\x0e\xe2z" +
	"'\x0a\xe2\x16\xe2A\x874-\xb9((\xbe\x84N\xbe" +
	"\x84\xa3\x05\x11'\xdf\xc0\xc97p\xf6\x09NJiG" +
	"\xe7\xef\xcf\x0f\xbe\xd6C&{\xf1H@\x98\x8dp\xc9" +
	"\xbf\x95{\xcf\x9d\xec\xfe\x15j\x8d@\xc8\x088\xd8\xa4" +
	" \x98l\xb1\x0f\xfa\x9f\xc9G\xfb\xbcu5\x99\x05r" +
	"\xba\x1fs\x95\x90\xfe\xfd\xe4\xf0\xe8\xe5\xb7\xfe\x82\x8a\x03" +
	"\xff\xad?\xd3\xc1\xe5\xca#\xc0\xa4\xc3\xa7\xa47\x95\x92" +
	"]\x9e&\x17\x8c\xb0\xe3\xdd\x9dk\xec\xb0[\x04\xf9\xb8" +
	"\x1a\xa7\xba\x1e\x15\xddA^]\x97v[\xe7u>t" +
	"02\x90\x80$\xa0\xe2}\xc0,\x074\xeb\x82m{" +
	"k\xab\x861\x04c\xf0\x1f\xe6\xcc\xba\x9b\xb2q\xc0\xa2" +
	"\xe1\xbca\xa1I#\x83\x10X\xfc\xe5\xfc\x97R)\x84" +
	"\x0a\xa3\xfe\xcc\xc9\xa8\xc9\xbf\x00\x00\x00\xff\xffeSK" +
	"\xe4"

func RegisterSchema(reg *schemas.Registry) {
	reg.Register(&schemas.Schema{
		String: schema_910958683ac350d8,
		Nodes: []uint64{
			0xb57a402998306cb9,
			0xb65a10541fbdb6e7,
			0xd272f1ad383445bb,
		},
		Compressed: true,
	})
}
