package main

import (
	"fmt"
	"log"

	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/wetware/go/lang"
	"github.com/wetware/go/system"
)

func main() {
	// Create a mock IPFS capability (in a real application, this would be a real IPFS node)
	ipfs := system.IPFS{}

	// Create some test values
	values := []core.Any{
		builtin.String("first"),
		builtin.String("second"),
		builtin.String("third"),
		builtin.String("fourth"),
	}

	// Create the IPLD linked list
	list, err := lang.NewIPLDLinkedList(ipfs, values...)
	if err != nil {
		log.Fatalf("Failed to create IPLD linked list: %v", err)
	}

	// Print information about the list
	fmt.Printf("IPLD Linked List created successfully!\n")
	fmt.Printf("Number of elements: %d\n", list.GetIPLDElementCount())
	fmt.Printf("Head CID: %s\n", list.GetIPLDHeadCID())

	// Demonstrate the linked list structure
	fmt.Printf("\nLinked List Structure:\n")
	fmt.Printf("Head -> [%s] -> [%s] -> [%s] -> [%s] -> nil\n",
		values[0], values[1], values[2], values[3])

	// Show that it's compatible with the builtin linked list
	count, err := list.Count()
	if err != nil {
		log.Fatalf("Failed to get count: %v", err)
	}
	fmt.Printf("\nBuiltin list compatibility - Count: %d\n", count)

	first, err := list.First()
	if err != nil {
		log.Fatalf("Failed to get first element: %v", err)
	}
	fmt.Printf("First element: %v\n", first)

	fmt.Printf("\nEach node in the IPLD DAG contains:\n")
	fmt.Printf("- value: the actual data\n")
	fmt.Printf("- next: CID pointer to the next node\n")
	fmt.Printf("- The last node has an empty 'next' field\n")
}
