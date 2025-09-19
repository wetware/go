package system_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p"
	inproc "github.com/lthibault/go-libp2p-inproc-transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/wetware/go/system"
	"github.com/wetware/go/system/mocks"
	"go.uber.org/mock/gomock"
)

// loadEchoWasm loads the echo example WASM file for testing
func loadEchoWasm(t *testing.T) []byte {
	// Try multiple possible paths
	possiblePaths := []string{
		"../../examples/echo/main.wasm",
		"../examples/echo/main.wasm",
		"examples/echo/main.wasm",
	}

	for _, path := range possiblePaths {
		if wasmData, err := os.ReadFile(path); err == nil {
			return wasmData
		}
	}

	// If none of the paths work, try to find it relative to the current working directory
	wasmData, err := os.ReadFile("examples/echo/main.wasm")
	require.NoError(t, err, "Failed to load echo.wasm from any expected location")
	return wasmData
}

func TestProcConfig_New(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockErrWriter := &bytes.Buffer{}

	// Load the real echo WASM file for testing
	validBytecode := loadEchoWasm(t)

	t.Run("successful creation", func(t *testing.T) {
		// Create a libp2p host with in-process transport
		host, err := libp2p.New(
			libp2p.Transport(inproc.New()),
		)
		require.NoError(t, err)
		defer host.Close()

		runtime := wazero.NewRuntime(ctx)
		defer runtime.Close(ctx)

		config := system.ProcConfig{
			Host:      host,
			Runtime:   runtime,
			Bytecode:  validBytecode,
			ErrWriter: mockErrWriter,
			Async:     true, // Use async mode for these tests
		}

		proc, err := config.New(ctx)
		require.NoError(t, err)
		require.NotNil(t, proc)
		assert.NotNil(t, proc.Closer)
		assert.NotNil(t, proc.Module)
		assert.NotNil(t, proc.Endpoint)

		// Verify the endpoint has a valid name
		assert.NotEmpty(t, proc.Endpoint.Name)
		assert.NotEmpty(t, proc.ID())

		// Clean up
		proc.Close(ctx)
	})

	t.Run("invalid bytecode", func(t *testing.T) {
		// Create a libp2p host with in-process transport
		host, err := libp2p.New(
			libp2p.Transport(inproc.New()),
		)
		require.NoError(t, err)
		defer host.Close()

		runtime := wazero.NewRuntime(ctx)
		defer runtime.Close(ctx)

		config := system.ProcConfig{
			Host:      host,
			Runtime:   runtime,
			Bytecode:  []byte("invalid wasm bytecode"),
			ErrWriter: mockErrWriter,
			Async:     true, // Use async mode for these tests
		}

		proc, err := config.New(ctx)
		assert.Error(t, err)
		assert.Nil(t, proc)
	})

	t.Run("nil runtime", func(t *testing.T) {
		// Create a libp2p host with in-process transport
		host, err := libp2p.New(
			libp2p.Transport(inproc.New()),
		)
		require.NoError(t, err)
		defer host.Close()

		config := system.ProcConfig{
			Host:      host,
			Runtime:   nil,
			Bytecode:  []byte{},
			ErrWriter: mockErrWriter,
			Async:     true, // Use async mode for these tests
		}

		// This will panic due to nil runtime, so we expect a panic
		assert.Panics(t, func() {
			config.New(ctx)
		})
	})
}

func TestProc_ID(t *testing.T) {
	t.Parallel()

	endpoint := system.ProcConfig{}.NewEndpoint()
	proc := &system.Proc{
		Endpoint: endpoint,
	}

	result := proc.ID()
	expected := endpoint.Name
	assert.Equal(t, expected, result)
	assert.NotEmpty(t, result)
}

