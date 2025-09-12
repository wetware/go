package lang

import (
	"context"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/ipfs/boxo/path"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/spy16/slurp/core"
	"github.com/spy16/slurp/reader"
	"github.com/wetware/go/util"
)

// IPFSPathReader is a ReaderMacro that handles IPFS and IPNS paths
func IPFSPathReader(ipfs iface.CoreAPI) func(*reader.Reader, rune) (core.Any, error) {
	return func(rd *reader.Reader, init rune) (core.Any, error) {
		beginPos := rd.Position()

		// Read the full path manually by reading runes until we hit whitespace or a delimiter
		var pathBuilder strings.Builder
		pathBuilder.WriteRune(init) // Start with the '/' character

		for {
			r, err := rd.NextRune()
			if err != nil {
				if err == io.EOF {
					break
				}
				return nil, &reader.Error{
					Cause: err,
					Begin: beginPos,
					End:   beginPos,
				}
			}

			// Check if this rune should terminate the path
			// Stop on whitespace, closing parentheses, and other delimiters
			if unicode.IsSpace(r) || r == ')' || r == ']' || r == '}' {
				rd.Unread(r)
				break
			}

			pathBuilder.WriteRune(r)
		}

		pathStr := pathBuilder.String()

		// Try to create an IPFS path - path.NewPath will validate the format
		_, err := path.NewPath(pathStr)
		if err == nil {
			// Successfully created an IPFS path, return the invokable Path type
			return NewPath(context.Background(), pathStr)
		}

		// If path.NewPath failed, it's a syntax error
		return nil, &reader.Error{
			Cause: fmt.Errorf("invalid IPFS/IPNS path: %s", err),
			Begin: beginPos,
			End:   beginPos,
		}
	}
}

type Path struct {
	Path path.Path
	Env  *util.IPFSEnv
}

func NewPath(ctx context.Context, pathStr string) (core.Any, error) {
	ipfsPath, err := path.NewPath(pathStr)
	if err != nil {
		return nil, fmt.Errorf("invalid IPFS path %s: %w", pathStr, err)
	}
	return &Path{Path: ipfsPath}, nil
}

// DefaultReaderFactory creates readers with IPFS path support
type DefaultReaderFactory struct {
	IPFS iface.CoreAPI
}

func (f DefaultReaderFactory) NewReader(r io.Reader) *reader.Reader {
	rd := reader.New(r)
	rd.SetMacro('/', false, IPFSPathReader(f.IPFS))
	return rd
}
