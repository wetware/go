# IPLD Linked List Implementation

This document describes the IPLD (InterPlanetary Linked Data) linked list implementation in the Wetware Go project.

## Overview

The `IPLDLinkedList` provides a persistent, immutable linked list structure that uses IPLD DAG (Directed Acyclic Graph) for storage. Each node in the list is stored as a separate IPLD node with a pointer to the next node.

## Structure

### IPLDLinkedList

```go
type IPLDLinkedList struct {
    *builtin.LinkedList
    HeadCID      string      // CID of the head node
    ElementCount int         // Number of elements in the list
    IPFS         system.IPFS // IPFS capability for accessing the DAG
}
```

### IPLD Node Structure

Each node in the IPLD DAG contains:
- `value`: The actual data stored in the node
- `next`: CID pointer to the next node (empty string for the last node)

## Usage

### Creating a New IPLD Linked List

```go
// Create an IPFS capability
ipfs := system.IPFS{}

// Create values
values := []core.Any{
    builtin.String("first"),
    builtin.String("second"),
    builtin.String("third"),
}

// Create the IPLD linked list
list, err := lang.NewIPLDLinkedList(ipfs, values...)
if err != nil {
    log.Fatal(err)
}
```

### Accessing List Properties

```go
// Get the number of elements
count := list.GetIPLDElementCount()

// Get the head CID
headCID := list.GetIPLDHeadCID()

// Use builtin linked list methods
first, err := list.First()
count, err := list.Count()
```

## Implementation Details

### Node Creation

1. **Reverse Order Construction**: Nodes are created in reverse order to establish proper links
2. **CBOR Serialization**: Each node is serialized using CBOR format for IPLD compatibility
3. **CID Generation**: Each node gets a unique CID (Content Identifier)

### Linked List Structure

```
Head -> [value1, next->node2] -> [value2, next->node3] -> [value3, next->node4] -> [value4, next->nil] -> nil
```

### Example DAG Structure

```
Node 1 (Head): {value: "first", next: "bafy2bzaceb..."}
Node 2:        {value: "second", next: "bafy2bzaceb..."}
Node 3:        {value: "third", next: "bafy2bzaceb..."}
Node 4:        {value: "fourth", next: ""}
```

## Benefits

1. **Persistence**: Data is stored in IPFS/IPLD for long-term persistence
2. **Immutability**: Once created, the list structure cannot be modified
3. **Content Addressing**: Each node is identified by its content hash (CID)
4. **Compatibility**: Maintains compatibility with the builtin linked list interface
5. **Decentralized**: Can be shared and accessed across the IPFS network

## Future Enhancements

1. **IPFS Integration**: Currently generates CIDs without storing to IPFS
2. **Node Retrieval**: Add methods to retrieve nodes by CID
3. **List Operations**: Add append, prepend, and other list operations
4. **Compression**: Optimize storage by compressing similar data
5. **Caching**: Add caching for frequently accessed nodes

## Example

See `examples/ipld_list/main.go` for a complete working example.

## Testing

Run the tests with:

```bash
go test ./lang -v -run TestNewIPLDLinkedList
```

## Dependencies

- `github.com/ipfs/go-ipld-cbor`: For CBOR serialization
- `github.com/spy16/slurp/core`: For core types
- `github.com/spy16/slurp/builtin`: For builtin linked list compatibility
- `github.com/wetware/go/system`: For IPFS capabilities 