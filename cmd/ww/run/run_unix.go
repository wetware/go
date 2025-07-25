//go:build unix && !darwin

package run

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

const (
	NOUSER  = 65534
	NOGROUP = 65534
)

// / Provide a guest executable at the specified path.
// It will attempt to create a hard link and fall back to copying if that fails.
func provideGuestExecutable(guestDir, guestBin string) error {

	guestExecutable := filepath.Join(guestDir, guestBin)
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	// err = os.Link(executable, dst)
	// if err == nil {
	// 	fmt.Printf("Linked executable: %s:%s\n", executable, dst)
	// 	return err
	// }

	fmt.Printf("Failed to create hard link for executable: %s:%s\n", executable, guestExecutable)
	input, err := os.Open(executable)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.Create(guestExecutable)
	if err != nil {
		return err
	}
	defer output.Close()
	_, err = io.Copy(output, input)
	if err != nil {
		return err
	}
	return linkSelfExe(executable, guestDir)
}

func linkSelfExe(src, tmpDir string) error {
	selfDir := filepath.Join(tmpDir, "proc", "self")
	if err := os.MkdirAll(selfDir, os.ModeDir); err != nil {
		return err
	}
	return os.Link(src, filepath.Join(selfDir, "exe"))
}

func setGuestPermissions(dir string) error {
	permissions := "rx"
	userFlag := fmt.Sprintf("u:%d:%s", NOUSER, permissions)
	groupFlag := fmt.Sprintf("g:%d:%s", NOGROUP, permissions)
	r := exec.Command("setfacl", "-R", "-m", userFlag, dir)
	if r.Err != nil {
		return r.Err
	}
	if err := r.Run(); err != nil {
		return err
	}
	r = exec.Command("setfacl", "-R", "-m", groupFlag, dir)
	if r.Err != nil {
		return r.Err
	}
	return r.Run()
}

func sysProcAttr(chroot string) *syscall.SysProcAttr {
	// return nil
	return &syscall.SysProcAttr{
		// Drop privileges to “nobody:nogroup”, so that even if
		// we're running as root, the child isn’t.
		Credential: &syscall.Credential{Uid: NOUSER, Gid: NOGROUP},

		// If the parent dies, kill the child.
		// Pdeathsig: syscall.SIGKILL,

		// Jail to the chroot directory.  Note that this SHOULD
		// be combined with a mount namespace (e.g. CLONE_NEWNS).
		Chroot: chroot,

		// Unshare into brand‑new namespaces:
		//    - PID ⇒ child is pid 1
		//    - UTS ⇒ no shared hostname
		//    - MOUNT ⇒ private mount table
		//    - NET ⇒ no inherited network interfaces
		//    - IPC ⇒ no SysV message queues/semaphores
		// Cloneflags: syscall.CLONE_NEWUTS |
		// 	syscall.CLONE_NEWPID |
		// 	syscall.CLONE_NEWNS |
		// 	syscall.CLONE_NEWNET |
		// 	syscall.CLONE_NEWIPC,
	}
}
