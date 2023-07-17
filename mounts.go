package main

import (
	"fmt"
	"io/fs"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

const (
	perm0755             = unix.S_IRWXU | unix.S_IRGRP | unix.S_IXGRP | unix.S_IROTH
	perm0555             = unix.S_IXUSR | unix.S_IRGRP | unix.S_IXGRP | unix.S_IROTH | unix.S_IXOTH
	perm1777             = unix.S_IRWXU | unix.S_IRWXG | unix.S_IRWXO | unix.S_ISVTX
	commonMntFlags       = unix.MS_NODEV | unix.MS_NOEXEC | unix.MS_NOSUID
	commonCgroupMntFlags = unix.MS_NODEV | unix.MS_NOEXEC | unix.MS_NOSUID | unix.MS_RELATIME
)

type mountInfo struct {
	source       string
	target       string
	fstype       string
	flags        uintptr
	data         string
	createTarget bool
	perm         uint32
}

func (e mountInfo) Info() string {
	out := "mount "

	if e.source != "" {
		out += e.source + ":" + e.target
	} else {
		out += e.target
	}

	if e.flags != uintptr(0) {
		out += ", flags: 0x" + strconv.FormatUint(uint64(e.flags), 16)
	}
	if e.data != "" {
		out += ", data: " + e.data
	}

	return out
}

type mounts []mountInfo

func (mnts mounts) Mount() error {
	for _, m := range mnts {
		log.Debugf(m.Info())
		if m.createTarget {
			err := os.Mkdir(m.target, fs.FileMode(m.perm))
			if err != nil && !os.IsExist(err) {
				return err
			}
		}

		if err := mount(m.source, m.target, m.fstype, m.flags, m.data); err != nil {
			return err
		}
	}
	return nil
}

func MakeMounts() mounts {
	mnts := []mountInfo{}

	mnts = append(mnts, mountInfo{
		source:       "devpts",
		target:       "/dev/pts",
		fstype:       "devpts",
		flags:        unix.MS_NOEXEC | unix.MS_NOSUID | unix.MS_NOATIME,
		data:         "mode=0755",
		createTarget: true,
		perm:         perm0755,
	})

	mnts = append(mnts, mountInfo{
		source:       "devtmpfs",
		target:       "/dev/mqueue",
		fstype:       "devtmpfs",
		flags:        commonMntFlags,
		data:         "",
		createTarget: true,
		perm:         perm0755,
	})

	mnts = append(mnts, mountInfo{
		source:       "shm",
		target:       "/dev/shm",
		fstype:       "tmpfs",
		flags:        commonMntFlags,
		data:         "",
		createTarget: true,
		perm:         perm1777,
	})

	mnts = append(mnts, mountInfo{
		source:       "hugetlbfs",
		target:       "/dev/hugepages",
		fstype:       "hugetlbfs",
		flags:        unix.MS_RELATIME,
		data:         "pagesize=2M",
		createTarget: true,
		perm:         perm0755,
	})

	mnts = append(mnts, mountInfo{
		source:       "proc",
		target:       "/proc",
		fstype:       "proc",
		flags:        commonMntFlags,
		data:         "",
		createTarget: true,
		perm:         perm0555,
	})

	mnts = append(mnts, mountInfo{
		source: "binfmt_misc",
		target: "/proc/sys/fs/binfmt_misc",
		fstype: "binfmt_misc",
		flags:  commonMntFlags | unix.MS_RELATIME,
		data:   "",
	})

	mnts = append(mnts, mountInfo{
		source:       "sys",
		target:       "/sys",
		fstype:       "sysfs",
		flags:        commonMntFlags,
		data:         "",
		createTarget: true,
		perm:         perm0555,
	})

	mnts = append(mnts, mountInfo{
		source:       "run",
		target:       "/run",
		fstype:       "tmpfs",
		flags:        unix.MS_NOSUID | unix.MS_NODEV,
		data:         "mode=0755",
		createTarget: true,
		perm:         perm0755,
	})

	return mnts
}

func MakeCgroupMounts() mounts {
	mnts := []mountInfo{}

	mnts = append(mnts, mountInfo{
		source: "tmpfs",
		target: "/sys/fs/cgroup",
		fstype: "tmpfs",
		flags:  unix.MS_NOSUID | unix.MS_NOEXEC | unix.MS_NODEV,
		data:   "mode=0755",
	})

	mnts = append(mnts, mountInfo{
		source:       "cgroup2",
		target:       "/sys/fs/cgroup/unified",
		fstype:       "cgroup2",
		flags:        commonCgroupMntFlags | unix.MS_RELATIME,
		data:         "nsdelegate",
		createTarget: true,
		perm:         perm0555,
	})

	mnts = append(mnts, mountInfo{
		source:       "cgroup",
		target:       "/sys/fs/cgroup/net_cls",
		fstype:       "cgroup",
		flags:        commonCgroupMntFlags,
		data:         "net_cls",
		createTarget: true,
		perm:         perm0555,
	})

	mnts = append(mnts, mountInfo{
		source:       "cgroup",
		target:       "/sys/fs/cgroup/net_prio",
		fstype:       "cgroup",
		flags:        commonCgroupMntFlags,
		data:         "net_prio",
		createTarget: true,
		perm:         perm0555,
	})

	mnts = append(mnts, mountInfo{
		source:       "cgroup",
		target:       "/sys/fs/cgroup/hugetlb",
		fstype:       "cgroup",
		flags:        commonCgroupMntFlags,
		data:         "hugetlb",
		createTarget: true,
		perm:         perm0555,
	})

	mnts = append(mnts, mountInfo{
		source:       "cgroup",
		target:       "/sys/fs/cgroup/pids",
		fstype:       "cgroup",
		flags:        commonCgroupMntFlags,
		data:         "pids",
		createTarget: true,
		perm:         perm0555,
	})

	mnts = append(mnts, mountInfo{
		source:       "cgroup",
		target:       "/sys/fs/cgroup/freezer",
		fstype:       "cgroup",
		flags:        commonCgroupMntFlags,
		data:         "freezer",
		createTarget: true,
		perm:         perm0555,
	})

	mnts = append(mnts, mountInfo{
		source:       "cgroup",
		target:       "/sys/fs/cgroup/cpu",
		fstype:       "cgroup",
		flags:        commonCgroupMntFlags,
		data:         "cpu",
		createTarget: true,
		perm:         perm0555,
	})

	mnts = append(mnts, mountInfo{
		source:       "cgroup",
		target:       "/sys/fs/cgroup/cpuacct",
		fstype:       "cgroup",
		flags:        commonCgroupMntFlags,
		data:         "cpuacct",
		createTarget: true,
		perm:         perm0555,
	})

	mnts = append(mnts, mountInfo{
		source:       "cgroup",
		target:       "/sys/fs/cgroup/devices",
		fstype:       "cgroup",
		flags:        commonCgroupMntFlags,
		data:         "devices",
		createTarget: true,
		perm:         perm0555,
	})

	mnts = append(mnts, mountInfo{
		source:       "cgroup",
		target:       "/sys/fs/cgroup/blkio",
		fstype:       "cgroup",
		flags:        commonCgroupMntFlags,
		data:         "blkio",
		createTarget: true,
		perm:         perm0555,
	})

	mnts = append(mnts, mountInfo{
		source:       "cgroup",
		target:       "/sys/fs/cgroup/memory",
		fstype:       "cgroup",
		flags:        commonCgroupMntFlags,
		data:         "memory",
		createTarget: true,
		perm:         perm0555,
	})

	mnts = append(mnts, mountInfo{
		source:       "cgroup",
		target:       "/sys/fs/cgroup/perf_event",
		fstype:       "cgroup",
		flags:        commonCgroupMntFlags,
		data:         "perf_event",
		createTarget: true,
		perm:         perm0555,
	})

	mnts = append(mnts, mountInfo{
		source:       "cgroup",
		target:       "/sys/fs/cgroup/cpuset",
		fstype:       "cgroup",
		flags:        commonCgroupMntFlags,
		data:         "cpuset",
		createTarget: true,
		perm:         perm0555,
	})

	return mnts
}

func MountAdditionalDrives(devices []Mounts, uid, gid int) error {
	for _, m := range devices {
		log.Infof("mounting %s at %s", m.DevicePath, m.MountPath)

		if err := os.Mkdir(m.MountPath, perm0755); err != nil {
			if os.IsExist(err) {
				log.Warnf("directory %s already exists", m.MountPath)
			} else {
				return fmt.Errorf("could not create directory %s", m.MountPath)
			}
		}

		if err := mount(m.DevicePath, m.MountPath, "ext4", unix.MS_RELATIME, ""); err != nil {
			return fmt.Errorf("error mounting drive: %v", err)
		}

		if err := unix.Chown(m.MountPath, uid, gid); err != nil {
			return fmt.Errorf("error setting permissions: %v", err)
		}
	}
	return nil
}
