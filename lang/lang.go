package lang

import (
	"fmt"
	"reflect"

	"capnproto.org/go/capnp/v3"
	"github.com/spy16/slurp/core"
)

type Invokable[T ~capnp.ClientKind] struct {
	Client T
}

// Invoke implements core.Invokable interface
func (v Invokable[T]) Invoke(args ...core.Any) (core.Any, error) {
	// If no arguments are provided, return the Value itself.
	if len(args) == 0 {
		return v, nil
	}

	// The first argument must be a string containing the method name to invoke.
	// This follows the convention of method calls in dynamic languages where
	// the method name is the first argument.
	methodName, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("first argument must be method name string, got %T", args[0])
	}

	// Use reflection to look up the named method on the underlying Cap'n Proto client.
	// The client implements the Cap'n Proto interface for the capability.
	val := reflect.ValueOf(v.Client)
	method := val.MethodByName(methodName)
	if !method.IsValid() {
		return nil, fmt.Errorf("method '%s' not found", methodName)
	}

	// Convert the remaining arguments to reflect.Values that match the method parameters.
	// This involves:
	// 1. Creating a slice to hold the converted arguments
	// 2. Getting the method's type information
	// 3. Converting each argument to the expected type
	methodArgs := make([]reflect.Value, len(args)-1)
	methodType := method.Type()
	for i := 0; i < len(args)-1; i++ {
		// Check if we have too many arguments
		if i >= methodType.NumIn() {
			return nil, fmt.Errorf("too many arguments for method '%s'", methodName)
		}

		// Convert the argument to the expected type using reflection
		arg := reflect.ValueOf(args[i+1])
		if !arg.Type().ConvertibleTo(methodType.In(i)) {
			return nil, fmt.Errorf("argument %d: cannot convert %v to %v", i+1, arg.Type(), methodType.In(i))
		}
		methodArgs[i] = arg.Convert(methodType.In(i))
	}

	// Call the method with the converted arguments.
	// This invokes the actual Cap'n Proto RPC call.
	results := method.Call(methodArgs)
	if len(results) == 0 {
		return nil, nil
	}

	// Handle any error return value.
	// By convention, if a method returns multiple values and the last one
	// is an error interface, we check it and return it if non-nil.
	if len(results) > 1 && !results[len(results)-1].IsNil() {
		if err, ok := results[len(results)-1].Interface().(error); ok {
			return nil, err
		}
	}

	// Return the first result value.
	// This is typically the actual return value of the method.
	return results[0].Interface(), nil
}