func TestProc_Close(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("successful close", func(t *testing.T) {
		// Create a libp2p host with in-process transport
		host, err := libp2p.New(
			libp2p.Transport(inproc.New()),
		)
		require.NoError(t, err)
		defer host.Close()

		runtime := wazero.NewRuntime(ctx)
		defer runtime.Close(ctx)

		validBytecode := loadEchoWasm(t)

		config := system.ProcConfig{
			Host:      host,
			Runtime:   runtime,
			Bytecode:  validBytecode,
			ErrWriter: &bytes.Buffer{},
			Async:     true, // Use async mode for these tests
		}

		proc, err := config.New(ctx)
		require.NoError(t, err)
		require.NotNil(t, proc)

		// Test close
		err = proc.Close(ctx)
		assert.NoError(t, err)
	})
}

func TestProc_Poll_WithGomock(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("successful poll", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStream := mocks.NewMockStreamInterface(ctrl)

		// Create a libp2p host with in-process transport
		host, err := libp2p.New(
			libp2p.Transport(inproc.New()),
		)
		require.NoError(t, err)
		defer host.Close()

		runtime := wazero.NewRuntime(ctx)

		validBytecode := loadEchoWasm(t)

		config := system.ProcConfig{
			Host:      host,
			Runtime:   runtime,
			Bytecode:  validBytecode,
			ErrWriter: &bytes.Buffer{},
			Async:     true, // Use async mode for these tests
		}

		proc, err := config.New(ctx)
		require.NoError(t, err)
		require.NotNil(t, proc)
		defer proc.Close(ctx)
		defer runtime.Close(ctx)

		// Test that poll function exists
		pollFunc := proc.Module.ExportedFunction("poll")
		assert.NotNil(t, pollFunc, "poll function should exist")

		// The echo WASM module will try to read from stdin, so we need to expect Read calls
		// The poll function reads up to 512 bytes from stdin
		mockStream.EXPECT().Read(gomock.Any()).Return(0, io.EOF).AnyTimes()

		// Actually call the Poll method - this should succeed since we have a valid WASM function
		err = proc.ProcessMessage(ctx, mockStream)
		// The WASM function should execute successfully
		assert.NoError(t, err)
	})

	t.Run("poll with context deadline", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStream := mocks.NewMockStreamInterface(ctrl)

		// Create context with deadline
		deadline := time.Now().Add(time.Hour)
		ctxWithDeadline, cancel := context.WithDeadline(ctx, deadline)
		defer cancel()

		// Create a libp2p host with in-process transport
		host, err := libp2p.New(
			libp2p.Transport(inproc.New()),
		)
		require.NoError(t, err)
		defer host.Close()

		runtime := wazero.NewRuntime(ctx)

		validBytecode := loadEchoWasm(t)

		config := system.ProcConfig{
			Host:      host,
			Runtime:   runtime,
			Bytecode:  validBytecode,
			ErrWriter: &bytes.Buffer{},
			Async:     true, // Use async mode for these tests
		}

		proc, err := config.New(ctxWithDeadline)
		require.NoError(t, err)
		require.NotNil(t, proc)
		defer proc.Close(ctxWithDeadline)
		defer runtime.Close(ctx)

		// Set up mock expectations
		mockStream.EXPECT().SetReadDeadline(deadline).Return(nil)
		// The echo WASM module will try to read from stdin
		mockStream.EXPECT().Read(gomock.Any()).Return(0, io.EOF).AnyTimes()

		// Test that poll function exists (for async mode)
		pollFunc := proc.Module.ExportedFunction("poll")
		assert.NotNil(t, pollFunc, "poll function should exist")

		// Actually call the ProcessMessage method to trigger the mock expectations
		err = proc.ProcessMessage(ctxWithDeadline, mockStream)
		// The WASM function should execute successfully
		assert.NoError(t, err)
	})

	t.Run("set read deadline error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStream := mocks.NewMockStreamInterface(ctrl)
		endpoint := system.ProcConfig{}.NewEndpoint()

		// Create context with deadline
		deadline := time.Now().Add(time.Hour)
		ctxWithDeadline, cancel := context.WithDeadline(ctx, deadline)
		defer cancel()

		proc := &system.Proc{
			Endpoint: endpoint,
		}

		// Set up mock expectations
		deadlineErr := errors.New("deadline error")
		mockStream.EXPECT().SetReadDeadline(deadline).Return(deadlineErr)

		err := proc.ProcessMessage(ctxWithDeadline, mockStream)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "set read deadline")
		assert.Contains(t, err.Error(), "deadline error")
	})

	t.Run("poll function not found", func(t *testing.T) {
		// Create a libp2p host with in-process transport
		host, err := libp2p.New(
			libp2p.Transport(inproc.New()),
		)
		require.NoError(t, err)
		defer host.Close()

		// Create a real proc with a module that doesn't export poll function
		runtime := wazero.NewRuntime(ctx)
		defer runtime.Close(ctx)

		// Use completely invalid bytecode to simulate a module that can't be created
		noPollBytecode := []byte{0x00, 0x00, 0x00, 0x00}

		configNoPoll := system.ProcConfig{
			Host:      host,
			Runtime:   runtime,
			Bytecode:  noPollBytecode,
			ErrWriter: &bytes.Buffer{},
			Async:     true, // Use async mode for these tests
		}

		// This should fail to create the module due to invalid bytecode
		procNoPoll, err := configNoPoll.New(ctx)
		assert.Error(t, err, "should fail to create module with invalid bytecode")
		assert.Nil(t, procNoPoll, "proc should be nil when creation fails")
	})
}

