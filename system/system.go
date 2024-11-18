package system

type ReleaseFunc func()

type ExitError interface {
	error
	ExitCode() uint32
}
