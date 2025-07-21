//go:build darwin

package run

import "syscall"

func sysProcAttr(chroot string) *syscall.SysProcAttr {
	return nil
	// // Note: on macOS, SysProcAttr only provides chroot, credential drop,
	// // session/PGID control and optional ptrace-based parent–death.  It
	// // does NOT support namespace unsharing (PID, mount, network, IPC) or
	// // a pdeath signal as on Linux, so isolation here is inherently weaker
	// // unless you layer on Apple’s sandbox or run the process in a VM/container.

	// return &syscall.SysProcAttr{
	// 	// Drop privileges to “nobody:nogroup”, so that even if
	// 	// we're running as root, the child isn’t.
	// 	Credential: &syscall.Credential{Uid: 65534, Gid: 65534},

	// 	// Jail to the chroot directory.  Note that this SHOULD
	// 	// be combined with a mount namespace (e.g. CLONE_NEWNS).
	// 	Chroot: chroot,

	// 	Setsid:  true, // new session
	// 	Setpgid: true, // new process group
	// 	Pgid:    0,    // child is its own group leader
	// 	Noctty:  true, // detach from any controlling TTY
	// }
}
