package lang

import (
	"fmt"

	"github.com/ipfs/go-cid"
	cbornode "github.com/ipfs/go-ipld-cbor"
	"github.com/spy16/slurp/core"
	"github.com/wetware/go/system"
)

// IPLDConsCell represents a single cons cell stored in IPLD
// This is the fundamental building block of Lisp lists
type IPLDConsCell struct {
	// The car (first element) of the cons cell
	Car core.Any
	// The cdr (rest) of the cons cell - can be another cons cell or nil
	Cdr *cid.Cid
	// IPFS capability for accessing the DAG
	IPFS system.IPFS
}

// NewIPLDConsCell creates a new IPLD cons cell
func NewIPLDConsCell(ipfs system.IPFS, car core.Any, cdr *cid.Cid) (*IPLDConsCell, error) {
	// Create the IPLD node for this cons cell
	node, err := createConsCellNode(car, cdr)
	if err != nil {
		return nil, err
	}

	nodeCID := node.Cid()
	return &IPLDConsCell{
		Car:  car,
		Cdr:  &nodeCID,
		IPFS: ipfs,
	}, nil
}

// createConsCellNode creates an IPLD node representing a cons cell
func createConsCellNode(car core.Any, cdr *cid.Cid) (*cbornode.Node, error) {
	nodeData := map[string]interface{}{
		"car": car,
		"cdr": "",
	}
	if cdr != nil {
		nodeData["cdr"] = cdr.String()
	}

	return cbornode.WrapObject(nodeData, cbornode.DefaultMultihash, -1)
}

// IPLDLinkedList represents a proper IPLD-based immutable/persistent linked list
// It's built out of IPLD cons cells, following Lisp's fundamental structure
type IPLDLinkedList struct {
	// The head cons cell of the list
	head *IPLDConsCell
	// Number of elements in the list
	ElementCount int
	// IPFS capability for accessing the DAG
	IPFS system.IPFS
}

// NewIPLDLinkedList creates a new IPLD-based linked list from values
// It builds the list using cons cells, following Lisp's cons-based list construction
func NewIPLDLinkedList(ipfs system.IPFS, values ...core.Any) (*IPLDLinkedList, error) {
	if len(values) == 0 {
		return &IPLDLinkedList{IPFS: ipfs}, nil
	}

	// Build the list using cons cells, starting from the end
	// This follows the natural cons-based list construction
	var currentCell *IPLDConsCell
	for i := len(values) - 1; i >= 0; i-- {
		var cdr *cid.Cid
		if currentCell != nil {
			cdr = currentCell.Cdr
		}

		cell, err := NewIPLDConsCell(ipfs, values[i], cdr)
		if err != nil {
			return nil, err
		}
		currentCell = cell
	}

	return &IPLDLinkedList{
		head:         currentCell,
		ElementCount: len(values),
		IPFS:         ipfs,
	}, nil
}

// GetIPLDHeadCID returns the CID of the head cons cell
func (ill *IPLDLinkedList) GetIPLDHeadCID() string {
	if ill.head == nil {
		return ""
	}
	return ill.head.Cdr.String()
}

// GetIPLDElementCount returns the number of elements in the IPLD list
func (ill *IPLDLinkedList) GetIPLDElementCount() int {
	return ill.ElementCount
}

// First returns the car of the head cons cell
func (ill *IPLDLinkedList) First() (core.Any, error) {
	if ill.head == nil {
		return nil, fmt.Errorf("empty list")
	}
	return ill.head.Car, nil
}

// Next returns the rest of the list (the cdr of the head cons cell)
func (ill *IPLDLinkedList) Next() (core.Seq, error) {
	if ill.head == nil || ill.head.Cdr == nil {
		return nil, nil
	}

	// For now, return nil since we don't have a way to reconstruct the rest
	// In a full implementation, we would fetch the next cons cell from IPFS using cdr
	return nil, nil
}

// Count returns the number of elements in the list
func (ill *IPLDLinkedList) Count() (int, error) {
	return ill.ElementCount, nil
}

// SExpr implements core.SExpressable
func (ill *IPLDLinkedList) SExpr() (string, error) {
	if ill.head == nil {
		return "()", nil
	}
	return fmt.Sprintf("(%v)", ill.head.Car), nil
}

// Conj adds elements to the collection
func (ill *IPLDLinkedList) Conj(items ...core.Any) (core.Seq, error) {
	// Handle empty list case
	if ill.head == nil {
		return NewIPLDLinkedList(ill.IPFS, items...)
	}

	// Create a new list by cons-ing the current car to the new items
	newValues := append([]core.Any{ill.head.Car}, items...)
	return NewIPLDLinkedList(ill.IPFS, newValues...)
}
