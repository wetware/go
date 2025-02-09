package protoutils

import (
	"path"

	"github.com/blang/semver/v4"
	"github.com/libp2p/go-libp2p/core/protocol"
)

type VersionedID struct {
	ID      protocol.ID
	Version semver.Version
}

func (v VersionedID) Path() string {
	return path.Clean(path.Join("/", v.String()))
}

func (v VersionedID) String() string {
	proto := string(v.ID)
	version := v.Version.String()
	return path.Join(proto, version)
}

func (v VersionedID) Unwrap() protocol.ID {
	proto := v.String()
	return protocol.ID(proto)
}

func (v VersionedID) WithChild(child string) VersionedID {
	proto := path.Join(v.String())
	return VersionedID{
		ID:      protocol.ID(proto),
		Version: v.Version,
	}
}

func (v VersionedID) WithChildProto(child protocol.ID) VersionedID {
	return v.WithChild(string(child))
}

func (v VersionedID) WithVersion(version semver.Version) VersionedID {
	return VersionedID{
		ID:      v.ID,
		Version: version,
	}
}

func Join(ids ...protocol.ID) protocol.ID {
	var ss []string
	for _, id := range ids {
		ss = append(ss, string(id))
	}

	proto := path.Join(ss...)
	return protocol.ID(proto)
}