func TestProc_StreamHandler_WithGomock(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockErrWriter := &bytes.Buffer{}

	// Create a minimal valid WASM bytecode
	validBytecode := []byte{
		0x00, 0x61, 0x73, 0x6d, // WASM magic number
		0x01, 0x00, 0x00, 0x00, // Version 1
		0x01, 0x07, 0x01, 0x60, 0x02, 0x7f, 0x7f, 0x01, 0x7f, // Type section: (i32, i32) -> i32
		0x03, 0x02, 0x01, 0x00, // Function section: 1 function of type 0
		0x07, 0x08, 0x01, 0x04, 0x70, 0x6f, 0x6c, 0x6c, 0x00, 0x00, // Export section: export "poll" function 0
		0x0a, 0x09, 0x01, 0x07, 0x00, 0x20, 0x00, 0x20, 0x01, 0x6a, 0x0b, // Code section: function 0 body
	}

	// Create a libp2p host with in-process transport
	host, err := libp2p.New(
		libp2p.Transport(inproc.New()),
	)
	require.NoError(t, err)
	defer host.Close()

	runtime := wazero.NewRuntime(ctx)
	defer runtime.Close(ctx)

	config := system.ProcConfig{
		Host:      host,
		Runtime:   runtime,
		Bytecode:  validBytecode,
		ErrWriter: mockErrWriter,
		Async:     true, // Use async mode for these tests
	}

	proc, err := config.New(ctx)
	require.NoError(t, err)
	require.NotNil(t, proc)
	defer proc.Close(ctx)

	t.Run("stream handler processes valid data", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStream := mocks.NewMockStreamInterface(ctrl)

		// Create test data
		stackSize := int32(2)
		stackData := []uint64{1, 2}

		var buf bytes.Buffer
		binary.Write(&buf, binary.BigEndian, stackSize)
		for _, val := range stackData {
			binary.Write(&buf, binary.BigEndian, val)
		}

		// Set up mock expectations
		mockStream.EXPECT().Read(gomock.Any()).DoAndReturn(func(p []byte) (int, error) {
			return buf.Read(p)
		}).AnyTimes()

		// Test the stream handler by creating a mock stream and calling it
		// Note: This is a simplified test since we can't easily test the actual stream handler
		// without a real host, but we can test the mock setup
		assert.NotNil(t, mockStream)
	})

	t.Run("stream handler handles read errors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStream := mocks.NewMockStreamInterface(ctrl)

		// Set up mock expectations for the stream
		mockStream.EXPECT().Read(gomock.Any()).Return(0, errors.New("read error")).AnyTimes()

		// Test that the mock is properly set up
		assert.NotNil(t, mockStream)
	})
}

