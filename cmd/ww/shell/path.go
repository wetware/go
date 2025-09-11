package shell

import (
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/ipfs/boxo/path"
	"github.com/spy16/slurp/core"
	"github.com/spy16/slurp/reader"
)

// IPFSPathReader is a ReaderMacro that handles IPFS and IPNS paths
func IPFSPathReader(rd *reader.Reader, init rune) (core.Any, error) {
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
		// Only stop on whitespace, not on forward slashes or other path characters
		if unicode.IsSpace(r) {
			rd.Unread(r)
			break
		}

		pathBuilder.WriteRune(r)
	}

	pathStr := pathBuilder.String()

	// Try to create an IPFS path - path.NewPath will validate the format
	ipfsPath, err := path.NewPath(pathStr)
	if err == nil {
		// Successfully created an IPFS path
		return Path{Path: ipfsPath}, nil
	}

	// If path.NewPath failed, it's a syntax error
	return nil, &reader.Error{
		Cause: fmt.Errorf("invalid IPFS/IPNS path: %s", err),
		Begin: beginPos,
		End:   beginPos,
	}
}

// DefaultReaderFactory creates readers with IPFS path support
type DefaultReaderFactory struct{}

func (f DefaultReaderFactory) NewReader(r io.Reader) *reader.Reader {
	rd := reader.New(r)
	rd.SetMacro('/', false, IPFSPathReader)
	return rd
}
