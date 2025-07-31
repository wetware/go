package lang

import (
	"fmt"
	"reflect"

	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/wetware/go/system"
)

// Session represents a login session with available capabilities.
// This interface abstracts the capabilities available after a successful terminal login.
// The auth.Terminal_login_Results type satisfies this interface.
type Session interface {
	// Console returns the console capability if available
	Console() system.Console
	// HasConsole returns true if the console capability is available
	HasConsole() bool
	// Ipfs returns the IPFS capability if available
	Ipfs() system.IPFS
	// HasIpfs returns true if the IPFS capability is available
	HasIpfs() bool
	// Exec returns the executor capability if available
	Exec() system.Executor
	// HasExec returns true if the executor capability is available
	HasExec() bool
}

// BuiltinInvokable wraps a builtin function to make it compatible with slurp's core.Invokable interface
type BuiltinInvokable struct {
	fn func(args ...core.Any) (core.Any, error)
}

// Invoke implements core.Invokable for BuiltinInvokable
func (b *BuiltinInvokable) Invoke(args ...core.Any) (core.Any, error) {
	return b.fn(args...)
}

// BuiltinEnvInvokable implements core.Invokable for the env function
type BuiltinEnvInvokable struct{}

func (BuiltinEnvInvokable) Invoke(args ...core.Any) (core.Any, error) {
	return BuiltinEnv(args...)
}

// BuiltinConsInvokable implements core.Invokable for the cons function
type BuiltinConsInvokable struct{}

func (BuiltinConsInvokable) Invoke(args ...core.Any) (core.Any, error) {
	return BuiltinCons(args...)
}

// BuiltinSumInvokable implements core.Invokable for the + function
type BuiltinSumInvokable struct{}

func (BuiltinSumInvokable) Invoke(args ...core.Any) (core.Any, error) {
	return BuiltinSum(args...)
}

// HelpObject is a printable object that renders the help text when printed.
type HelpObject struct{}

func (HelpObject) String() string {
	return getDefaultHelpText()
}

// DefaultEnvironment returns a map of functions that should be available by default
// in the Wetware shell environment. This includes common Lisp functions and
// Wetware-specific utilities.
func DefaultEnvironment() map[string]core.Any {
	return map[string]core.Any{
		// Basic Lisp functions
		"nil":   builtin.Nil{},
		"true":  builtin.Bool(true),
		"false": builtin.Bool(false),

		// List operations
		"cons":  &BuiltinInvokable{fn: BuiltinCons},
		"conj":  &BuiltinInvokable{fn: BuiltinConj},
		"list":  &BuiltinInvokable{fn: BuiltinList},
		"first": &BuiltinInvokable{fn: BuiltinFirst},
		"rest":  &BuiltinInvokable{fn: BuiltinRest},
		"count": &BuiltinInvokable{fn: BuiltinCount},

		// Comparison and arithmetic
		"=":  &BuiltinInvokable{fn: BuiltinEq},
		"+":  &BuiltinInvokable{fn: BuiltinSum},
		"-":  &BuiltinInvokable{fn: BuiltinSub},
		"*":  &BuiltinInvokable{fn: BuiltinMul},
		"/":  &BuiltinInvokable{fn: BuiltinDiv},
		">":  &BuiltinInvokable{fn: BuiltinGt},
		"<":  &BuiltinInvokable{fn: BuiltinLt},
		">=": &BuiltinInvokable{fn: BuiltinGte},
		"<=": &BuiltinInvokable{fn: BuiltinLte},

		// Utility functions
		"type": &BuiltinInvokable{fn: BuiltinType},
		"str":  &BuiltinInvokable{fn: BuiltinStr},
		// Note: println is not included in default environment - only added when console capability is available

		// Environment introspection
		"env":       &BuiltinInvokable{fn: BuiltinEnv},
		"namespace": &BuiltinInvokable{fn: BuiltinNamespace},
		"keys":      &BuiltinInvokable{fn: BuiltinKeys},
		"doc":       &BuiltinInvokable{fn: BuiltinDoc},
	}
}

// BuiltinCons implements the cons function for building lists
func BuiltinCons(args ...core.Any) (core.Any, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("cons requires exactly 2 arguments, got %d", len(args))
	}

	first := args[0]
	rest := args[1]

	// If rest is already a sequence, use it directly
	if seq, ok := rest.(core.Seq); ok {
		return builtin.Cons(first, seq)
	}

	// Otherwise, create a new list with just the rest element
	restSeq := builtin.NewList(rest)
	return builtin.Cons(first, restSeq)
}

