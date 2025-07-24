package lang

import (
	"testing"

	"capnproto.org/go/capnp/v3"
	"github.com/spy16/slurp/core"
)

// TestInvokableCreation tests creating an Invokable instance
func TestInvokableCreation(t *testing.T) {
	// Create a real capnp.Client (null client for testing)
	client := capnp.Client{}
	invokable := Invokable[capnp.Client]{Client: client}

	if invokable.Client != client {
		t.Errorf("Expected client to be the capnp client, got %v", invokable.Client)
	}
}

// TestInvokeWithNoArgs tests invoking with no arguments (should return the invokable itself)
func TestInvokeWithNoArgs(t *testing.T) {
	client := capnp.Client{}
	invokable := Invokable[capnp.Client]{Client: client}

	result, err := invokable.Invoke()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result != invokable {
		t.Errorf("Expected result to be the invokable itself, got %v", result)
	}
}

// TestInvokeWithInvalidFirstArg tests when first argument is not a string
func TestInvokeWithInvalidFirstArg(t *testing.T) {
	client := capnp.Client{}
	invokable := Invokable[capnp.Client]{Client: client}

	result, err := invokable.Invoke(123, "arg")
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result when error occurs, got %v", result)
	}

	expectedErr := "first argument must be method name string, got int"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

// TestInvokeWithNonExistentMethod tests invoking a non-existent method
func TestInvokeWithNonExistentMethod(t *testing.T) {
	client := capnp.Client{}
	invokable := Invokable[capnp.Client]{Client: client}

	result, err := invokable.Invoke("NonExistentMethod")
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result when error occurs, got %v", result)
	}

	expectedErr := "method 'NonExistentMethod' not found"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

// TestInvokeWithTooManyArgs tests when too many arguments are provided
func TestInvokeWithTooManyArgs(t *testing.T) {
	client := capnp.Client{}
	invokable := Invokable[capnp.Client]{Client: client}

	// Try to call IsValid with extra args (IsValid takes no arguments)
	result, err := invokable.Invoke("IsValid", "extra", "args")
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result when error occurs, got %v", result)
	}

	expectedErr := "too many arguments for method 'IsValid'"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

// TestInvokableImplementsCoreInvokable tests that Invokable implements core.Invokable
func TestInvokableImplementsCoreInvokable(t *testing.T) {
	client := capnp.Client{}
	invokable := Invokable[capnp.Client]{Client: client}

	// This will cause a compile error if Invokable doesn't implement core.Invokable
	var _ core.Invokable = invokable
}

// TestInvokeWithValidMethod tests invoking a valid method on capnp.Client
func TestInvokeWithValidMethod(t *testing.T) {
	client := capnp.Client{}
	invokable := Invokable[capnp.Client]{Client: client}

	result, err := invokable.Invoke("IsValid")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// IsValid should return false for a null client
	expected := false
	if result != expected {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// TestInvokeWithStringMethod tests invoking String method
func TestInvokeWithStringMethod(t *testing.T) {
	client := capnp.Client{}
	invokable := Invokable[capnp.Client]{Client: client}

	result, err := invokable.Invoke("String")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// String should return a string representation
	if result == nil {
		t.Error("Expected non-nil result from String method")
	}

	_, ok := result.(string)
	if !ok {
		t.Errorf("Expected string result, got %T", result)
	}
}

// BenchmarkInvoke benchmarks the Invoke method
func BenchmarkInvoke(b *testing.B) {
	client := capnp.Client{}
	invokable := Invokable[capnp.Client]{Client: client}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := invokable.Invoke("IsValid")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkInvokeWithNoArgs benchmarks the Invoke method with no args
func BenchmarkInvokeWithNoArgs(b *testing.B) {
	client := capnp.Client{}
	invokable := Invokable[capnp.Client]{Client: client}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := invokable.Invoke()
		if err != nil {
			b.Fatal(err)
		}
	}
}
