using Go = import "/go.capnp";

@0xda965b22da734daf;

$Go.package("system");
$Go.import("github.com/wetware/go/system");

# Console
###

interface Console {
  println @0 (output :Text) -> (n :UInt32);
}

# Executor
###

interface Cell {
  wait @0 () -> (result :MaybeError);

  struct MaybeError {
    union {
        ok @0 :Void;
        err   :group {
            status @1 :UInt32;
            body   @2 :Data;
        }
      }
    }
}

interface Executor {
    spawn @0 (IPFS :IPFS, command :Command) -> (cell :OptionalCell);

  struct Command {
    # Command to execute
		path @0 :Text;
    args @1 :List(Text);
    env @2 :List(Text);
		dir @3 :Text;
		# stdin @4 :Data,
		# stdout @5 :Data,
		# stderr @6 :Data,
	  extraCaps @4 :List(CapDescriptor);

    struct CapDescriptor {
      name   @0 :Text;
      client @1 :Capability;
    }
  }

    struct OptionalCell {
        union {
            cell @0 :Cell;
            err     :group {
                status @1 :UInt32;
                body   @2 :Data;
            }
        }
    }
}


# IPFS
###

interface IPFS {
  # IPFS interface for remote operations over libp2p

  add @0 (data :Data) -> (cid :Text);
  # Add data to IPFS
  
  cat @1 (cid :Text) -> (body :Data);
  # Get data from IPFS by CID
  
  ls @2 (path :Text) -> (entries :List(Entry));
  # List contents of a directory or object
  
  stat @3 (cid :Text) -> (info :NodeInfo);
  # Get information about a CID
  
  pin @4 (cid :Text) -> (success :Bool);
  # Pin a CID
  
  unpin @5 (cid :Text) -> (success :Bool);
  # Unpin a CID
  
  pins @6 () -> (cids :List(Text));
  # List pinned CIDs
  
  id @7 () -> (peerInfo :PeerInfo);
  # Get peer information
  
  connect @8 (addr :Text) -> (success :Bool);
  # Connect to a peer
  
  peers @9 () -> (peerList :List(PeerInfo));
  # List connected peers

  resolveNode @10 (path :Text) -> (cid :Text, node :import "/anchor/anchor.capnp".Node);
}

struct Entry {
# Entry in a directory listing
  name @0 :Text;
  type @1 :EntryType;
  size @2 :UInt64;
  cid @3 :Text;
}

enum EntryType {
# Type of entry in directory
  file @0;
  directory @1;
  symlink @2;
}

struct NodeInfo {
# Information about an IPFS node
  cid @0 :Text;
  size @1 :UInt64;
  cumulativeSize @2 :UInt64;
  nodeType :union {
    file @3 :FileInfo;
    directory @4 :DirectoryInfo;
    symlink @5 :SymlinkInfo;
  }
}

struct FileInfo {
# Information about a file node
  # No additional fields needed for files
}

struct DirectoryInfo {
# Information about a directory node
  links @0 :List(Link);
}

struct SymlinkInfo {
# Information about a symlink node
  target @0 :Text;
}

struct Link {
# Link in an IPFS object
  name @0 :Text;
  size @1 :UInt64;
  cid @2 :Text;
}

struct PeerInfo {
# Information about a peer
  id @0 :Text;
  addresses @1 :List(Text);
  protocols @2 :List(Text);
} 