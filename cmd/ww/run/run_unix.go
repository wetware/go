//go:build unix && !darwin

package run

import "syscall"

func sysProcAttr(chroot string) *syscall.SysProcAttr {
	return nil
	// return &syscall.SysProcAttr{
	// 	// Drop privileges to “nobody:nogroup”, so that even if
	// 	// we're running as root, the child isn’t.
	// 	Credential: &syscall.Credential{Uid: 65534, Gid: 65534},

	// 	// If the parent dies, kill the child.
	// 	Pdeathsig: syscall.SIGKILL,

	// 	// Jail to the chroot directory.  Note that this SHOULD
	// 	// be combined with a mount namespace (e.g. CLONE_NEWNS).
	// 	Chroot: chroot,

	// 	// Unshare into brand‑new namespaces:
	// 	//    - PID ⇒ child is pid 1
	// 	//    - UTS ⇒ no shared hostname
	// 	//    - MOUNT ⇒ private mount table
	// 	//    - NET ⇒ no inherited network interfaces
	// 	//    - IPC ⇒ no SysV message queues/semaphores
	// 	Cloneflags: syscall.CLONE_NEWUTS |
	// 		syscall.CLONE_NEWPID |
	// 		syscall.CLONE_NEWNS |
	// 		syscall.CLONE_NEWNET |
	// 		syscall.CLONE_NEWIPC,
	// }
}
