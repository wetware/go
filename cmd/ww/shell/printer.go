package shell

import (
	"fmt"
	"os"
	"strings"

	"github.com/spy16/slurp/core"
	"github.com/wetware/go/lang"
)

// ANSI color codes for enhanced output formatting
const (
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorWhite   = "\033[37m"
	colorBold    = "\033[1m"
	colorDim     = "\033[2m"
)

type printer struct{}

func (printer) Print(val interface{}) error {
	if err, ok := val.(error); ok {
		// Enhanced error formatting with colors
		fmt.Fprintf(os.Stdout, "%s%s%s\n", colorRed, err.Error(), colorReset)
		return nil
	}

	// Handle different types for rendering with enhanced formatting
	switch v := val.(type) {

	case *lang.Buffer:
		// Enhanced buffer display with hex preview
		if len(v.Mem) > 0 {
			fmt.Fprintf(os.Stdout, "%sBuffer (%d bytes):%s\n", colorBold, len(v.Mem), colorReset)
			fmt.Fprintf(os.Stdout, "%s%s%s\n", colorCyan, v.String(), colorReset)
			if len(v.Mem) <= 64 {
				fmt.Fprintf(os.Stdout, "%sHex: %s%s\n", colorDim, v.AsHex(), colorReset)
			}
		} else {
			fmt.Fprintf(os.Stdout, "%sEmpty buffer%s\n", colorYellow, colorReset)
		}

	case lang.Map:
		// Pretty print maps with indentation
		printMap(v, 0, true)

	case string:
		// Enhanced string output with syntax highlighting for IPFS paths
		if strings.HasPrefix(v, "/ipfs/") || strings.HasPrefix(v, "/ipld/") {
			fmt.Fprintf(os.Stdout, "%s%s%s\n", colorBlue, v, colorReset)
		} else {
			fmt.Fprintf(os.Stdout, "%s%s%s\n", colorGreen, v, colorReset)
		}

	case core.SExpressable:
		form, err := v.SExpr()
		if err != nil {
			return err
		}
		// Enhanced s-expression formatting
		fmt.Fprintf(os.Stdout, "%s%s%s\n", colorDim, form, colorReset)

	case core.Any:
		// For core.Any types, try to convert to string
		if str, ok := v.(string); ok {
			fmt.Fprintf(os.Stdout, "%s%s%s\n", colorGreen, str, colorReset)
		} else {
			fmt.Fprintf(os.Stdout, "%s%+v%s\n", colorYellow, v, colorReset)
		}
	default:
		fmt.Fprintf(os.Stdout, "%s%+v%s\n", colorYellow, v, colorReset)
	}
	return nil
}

// printMap recursively prints a map with proper indentation and colors
func printMap(m lang.Map, indent int, useColors bool) {
	indentStr := strings.Repeat("  ", indent)

	for key, value := range m {
		keyStr := fmt.Sprintf("%v", key)

		// Print key with color
		if useColors {
			fmt.Fprintf(os.Stdout, "%s%s%s%s: ", indentStr, colorCyan, keyStr, colorReset)
		} else {
			fmt.Fprintf(os.Stdout, "%s%s: ", indentStr, keyStr)
		}

		// Handle nested maps recursively
		if nestedMap, ok := value.(lang.Map); ok {
			fmt.Fprintf(os.Stdout, "\n")
			printMap(nestedMap, indent+1, useColors)
		} else {
			// Print value with appropriate color
			if useColors {
				fmt.Fprintf(os.Stdout, "%s%v%s\n", colorGreen, value, colorReset)
			} else {
				fmt.Fprintf(os.Stdout, "%v\n", value)
			}
		}
	}
}