// BuiltinConj implements the conj function for adding elements to collections
func BuiltinConj(args ...core.Any) (core.Any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("conj requires at least 1 argument")
	}

	coll := args[0]
	items := args[1:]

	// Use reflection to call the Conj method if it exists
	rval := reflect.ValueOf(coll)
	conjVal := rval.MethodByName("Conj")

	if !conjVal.IsZero() {
		// Call the Conj method
		args := make([]reflect.Value, len(items))
		for i, val := range items {
			args[i] = reflect.ValueOf(val)
		}

		results := conjVal.Call(args)
		if len(results) >= 2 && !results[1].IsNil() {
			return nil, results[1].Interface().(error)
		}
		return results[0].Interface(), nil
	}

	// Fallback: try to use cons for lists
	if seq, ok := coll.(core.Seq); ok {
		result := seq
		var err error
		for _, item := range items {
			result, err = builtin.Cons(item, result)
			if err != nil {
				return nil, err
			}
		}
		return result, nil
	}

	return nil, fmt.Errorf("type '%s' has no method Conj", reflect.TypeOf(coll))
}

// BuiltinList creates a new list from the given arguments
func BuiltinList(args ...core.Any) (core.Any, error) {
	// builtin.NewList() with no arguments returns an empty list, not nil
	return builtin.NewList(args...), nil
}

// BuiltinFirst returns the first element of a sequence
func BuiltinFirst(args ...core.Any) (core.Any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("first requires exactly 1 argument, got %d", len(args))
	}

	if seq, ok := args[0].(core.Seq); ok {
		return seq.First()
	}

	return nil, fmt.Errorf("first argument must be a sequence, got %T", args[0])
}

// BuiltinRest returns the rest of a sequence (everything after the first element)
func BuiltinRest(args ...core.Any) (core.Any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("rest requires exactly 1 argument, got %d", len(args))
	}

	if seq, ok := args[0].(core.Seq); ok {
		return seq.Next()
	}

	return nil, fmt.Errorf("first argument must be a sequence, got %T", args[0])
}

// BuiltinCount returns the number of elements in a sequence
func BuiltinCount(args ...core.Any) (core.Any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("count requires exactly 1 argument, got %d", len(args))
	}

	if seq, ok := args[0].(core.Seq); ok {
		count, err := seq.Count()
		if err != nil {
			return nil, err
		}
		return builtin.Int64(count), nil
	}

	// For non-sequences, return 0
	return builtin.Int64(0), nil
}

// BuiltinEq implements equality comparison
func BuiltinEq(args ...core.Any) (core.Any, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("= requires exactly 2 arguments, got %d", len(args))
	}

	result, err := core.Eq(args[0], args[1])
	if err != nil {
		return nil, err
	}
	return builtin.Bool(result), nil
}

// BuiltinSum implements addition
func BuiltinSum(args ...core.Any) (core.Any, error) {
	if len(args) == 0 {
		return builtin.Int64(0), nil
	}

	var sum int64
	for _, arg := range args {
		switch v := arg.(type) {
		case builtin.Int64:
			sum += int64(v)
		case int:
			sum += int64(v)
		case int64:
			sum += v
		default:
			return nil, fmt.Errorf("+ requires numeric arguments, got %T", arg)
		}
	}

	return builtin.Int64(sum), nil
}

// BuiltinSub implements subtraction
func BuiltinSub(args ...core.Any) (core.Any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("- requires at least 1 argument")
	}
	if len(args) == 1 {
		// Unary minus
		switch v := args[0].(type) {
		case builtin.Int64:
			return builtin.Int64(-int64(v)), nil
		case int:
			return builtin.Int64(-int64(v)), nil
		case int64:
			return builtin.Int64(-v), nil
		default:
			return nil, fmt.Errorf("- requires numeric arguments, got %T", args[0])
		}
	}

	// Binary minus
	switch v1 := args[0].(type) {
	case builtin.Int64:
		switch v2 := args[1].(type) {
		case builtin.Int64:
			return builtin.Int64(int64(v1) - int64(v2)), nil
		case int:
			return builtin.Int64(int64(v1) - int64(v2)), nil
		case int64:
			return builtin.Int64(int64(v1) - v2), nil
		default:
			return nil, fmt.Errorf("- requires numeric arguments, got %T", args[1])
		}
	default:
		return nil, fmt.Errorf("- requires numeric arguments, got %T", args[0])
	}
}

