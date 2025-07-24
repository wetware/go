using Go = import "/go.capnp";

@0xda965b22da734daf;

$Go.package("system");
$Go.import("github.com/wetware/go/system");

interface IPFS {
# IPFS interface for remote operations over libp2p
  add @0 (data :Data) -> (cid :Text);
  # Add data to IPFS
  
  cat @1 (cid :Text) -> (data :Data);
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
  type @3 :Text;
  links @4 :List(Link);
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