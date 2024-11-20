package system

import (
	"github.com/blang/semver/v4"
	protoutils "github.com/wetware/go/util/proto"
)

const Version = "0.1.0"

var Proto = protoutils.VersionedID{
	ID:      "ww",
	Version: semver.MustParse(Version),
}

type ExitError interface {
	error
	ExitCode() uint32
}
