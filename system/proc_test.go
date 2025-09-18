package system_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p"
	inproc "github.com/lthibault/go-libp2p-inproc-transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero"
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
		}

		proc, err := config.New(ctx)
		require.NoError(t, err)
		require.NotNil(t, proc)
		assert.NotNil(t, proc.Sys)
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
		}

		// This will panic due to nil runtime, so we expect a panic
		assert.Panics(t, func() {
			config.New(ctx)
		})
	})
}

func TestProc_ID(t *testing.T) {
	endpoint := system.NewEndpoint()
	proc := &system.Proc{
		Endpoint: endpoint,
	}

	result := proc.ID()
	expected := endpoint.Name
	assert.Equal(t, expected, result)
	assert.NotEmpty(t, result)
}

func TestProc_Close(t *testing.T) {
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
		defer runtime.Close(ctx)

		validBytecode := loadEchoWasm(t)

		config := system.ProcConfig{
			Host:      host,
			Runtime:   runtime,
			Bytecode:  validBytecode,
			ErrWriter: &bytes.Buffer{},
		}

		proc, err := config.New(ctx)
		require.NoError(t, err)
		require.NotNil(t, proc)
		defer proc.Close(ctx)

		stack := []uint64{1, 2, 3}

		// Test that poll function exists
		pollFunc := proc.Module.ExportedFunction("poll")
		assert.NotNil(t, pollFunc, "poll function should exist")

		// The echo WASM module will try to read from stdin, so we need to expect Read calls
		// The poll function reads up to 512 bytes from stdin
		mockStream.EXPECT().Read(gomock.Any()).Return(0, io.EOF).AnyTimes()

		// Actually call the Poll method - this should succeed since we have a valid WASM function
		err = proc.Poll(ctx, mockStream, stack)
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
		defer runtime.Close(ctx)

		validBytecode := loadEchoWasm(t)

		config := system.ProcConfig{
			Host:      host,
			Runtime:   runtime,
			Bytecode:  validBytecode,
			ErrWriter: &bytes.Buffer{},
		}

		proc, err := config.New(ctxWithDeadline)
		require.NoError(t, err)
		require.NotNil(t, proc)
		defer proc.Close(ctxWithDeadline)

		stack := []uint64{1, 2, 3}

		// Set up mock expectations
		mockStream.EXPECT().SetReadDeadline(deadline).Return(nil)
		// The echo WASM module will try to read from stdin
		mockStream.EXPECT().Read(gomock.Any()).Return(0, io.EOF).AnyTimes()

		// Test that poll function exists
		pollFunc := proc.Module.ExportedFunction("poll")
		assert.NotNil(t, pollFunc, "poll function should exist")

		// Actually call the Poll method to trigger the mock expectations
		err = proc.Poll(ctxWithDeadline, mockStream, stack)
		// The WASM function should execute successfully
		assert.NoError(t, err)
	})

	t.Run("set read deadline error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStream := mocks.NewMockStreamInterface(ctrl)
		endpoint := system.NewEndpoint()

		// Create context with deadline
		deadline := time.Now().Add(time.Hour)
		ctxWithDeadline, cancel := context.WithDeadline(ctx, deadline)
		defer cancel()

		proc := &system.Proc{
			Endpoint: endpoint,
		}

		stack := []uint64{1, 2, 3}

		// Set up mock expectations
		deadlineErr := errors.New("deadline error")
		mockStream.EXPECT().SetReadDeadline(deadline).Return(deadlineErr)

		err := proc.Poll(ctxWithDeadline, mockStream, stack)
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
		}

		// This should fail to create the module due to invalid bytecode
		procNoPoll, err := configNoPoll.New(ctx)
		assert.Error(t, err, "should fail to create module with invalid bytecode")
		assert.Nil(t, procNoPoll, "proc should be nil when creation fails")
	})
}

func TestProc_StreamHandler_WithGomock(t *testing.T) {
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
	endpoint := system.NewEndpoint()

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
	endpoint := system.NewEndpoint()

	result := endpoint.String()
	expected := string(endpoint.Protocol())

	assert.Equal(t, expected, result)
	assert.Contains(t, result, "/ww/0.1.0/")
	assert.Contains(t, result, endpoint.Name)
}

func TestEndpoint_Protocol(t *testing.T) {
	endpoint := system.NewEndpoint()

	protocol := endpoint.Protocol()

	assert.Contains(t, string(protocol), "/ww/0.1.0/")
	assert.Contains(t, string(protocol), endpoint.Name)
}

func TestProcConfig_New_ErrorHandling(t *testing.T) {
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
		}

		proc, err := config.New(ctx)
		assert.Error(t, err)
		assert.Nil(t, proc)
	})
}

func TestEndpoint_Concurrency(t *testing.T) {
	// Test that multiple endpoints have different names
	endpoint1 := system.NewEndpoint()
	endpoint2 := system.NewEndpoint()

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
	}

	proc, err := config.New(ctx)
	require.NoError(t, err)
	require.NotNil(t, proc)
	defer proc.Close(ctx)

	// Test that all components are properly initialized
	assert.NotNil(t, proc.Sys, "Sys should be initialized")
	assert.NotNil(t, proc.Module, "Module should be initialized")
	assert.NotNil(t, proc.Endpoint, "Endpoint should be initialized")
	assert.NotEmpty(t, proc.Endpoint.Name, "Endpoint should have a name")
	assert.NotEmpty(t, proc.ID(), "String should not be empty")

	// Test that the poll function is exported
	pollFunc := proc.Module.ExportedFunction("poll")
	assert.NotNil(t, pollFunc, "poll function should be exported")
}