// BuiltinMul implements multiplication
func BuiltinMul(args ...core.Any) (core.Any, error) {
	if len(args) == 0 {
		return builtin.Int64(1), nil
	}

	var product int64 = 1
	for _, arg := range args {
		switch v := arg.(type) {
		case builtin.Int64:
			product *= int64(v)
		case int:
			product *= int64(v)
		case int64:
			product *= v
		default:
			return nil, fmt.Errorf("* requires numeric arguments, got %T", arg)
		}
	}

	return builtin.Int64(product), nil
}

// BuiltinDiv implements division
func BuiltinDiv(args ...core.Any) (core.Any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("/ requires at least 2 arguments")
	}

	switch v1 := args[0].(type) {
	case builtin.Int64:
		switch v2 := args[1].(type) {
		case builtin.Int64:
			if int64(v2) == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return builtin.Int64(int64(v1) / int64(v2)), nil
		case int:
			if v2 == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return builtin.Int64(int64(v1) / int64(v2)), nil
		case int64:
			if v2 == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return builtin.Int64(int64(v1) / v2), nil
		default:
			return nil, fmt.Errorf("/ requires numeric arguments, got %T", args[1])
		}
	default:
		return nil, fmt.Errorf("/ requires numeric arguments, got %T", args[0])
	}
}

// BuiltinGt implements greater than comparison
func BuiltinGt(args ...core.Any) (core.Any, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("> requires exactly 2 arguments, got %d", len(args))
	}

	switch v1 := args[0].(type) {
	case builtin.Int64:
		switch v2 := args[1].(type) {
		case builtin.Int64:
			return builtin.Bool(int64(v1) > int64(v2)), nil
		case int:
			return builtin.Bool(int64(v1) > int64(v2)), nil
		case int64:
			return builtin.Bool(int64(v1) > v2), nil
		default:
			return nil, fmt.Errorf("> requires comparable arguments, got %T", args[1])
		}
	default:
		return nil, fmt.Errorf("> requires comparable arguments, got %T", args[0])
	}
}

// BuiltinLt implements less than comparison
func BuiltinLt(args ...core.Any) (core.Any, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("< requires exactly 2 arguments, got %d", len(args))
	}

	switch v1 := args[0].(type) {
	case builtin.Int64:
		switch v2 := args[1].(type) {
		case builtin.Int64:
			return builtin.Bool(int64(v1) < int64(v2)), nil
		case int:
			return builtin.Bool(int64(v1) < int64(v2)), nil
		case int64:
			return builtin.Bool(int64(v1) < v2), nil
		default:
			return nil, fmt.Errorf("< requires comparable arguments, got %T", args[1])
		}
	default:
		return nil, fmt.Errorf("< requires comparable arguments, got %T", args[0])
	}
}

// BuiltinGte implements greater than or equal comparison
func BuiltinGte(args ...core.Any) (core.Any, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(">= requires exactly 2 arguments, got %d", len(args))
	}

	switch v1 := args[0].(type) {
	case builtin.Int64:
		switch v2 := args[1].(type) {
		case builtin.Int64:
			return builtin.Bool(int64(v1) >= int64(v2)), nil
		case int:
			return builtin.Bool(int64(v1) >= int64(v2)), nil
		case int64:
			return builtin.Bool(int64(v1) >= v2), nil
		default:
			return nil, fmt.Errorf(">= requires comparable arguments, got %T", args[1])
		}
	default:
		return nil, fmt.Errorf(">= requires comparable arguments, got %T", args[0])
	}
}

// BuiltinLte implements less than or equal comparison
func BuiltinLte(args ...core.Any) (core.Any, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("<= requires exactly 2 arguments, got %d", len(args))
	}

	switch v1 := args[0].(type) {
	case builtin.Int64:
		switch v2 := args[1].(type) {
		case builtin.Int64:
			return builtin.Bool(int64(v1) <= int64(v2)), nil
		case int:
			return builtin.Bool(int64(v1) <= int64(v2)), nil
		case int64:
			return builtin.Bool(int64(v1) <= v2), nil
		default:
			return nil, fmt.Errorf("<= requires comparable arguments, got %T", args[1])
		}
	default:
		return nil, fmt.Errorf("<= requires comparable arguments, got %T", args[0])
	}
}

// BuiltinType returns the type name of the argument
func BuiltinType(args ...core.Any) (core.Any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("type requires exactly 1 argument, got %d", len(args))
	}

	typeName := reflect.TypeOf(args[0]).String()
	return builtin.String(typeName), nil
}