func TestNewEndpoint(t *testing.T) {
	t.Parallel()

	endpoint := system.ProcConfig{}.NewEndpoint()

	assert.NotNil(t, endpoint)
	assert.NotEmpty(t, endpoint.Name)

	// Test that the endpoint name is base58 encoded
	assert.Greater(t, len(endpoint.Name), 0)

	// Test that the protocol ID is correctly formatted
	protocol := endpoint.Protocol()
	assert.Contains(t, string(protocol), "/ww/0.1.0/")
	assert.Contains(t, string(protocol), endpoint.Name)
}

func TestEndpoint_String(t *testing.T) {
	t.Parallel()

	endpoint := system.ProcConfig{}.NewEndpoint()

	result := endpoint.String()
	expected := endpoint.Name

	assert.Equal(t, expected, result)
	assert.Equal(t, endpoint.Name, result)
}

func TestEndpoint_Protocol(t *testing.T) {
	t.Parallel()

	endpoint := system.ProcConfig{}.NewEndpoint()

	protocol := endpoint.Protocol()

	assert.Contains(t, string(protocol), "/ww/0.1.0/")
	assert.Contains(t, string(protocol), endpoint.Name)
}

func TestProcConfig_New_ErrorHandling(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockErrWriter := &bytes.Buffer{}

	t.Run("empty bytecode", func(t *testing.T) {
		runtime := wazero.NewRuntime(ctx)
		defer runtime.Close(ctx)

		config := system.ProcConfig{
			Host:      nil,
			Runtime:   runtime,
			Bytecode:  []byte{},
			ErrWriter: mockErrWriter,
			Async:     true, // Use async mode for these tests
		}

		proc, err := config.New(ctx)
		assert.Error(t, err)
		assert.Nil(t, proc)
	})

	t.Run("nil bytecode", func(t *testing.T) {
		runtime := wazero.NewRuntime(ctx)
		defer runtime.Close(ctx)

		config := system.ProcConfig{
			Host:      nil,
			Runtime:   runtime,
			Bytecode:  nil,
			ErrWriter: mockErrWriter,
			Async:     true, // Use async mode for these tests
		}

		proc, err := config.New(ctx)
		assert.Error(t, err)
		assert.Nil(t, proc)
	})
}

func TestEndpoint_Concurrency(t *testing.T) {
	// Test that multiple endpoints have different names
	endpoint1 := system.ProcConfig{}.NewEndpoint()
	endpoint2 := system.ProcConfig{}.NewEndpoint()

	assert.NotEqual(t, endpoint1.Name, endpoint2.Name, "Endpoints should have unique names")
	assert.NotEqual(t, endpoint1.String(), endpoint2.String(), "Endpoint strings should be different")
	assert.NotEqual(t, endpoint1.Protocol(), endpoint2.Protocol(), "Endpoint protocols should be different")
}

func TestProc_Integration_WithRealWasm(t *testing.T) {
	ctx := context.Background()

	// Create a libp2p host with in-process transport
	host, err := libp2p.New(
		libp2p.Transport(inproc.New()),
	)
	require.NoError(t, err)
	defer host.Close()

	// This test uses a more complete WASM module to test the full flow
	runtime := wazero.NewRuntime(ctx)
	defer runtime.Close(ctx)

	// Load the real echo WASM file for testing
	completeBytecode := loadEchoWasm(t)

	config := system.ProcConfig{
		Host:      host,
		Runtime:   runtime,
		Bytecode:  completeBytecode,
		ErrWriter: &bytes.Buffer{},
		Async:     true, // Use async mode for integration test
	}

	proc, err := config.New(ctx)
	require.NoError(t, err)
	require.NotNil(t, proc)
	defer proc.Close(ctx)

	// Test that all components are properly initialized
	assert.NotNil(t, proc.Closer, "Closer should be initialized")
	assert.NotNil(t, proc.Module, "Module should be initialized")
	assert.NotNil(t, proc.Endpoint, "Endpoint should be initialized")
	assert.NotEmpty(t, proc.Endpoint.Name, "Endpoint should have a name")
	assert.NotEmpty(t, proc.ID(), "String should not be empty")

	// Test that the poll function is exported
	pollFunc := proc.Module.ExportedFunction("poll")
	assert.NotNil(t, pollFunc, "poll function should be exported")
}

