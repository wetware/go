using Go = import "/go.capnp";

@0xdd885ad027c7a7c5;

$Go.package("anchor");
$Go.import("github.com/wetware/go/anchor");

# Block interface - represents a basic block with CID and raw data
interface Block {
  # Cid returns the content identifier for this block
  cid @0 () -> (cid :Data);
  
  # RawData returns the raw data of this block
  rawData @1 () -> (data :Data);
}

# Resolver interface - provides path resolution and tree traversal
interface Resolver {
  # Resolve resolves a path through this node, stopping at any link boundary
  # and returning the object found as well as the remaining path to traverse
  resolvePath @0 (path :List(Text)) -> (node :Node, remainingPath :List(Text));
  
  # Tree lists all paths within the object under 'path', and up to the given depth.
  # To list the entire object (similar to `find .`) pass "" and -1
  tree @1 (path :Text, depth :Int32) -> (paths :List(Text));
}

# Node interface - represents an IPLD node, inherits from Block and Resolver
interface Node extends(Block, Resolver) {
  # ResolveLink is a helper function that calls resolve and asserts the
  # output is a link
  resolveLink @0 (path :List(Text)) -> (link :Link, remainingPath :List(Text));
  
  # Copy returns a deep copy of this node
  copy @1 () -> (node :Node);
  
  # Links is a helper function that returns all links within this object
  links @2 () -> (links :List(Link));
  
  # Stat returns statistics for this node
  stat @3 () -> (stat :NodeStat);
  
  # Size returns the size in bytes of the serialized object
  size @4 () -> (size :UInt64);
}

# Link represents an IPFS Merkle DAG Link between Nodes
struct Link {
  # utf string name. should be unique per object
  name @0 :Text;
  
  # cumulative size of target object
  size @1 :UInt64;
  
  # multihash of the target object
  cid @2 :Data;
}

# NodeStat is a statistics object for a Node. Mostly sizes.
struct NodeStat {
  # Hash of the node
  hash @0 :Data;
  
  # number of links in link table
  numLinks @1 :Int32;
  
  # size of the raw, encoded data
  blockSize @2 :UInt64;
  
  # size of the links segment
  linksSize @3 :UInt64;
  
  # size of the data segment
  dataSize @4 :UInt64;
  
  # cumulative size of object and its references
  cumulativeSize @5 :UInt64;
}