// BuiltinStr converts the argument to a string
func BuiltinStr(args ...core.Any) (core.Any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("str requires exactly 1 argument, got %d", len(args))
	}

	// Handle builtin.String specially to avoid double-quoting
	if str, ok := args[0].(builtin.String); ok {
		return str, nil
	}

	return builtin.String(fmt.Sprintf("%v", args[0])), nil
}

// BuiltinPrintln is a placeholder that will be replaced with the actual console function
func BuiltinPrintln(args ...core.Any) (core.Any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("println requires exactly 1 argument, got %d", len(args))
	}

	// This is a placeholder - in practice, this will be replaced with ConsolePrintln
	return builtin.String(fmt.Sprintf("%v", args[0])), nil
}

// BuiltinEnv returns information about the current environment
func BuiltinEnv(args ...core.Any) (core.Any, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("env takes no arguments, got %d", len(args))
	}

	// Return a map with environment information
	envInfo := Map{
		builtin.Keyword("name"):    builtin.String("Wetware Shell"),
		builtin.Keyword("version"): builtin.String("0.1.0"),
		builtin.Keyword("type"):    builtin.String("lisp"),
	}

	return envInfo, nil
}

// BuiltinNamespace returns the current environment as a map-like object
func BuiltinNamespace(args ...core.Any) (core.Any, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("namespace takes no arguments, got %d", len(args))
	}

	// Get the current environment and convert it to our Map type
	env := DefaultEnvironment()
	namespace := make(Map)
	for key, value := range env {
		namespace[builtin.Keyword(key)] = value
	}

	return namespace, nil
}

// BuiltinKeys returns a list of keys from a map-like object
func BuiltinKeys(args ...core.Any) (core.Any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("keys requires exactly 1 argument (a map), got %d", len(args))
	}

	// Handle different types of map-like objects
	switch m := args[0].(type) {
	case Map:
		// Handle our custom Map type - return keywords
		keys := make([]core.Any, 0, len(m))
		for key := range m {
			keys = append(keys, key)
		}
		return builtin.NewList(keys...), nil

	case map[string]core.Any:
		// Handle the environment map - convert strings to keywords
		keys := make([]core.Any, 0, len(m))
		for key := range m {
			keys = append(keys, builtin.Keyword(key))
		}
		return builtin.NewList(keys...), nil

	default:
		return nil, fmt.Errorf("keys requires a map-like object, got %T", args[0])
	}
}

// getDefaultHelpText returns the default help text if help.ww file is not found
func getDefaultHelpText() string {
	return `Wetware Shell - A Lisp-based shell for distributed computing

Available functions:
  (cons x xs)     - Add x to the front of list xs
  (conj coll x)   - Add x to collection coll
  (list x y z)    - Create a list with elements x, y, z
  (first xs)      - Get the first element of sequence xs
  (rest xs)       - Get all but the first element of sequence xs
  (count xs)      - Get the number of elements in sequence xs
  
  (= x y)         - Test if x equals y
  (+ x y z)       - Sum of numbers
  (- x y)         - Difference of numbers
  (* x y z)       - Product of numbers
  (/ x y)         - Quotient of numbers
  (> x y)         - Test if x is greater than y
  (< x y)         - Test if x is less than y
  
  (type x)        - Get the type of x
  (str x)         - Convert x to string
  (println x)     - Print x to console
  
  (env)           - Get environment information
  (help)          - Show this help
  (help fn)       - Show help for function fn

Special forms:
  (def name value)     - Define a global variable
  (let [x 1 y 2] body) - Create local bindings
  (fn [params] body)   - Define a function
  (if test then else)  - Conditional expression
  (do expr1 expr2)     - Execute expressions in sequence
  (quote x)            - Return x without evaluation

Wetware-specific:
  (ipfs.cat path)      - Read data from IPFS path
  (ipfs.add data)      - Add data to IPFS
  (ipfs.ls path)       - List contents of IPFS path
  (ipfs.stat path)     - Get info about IPFS path
  (go path body)       - Spawn a process

Type 'quit' to exit.`
}