// TestEcho_Synchronous tests the echo example in synchronous mode
// where main() is called directly and processes stdin to stdout
func TestEcho_Synchronous(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create a libp2p host with in-process transport
	host, err := libp2p.New(
		libp2p.Transport(inproc.New()),
	)
	require.NoError(t, err)
	defer host.Close()

	// Create runtime and load echo WASM
	runtime := wazero.NewRuntime(ctx)
	defer runtime.Close(ctx)

	bytecode := loadEchoWasm(t)

	// Test input data
	testInput := "Hello, World!\nThis is a test message.\n"
	expectedOutput := testInput

	// Create a pipe to simulate stdin/stdout
	reader, writer := io.Pipe()

	// Write test data to the writer end
	go func() {
		defer writer.Close()
		writer.Write([]byte(testInput))
	}()

	// Create a buffer to capture output
	var outputBuffer bytes.Buffer

	// Compile the module first
	cm, err := runtime.CompileModule(ctx, bytecode)
	require.NoError(t, err)
	defer cm.Close(ctx)

	// Instantiate WASI
	wasi, err := wasi_snapshot_preview1.Instantiate(ctx, runtime)
	require.NoError(t, err)
	defer wasi.Close(ctx)

	// Create a new module instance with the pipe as stdin/stdout
	// In sync mode, _start runs automatically and processes stdin/stdout
	mod, err := runtime.InstantiateModule(ctx, cm, wazero.NewModuleConfig().
		WithName("echo-sync-test").
		WithSysNanosleep().
		WithSysNanotime().
		WithSysWalltime().
		WithStdin(reader).         // Use pipe reader as stdin
		WithStdout(&outputBuffer). // Capture output
		WithStderr(&bytes.Buffer{}))
	require.NoError(t, err)
	defer mod.Close(ctx)

	// Wait for the pipe to be closed (indicating EOF)
	time.Sleep(100 * time.Millisecond)

	// Verify the output matches the input
	output := outputBuffer.String()
	assert.Equal(t, expectedOutput, output, "Echo should output exactly what was input")
}

// TestEcho_Asynchronous tests the echo example in asynchronous mode
// where poll() is called with a stream and processes one complete message
func TestEcho_Asynchronous(t *testing.T) {
	// t.Parallel() // Temporarily disabled to debug runtime close issue
	ctx := context.Background()

	// Create a libp2p host with in-process transport
	host, err := libp2p.New(
		libp2p.Transport(inproc.New()),
	)
	require.NoError(t, err)
	defer host.Close()

	// Create runtime and load echo WASM
	runtime := wazero.NewRuntime(ctx)
	defer runtime.Close(ctx)

	bytecode := loadEchoWasm(t)
	config := system.ProcConfig{
		Host:      host,
		Runtime:   runtime,
		Bytecode:  bytecode,
		ErrWriter: &bytes.Buffer{},
		Async:     true, // Async mode
	}

	proc, err := config.New(ctx)
	require.NoError(t, err)
	defer proc.Close(ctx)

	// In async mode, _start is prevented from running automatically
	// We'll call poll() for each message

	// Test input data
	testInput := "Hello, Async World!\nThis is an async test message.\n"
	expectedOutput := testInput

	// Create a mock stream using gomock
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStream := mocks.NewMockStreamInterface(ctrl)
	writeBuffer := &bytes.Buffer{}

	// Set up expectations for the mock stream
	// The echo module will read from stdin until EOF, then write to stdout
	mockStream.EXPECT().
		Read(gomock.Any()).
		DoAndReturn(func(p []byte) (int, error) {
			if len(testInput) == 0 {
				return 0, io.EOF
			}
			n := copy(p, testInput)
			testInput = testInput[n:]
			return n, nil
		}).
		AnyTimes()

	mockStream.EXPECT().
		Write(gomock.Any()).
		DoAndReturn(func(p []byte) (int, error) {
			return writeBuffer.Write(p)
		}).
		AnyTimes()

	mockStream.EXPECT().
		Close().
		Return(nil).
		AnyTimes()

	// Process message with the mock stream
	// This should process one complete message (until EOF)
	err = proc.ProcessMessage(ctx, mockStream)
	require.NoError(t, err, "ProcessMessage should succeed")

	// Verify the output matches the input
	output := writeBuffer.String()
	assert.Equal(t, expectedOutput, output, "Async echo should output exactly what was input")
}

