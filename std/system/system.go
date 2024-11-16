package system

const (
	StatusAsync = 0x00ff0000 + iota
	StatusInvalidArgs
	StatusFailed
)

// type StatusCode uint

// func (status StatusCode) Error() string {
// 	switch status {
// 	case StatusAsync:
// 		return "awaiting method calls"
// 	case StatusInvalidArgs:
// 		return "invalid arguments"
// 	case StatusFailed:
// 		return "application failed"
// 	}

// 	return status.Unwrap().Error()
// }

// func (status StatusCode) Unwrap() error {
// 	switch status.ExitCode() {
// 	case sys.ExitCodeContextCanceled:
// 		return context.Canceled
// 	case sys.ExitCodeDeadlineExceeded:
// 		return context.DeadlineExceeded
// 	}

// 	return sys.NewExitError(status.ExitCode())
// }

// func (status StatusCode) ExitCode() uint32 {
// 	return uint32(status)
// }