// BuiltinHelp provides help information
func BuiltinHelp(args ...core.Any) (core.Any, error) {
	if len(args) == 0 {
		// Load help content from help.ww file
		helpText := getDefaultHelpText()
		return builtin.String(helpText), nil
	}

	if len(args) == 1 {
		// Function-specific help
		funcName, ok := args[0].(builtin.String)
		if !ok {
			return nil, fmt.Errorf("help function name must be a string, got %T", args[0])
		}

		// Return help for specific function
		helpMap := map[string]string{
			"cons":    "(cons x xs) - Add x to the front of list xs",
			"conj":    "(conj coll x) - Add x to collection coll",
			"list":    "(list x y z) - Create a list with elements x, y, z",
			"first":   "(first xs) - Get the first element of sequence xs",
			"rest":    "(rest xs) - Get all but the first element of sequence xs",
			"count":   "(count xs) - Get the number of elements in sequence xs",
			"=":       "(= x y) - Test if x equals y",
			"+":       "(+ x y z) - Sum of numbers",
			"-":       "(- x y) - Difference of numbers",
			"*":       "(* x y z) - Product of numbers",
			"/":       "(/ x y) - Quotient of numbers",
			">":       "(> x y) - Test if x is greater than y",
			"<":       "(< x y) - Test if x is less than y",
			"type":    "(type x) - Get the type of x",
			"str":     "(str x) - Convert x to string",
			"println": "(println x) - Print x to console",
			"env":     "(env) - Get environment information",
			"help":    "(help) or (help fn) - Show help information",
		}

		if help, exists := helpMap[string(funcName)]; exists {
			return builtin.String(help), nil
		}

		return builtin.String(fmt.Sprintf("No help available for '%s'", funcName)), nil
	}

	return nil, fmt.Errorf("help takes 0 or 1 arguments, got %d", len(args))
}

// BuiltinDoc provides documentation for functions
func BuiltinDoc(args ...core.Any) (core.Any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("doc requires exactly 1 argument, got %d", len(args))
	}

	funcName, ok := args[0].(builtin.String)
	if !ok {
		return nil, fmt.Errorf("doc function name must be a string, got %T", args[0])
	}

	// Return documentation for specific function
	docMap := map[string]string{
		"cons":    "Adds an element to the front of a list. Returns a new list with the element prepended.",
		"conj":    "Adds elements to a collection. Works with lists, vectors, and other collection types.",
		"list":    "Creates a new list containing the given elements.",
		"first":   "Returns the first element of a sequence. Returns nil for empty sequences.",
		"rest":    "Returns all elements of a sequence except the first. Returns an empty sequence for single-element sequences.",
		"count":   "Returns the number of elements in a sequence.",
		"=":       "Tests equality between two values. Returns true if the values are equal, false otherwise.",
		"+":       "Adds numbers together. Returns the sum of all arguments.",
		"-":       "Subtracts numbers. With one argument, returns the negative. With two arguments, returns the difference.",
		"*":       "Multiplies numbers together. Returns the product of all arguments.",
		"/":       "Divides numbers. Returns the quotient of the first argument divided by the second.",
		">":       "Tests if the first number is greater than the second. Returns true or false.",
		"<":       "Tests if the first number is less than the second. Returns true or false.",
		"type":    "Returns the type name of the argument as a string.",
		"str":     "Converts the argument to a string representation.",
		"println": "Prints the argument to the console and returns the number of bytes written.",
		"env":     "Returns information about the current environment as a map.",
		"help":    "Provides help information. Call with no arguments for general help, or with a function name for specific help.",
		"doc":     "Provides detailed documentation for functions.",
	}

	if doc, exists := docMap[string(funcName)]; exists {
		return builtin.String(doc), nil
	}

	return builtin.String(fmt.Sprintf("No documentation available for '%s'", funcName)), nil
}

// GlobalEnvConfig holds configuration for creating an environment with system capabilities
type GlobalEnvConfig struct {
	Console  system.Console
	IPFS     system.IPFS
	Executor system.Executor
}

// New creates a new environment map with the configured capabilities
func (cfg GlobalEnvConfig) New() map[string]core.Any {
	env := DefaultEnvironment()

	// Add console capability if available
	if cfg.HasConsole() {
		env["println"] = ConsolePrintln{Console: cfg.Console}
	}

	// Add IPFS capability if available
	if cfg.HasIpfs() {
		env["ipfs"] = IPFSObject{IPFS: cfg.IPFS}
	}

	// Add executor capability if available
	if cfg.HasExec() {
		env["go"] = Go{Executor: cfg.Executor}
	}

	return env
}
func (cfg GlobalEnvConfig) HasConsole() bool {
	return cfg.Console.IsValid()
}
func (cfg GlobalEnvConfig) HasIpfs() bool {
	return cfg.IPFS.IsValid()
}
func (cfg GlobalEnvConfig) HasExec() bool {
	return cfg.Executor.IsValid()
}