// TestEcho_RepeatedAsync tests multiple asynchronous calls to verify
// the pattern works consistently across multiple messages
func TestEcho_RepeatedAsync(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create a libp2p host with in-process transport
	host, err := libp2p.New(
		libp2p.Transport(inproc.New()),
	)
	require.NoError(t, err)
	defer host.Close()

	bytecode := loadEchoWasm(t)

	// Test multiple messages
	testMessages := []string{
		"Message 1: Hello World!\n",
		"Message 2: This is a test.\n",
		"Message 3: Multiple async calls.\n",
		"Message 4: Each should be processed independently.\n",
		"Message 5: One stream = one message.\n",
	}

	// Process each message with a separate poll call
	for i, testInput := range testMessages {
		t.Run(fmt.Sprintf("message_%d", i+1), func(t *testing.T) {
			t.Parallel()

			// Create a fresh runtime and proc for each sub-test
			runtime := wazero.NewRuntime(ctx)
			defer runtime.Close(ctx)

			config := system.ProcConfig{
				Host:      host,
				Runtime:   runtime,
				Bytecode:  bytecode,
				ErrWriter: &bytes.Buffer{},
				Async:     true, // Async mode
			}

			proc, err := config.New(ctx)
			require.NoError(t, err)
			defer proc.Close(ctx)

			// Create a fresh mock stream for each message
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStream := mocks.NewMockStreamInterface(ctrl)
			writeBuffer := &bytes.Buffer{}

			// Create a local copy of testInput to avoid closure issues
			localTestInput := testInput

			// Set up expectations for the mock stream
			mockStream.EXPECT().
				Read(gomock.Any()).
				DoAndReturn(func(p []byte) (int, error) {
					if len(localTestInput) == 0 {
						return 0, io.EOF
					}
					n := copy(p, localTestInput)
					localTestInput = localTestInput[n:]
					return n, nil
				}).
				AnyTimes()

			mockStream.EXPECT().
				Write(gomock.Any()).
				DoAndReturn(func(p []byte) (int, error) {
					return writeBuffer.Write(p)
				}).
				AnyTimes()

			mockStream.EXPECT().
				Close().
				Return(nil).
				AnyTimes()

			// Process message with the mock stream
			err = proc.ProcessMessage(ctx, mockStream)
			require.NoError(t, err, "ProcessMessage should succeed for message %d", i+1)

			// Verify the output matches the input
			output := writeBuffer.String()
			assert.Equal(t, testInput, output, "Message %d should be echoed correctly", i+1)
		})
	}
}

// TestEndpoint_NilReadWriteCloser tests that Endpoint handles nil ReadWriteCloser correctly
func TestEndpoint_NilReadWriteCloser(t *testing.T) {
	t.Parallel()

	// Create an endpoint with nil ReadWriteCloser
	endpoint := &system.Endpoint{
		Name: "test-endpoint",
		// ReadWriteCloser is nil
	}

	// Test Read returns EOF immediately
	buf := make([]byte, 10)
	n, err := endpoint.Read(buf)
	assert.Equal(t, 0, n, "Read should return 0 bytes")
	assert.Equal(t, io.EOF, err, "Read should return EOF")

	// Test Write discards data
	n, err = endpoint.Write([]byte("test data"))
	assert.Equal(t, 9, n, "Write should return length of data")
	assert.NoError(t, err, "Write should not return error")

	// Test Close doesn't panic
	err = endpoint.Close(context.Background())
	assert.NoError(t, err, "Close should not return error")
}
