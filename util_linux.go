package main

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

func pivotRoot(newroot, oldroot string) error {
	tmpold := filepath.Join(newroot, "/pivot_root")

	err := unix.Mount(newroot, newroot, "", unix.MS_BIND|unix.MS_REC, "")
	if err != nil {
		return err
	}

	err = mount(newroot, newroot, "", unix.MS_BIND|unix.MS_REC, "")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(tmpold, 0700); err != nil {
		return err
	}

	if err := unix.PivotRoot(newroot, tmpold); err != nil {
		return err
	}

	if err := os.Chdir("/"); err != nil {
		return err
	}

	tmpold = "/pivot_root"
	if err := unmount(tmpold, unix.MNT_DETACH); err != nil {
		return err
	}

	if err := os.RemoveAll(tmpold); err != nil {
		return err
	}

	return nil
}
