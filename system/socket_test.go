package system_test

import (
	"bytes"
	"context"
	"io"
	"net"
	"strings"
	"testing"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"github.com/libp2p/go-libp2p"
	"github.com/stretchr/testify/require"
	"github.com/wetware/go/auth"
	"github.com/wetware/go/system"
)

func TestCapnpSocket(t *testing.T) {
	t.Parallel()

	sock := system.Socket{}
	host, guest := net.Pipe()

	term := auth.Terminal_ServerToClient(terminalTestServer{T: t})

	conn := rpc.NewConn(rpc.NewStreamTransport(struct {
		io.Reader
		io.WriteCloser
	}{
		Reader:      sock.Connect(context.TODO(), system.NewReadPipe(host)),
		WriteCloser: sock.Bind(context.TODO(), system.NewWritePipe(guest)),
	}), &rpc.Options{
		BootstrapClient: capnp.Client(term), // TODO:  add test client.
	})
	defer conn.Close()

	client := conn.Bootstrap(context.TODO())
	defer client.Release()

	err := client.Resolve(context.TODO())
	require.NoError(t, err)

	h, err := libp2p.New(
		libp2p.NoListenAddrs,
		libp2p.NoTransports,
		libp2p.NoSecurity)
	require.NoError(t, err)

	f, release := auth.Terminal(client).Login(context.TODO(), func(prompt auth.Terminal_login_Params) error {
		pk := h.Peerstore().PrivKey(h.ID())
		signer := auth.Signer_ServerToClient(&auth.SignOnce{PrivKey: pk})
		return prompt.SetAccount(signer)
	})
	defer release()

	reader := f.Stdio().Reader()
	fr, releaseR := reader.Read(context.TODO(), func(read auth.ReadPipe_read_Params) error {
		read.SetSize(1024)
		return nil
	})
	defer releaseR()
	res, err := fr.Struct()
	require.NoError(t, err)
	b, err := res.Data()
	require.NoError(t, err)
	require.Equal(t, "stdin test", string(b))

	writer := f.Stdio().Writer()
	fw, releaseW := writer.Write(context.TODO(), func(write auth.WritePipe_write_Params) error {
		return write.SetData([]byte("test stdout"))
	})
	defer releaseW()

	resW, errW := fw.Struct()
	require.NoError(t, errW)
	require.Equal(t, int64(len("test stdout")), resW.N())

}

type terminalTestServer struct {
	T      *testing.T
	Reader auth.ReadPipe
	Writer auth.WritePipe
}

func (t terminalTestServer) Login(ctx context.Context, login auth.Terminal_login) error {
	res, err := login.AllocResults()
	require.NoError(t.T, err)

	sock, err := res.NewStdio()
	require.NoError(t.T, err)

	rbuf := strings.NewReader("stdin test")
	err = sock.SetReader(system.NewReadPipe(rbuf))
	require.NoError(t.T, err)

	wbuf := new(bytes.Buffer)
	err = sock.SetWriter(system.NewWritePipe(nopCloser{wbuf}))
	require.NoError(t.T, err)

	return nil
}

func TestPipeReader(t *testing.T) {
	t.Parallel()

	buf := strings.NewReader("test")
	pipe := system.NewReadPipe(buf)
	defer pipe.Release()

	r := system.Socket{}.Connect(context.TODO(), pipe)
	b, err := io.ReadAll(r)
	require.NoError(t, err)
	require.Equal(t, "test", string(b))
}

func TestPipeWriter(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	pipe := system.NewWritePipe(nopCloser{Writer: buf})
	defer pipe.Release()

	wc := system.Socket{}.Bind(context.TODO(), pipe)
	n, err := io.Copy(wc, strings.NewReader("test"))
	require.NoError(t, err)
	require.Equal(t, int64(len("test")), n)
	require.Equal(t, "test", buf.String())
	require.NoError(t, wc.Close())
}

func TestReadPipe(t *testing.T) {
	t.Parallel()

	pipe := system.NewReadPipe(strings.NewReader("test"))
	defer pipe.Release()

	f, release := pipe.Read(context.TODO(), func(read auth.ReadPipe_read_Params) error {
		read.SetSize(int64(len("test")))
		return nil
	})
	defer release()

	res, err := f.Struct()
	require.NoError(t, err)
	data, err := res.Data()
	require.NoError(t, err)
	require.Equal(t, "test", string(data))
	require.False(t, res.Eof())

	f, release = pipe.Read(context.TODO(), func(read auth.ReadPipe_read_Params) error {
		read.SetSize(int64(len("test")))
		return nil
	})
	defer release()

	res, err = f.Struct()
	require.NoError(t, err)
	require.True(t, res.Eof(), "should report EOF")
}

func TestWritePipe(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	pipe := system.NewWritePipe(nopCloser{Writer: buf})
	defer pipe.Release()

	f, release := pipe.Write(context.TODO(), func(write auth.WritePipe_write_Params) error {
		return write.SetData([]byte("test"))
	})
	defer release()

	res, err := f.Struct()
	require.NoError(t, err)
	require.Equal(t, int64(len("test")), res.N())
}

type nopCloser struct{ io.Writer }

func (nopCloser) Close() error { return nil }
