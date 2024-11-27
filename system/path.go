package system

// var _ memdb.Indexer = (*PathIndexer)(nil)
// var _ memdb.MultiIndexer = (*PathIndexer)(nil)

// type PathIndexer struct{}

// // FromArgs is called to build the exact index key from a list of arguments.
// func (i PathIndexer) FromArgs(args ...interface{}) ([]byte, error) {
// 	panic("NOT IMPLEMENTED")
// }

// // FromObject extracts index values from an object. The return values
// // are the same as a SingleIndexer except there can be multiple index
// // values.
// func (i PathIndexer) FromObject(raw interface{}) (ok bool, ixs [][]byte, err error) {
// 	// switch p := raw.(type) {
// 	// case interface{ Protocol() protocol.ID }:
// 	// 	var path Path
// 	// 	proto := string(p.Protocol())
// 	// 	if path.Multiaddr, err = ma.NewMultiaddr(proto); err != nil {
// 	// 		return
// 	// 	}

// 	// 	var host peer.ID
// 	// 	if host, err = path.Peer(); err != nil {
// 	// 		return
// 	// 	}

// 	// 	var id protocol.ID
// 	// 	if id, err = path.Proto(); err != nil {
// 	// 		return
// 	// 	}

// 	// 	var version semver.Version
// 	// 	if version, err = path.Version(); err != nil {
// 	// 		return
// 	// 	}

// 	// }
// 	panic("NOT IMPLEMENTED")
// }
