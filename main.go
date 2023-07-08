package main

import (
	"errors"
	"io/fs"
	"log"
	"os"

	"golang.org/x/sys/unix"
)

func main() {
	log.Println("started init")

	const (
		chmod0755      = unix.S_IRWXU | unix.S_IRGRP | unix.S_IXGRP | unix.S_IROTH
		chmod0555      = unix.S_IXUSR | unix.S_IRGRP | unix.S_IXGRP | unix.S_IROTH | unix.S_IXOTH
		chmod1777      = unix.S_IRWXU | unix.S_IRWXG | unix.S_IRWXO | unix.S_ISVTX
		commonMntFlags = unix.MS_NODEV | unix.MS_NOEXEC | unix.MS_NOSUID
	)

	log.Println("mounting /dev")
	err := mount("devtmpfs", "/dev", "devtmpfs", unix.MS_NOSUID, "mode=0755")
	if err != nil {
		log.Fatal(err)
	}

	err = os.Mkdir("/newroot", chmod0755)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			log.Fatal(err)
		}
	}

	log.Println("Mounting newroot fs")
	if err := mount("/dev/vda", "/newroot", "ext4", unix.MS_RELATIME, ""); err != nil {
		log.Fatal(err)
	}

	log.Println("Moving /dev")
	if err := mount("/dev", "/newroot/dev", "", unix.MS_MOVE, ""); err != nil {
		log.Fatal(err)
	}

	log.Println("Removing /fly")
	if err := os.RemoveAll("/init"); err != nil {
		log.Fatal(err)
	}

	log.Println("Switching root")
	if err := os.Chdir("/newroot"); err != nil {
		log.Fatal(err)
	}

	if err := mount(".", "/", "", unix.MS_MOVE, ""); err != nil {
		log.Fatal(err)
	}

	if err := unix.Chroot("."); err != nil {
		log.Fatal(err)
	}

	if err := os.Chdir("/"); err != nil {
		log.Fatal(err)
	}

	log.Println("Mounting /dev/pts")
	err = os.Mkdir("/dev/pts", chmod0755)
	if err != nil {
		if !errors.Is(err, os.ErrExist) {
			log.Fatal(err)
		}
	}

	if err := mount("devpts",
		"/dev/pts",
		"devpts",
		unix.MS_NOEXEC|unix.MS_NOSUID|unix.MS_NOATIME,
		"mode=0620,gid=5,ptmxmode=666",
	); err != nil {
		log.Fatal(err)
	}

	if err := mount("devpts",
		"/dev/pts",
		"devpts",
		unix.MS_NOEXEC|unix.MS_NOSUID|unix.MS_NOATIME,
		"mode=0620,gid=5,ptmxmode=666",
	); err != nil {
		log.Fatal(err)
	}

	log.Println("Mounting /dev/mqueue")
	if err := os.Mkdir("/dev/mqueue", chmod0755); err != nil {
		if !errors.Is(err, os.ErrExist) {
			log.Fatal(err)
		}
	}

	if err := mount("mqueue",
		"/dev/mqueue",
		"mqueue",
		commonMntFlags,
		"mode=0755",
	); err != nil {
		log.Fatal(err)
	}

	log.Println("Mounting /dev/shm")
	if err := os.Mkdir("/dev/shm", chmod1777); err != nil {
		if !errors.Is(err, os.ErrExist) {
			log.Fatal(err)
		}
	}

	if err := mount("shm",
		"/dev/shm",
		"tmpfs",
		unix.MS_NOSUID|unix.MS_NODEV,
		"",
	); err != nil {
		log.Fatal(err)
	}

	log.Println("Mounting /dev/hugepages")
	if err := os.Mkdir("/dev/hugepages", chmod0755); err != nil {
		if !errors.Is(err, os.ErrExist) {
			log.Fatal(err)
		}
	}

	if err := mount("hugetlbfs",
		"/dev/hugepages",
		"hugetlbfs",
		unix.MS_RELATIME,
		"pagesize=2M",
	); err != nil {
		log.Fatal(err)
	}

	log.Println("Mounting /proc")
	if err := os.Mkdir("/proc", chmod0555); err != nil {
		if !errors.Is(err, os.ErrExist) {
			log.Fatal(err)
		}
	}

	if err := mount("proc",
		"/proc",
		"proc",
		commonMntFlags,
		"",
	); err != nil {
		log.Fatal(err)
	}

	if err := mount("binfmt_misc",
		"/proc/sys/fs/binfmt_misc",
		"binfmt_misc",
		commonMntFlags|unix.MS_RELATIME,
		"",
	); err != nil {
		log.Fatal(err)
	}

	log.Println("Mounting /sys")
	if err := os.Mkdir("/sys", chmod0555); err != nil {
		log.Fatal(err)
	}

	if err := mount("sys",
		"/sys",
		"sysfs",
		commonMntFlags,
		"",
	); err != nil {
		log.Fatal(err)
	}

	log.Println("Mounting /run")
	if err := os.Mkdir("/run", chmod0755); err != nil {
		if !errors.Is(err, os.ErrExist) {
			log.Fatal(err)
		}
	}

	if err := mount("run",
		"/run",
		"tmpfs",
		unix.MS_NOSUID|unix.MS_NODEV,
		"mode=0755",
	); err != nil {
		log.Fatal(err)
	}

	if err := os.Mkdir("/run/lock", fs.FileMode(^uint32(0))); err != nil {
		log.Fatal(err)
	}

	if err := unix.Symlinkat("/proc/self/fd", 0, "/dev/fd"); err != nil {
		log.Fatal(err)
	}

	if err := unix.Symlinkat("/proc/self/fd/0", 0, "/dev/stdin"); err != nil {
		log.Fatal(err)
	}

	if err := unix.Symlinkat("/proc/self/fd/1", 0, "/dev/stdout"); err != nil {
		log.Fatal(err)
	}

	if err := unix.Symlinkat("/proc/self/fd/2", 0, "/dev/stderr"); err != nil {
		log.Fatal(err)
	}

	if err := os.Mkdir("/root", unix.S_IRWXU); err != nil {
		if !errors.Is(err, os.ErrExist) {
			log.Fatal(err)
		}
	}

	const commonCgroupMntFlags = unix.MS_NODEV | unix.MS_NOEXEC | unix.MS_NOSUID | unix.MS_RELATIME

	log.Println("Mounting cgroup")
	if err := mount("tmpfs",
		"/sys/fs/cgroup",
		"tmpfs",
		unix.MS_NOSUID|unix.MS_NOEXEC|unix.MS_NODEV,
		"mode=0755",
	); err != nil {
		log.Fatal(err)
	}

	log.Println("Mounting cgroup2")
	if err := os.Mkdir("/sys/fs/cgroup/unified", chmod0555); err != nil {
		if !errors.Is(err, os.ErrExist) {
			log.Fatal(err)
		}
	}

	if err := mount("cgroup2",
		"/sys/fs/cgroup/unified",
		"cgroup2",
		commonCgroupMntFlags|unix.MS_RELATIME,
		"nsdelegate",
	); err != nil {
		log.Fatal(err)
	}

	log.Println("Mounting /sys/fs/cgroup/net_cls,net_prio")
	if err := os.Mkdir("/sys/fs/cgroup/net_cls", chmod0555); err != nil {
		if !errors.Is(err, os.ErrExist) {
			log.Fatal(err)
		}
	}

	if err := os.Mkdir("/sys/fs/cgroup/net_prio", chmod0555); err != nil {
		if !errors.Is(err, os.ErrExist) {
			log.Fatal(err)
		}
	}

	if err := mount("cgroup",
		"/sys/fs/cgroup/net_cls",
		"cgroup",
		commonCgroupMntFlags,
		"net_cls",
	); err != nil {
		log.Fatal(err)
	}

	if err := mount("cgroup",
		"/sys/fs/cgroup/net_prio",
		"cgroup",
		commonCgroupMntFlags,
		"net_prio",
	); err != nil {
		log.Fatal(err)
	}

	log.Println("Mounting /sys/fs/cgroup/hugetlb")
	if err := os.Mkdir("/sys/fs/cgroup/hugetlb", chmod0555); err != nil {
		if !errors.Is(err, os.ErrExist) {
			log.Fatal(err)
		}
	}

	if err := mount("cgroup",
		"/sys/fs/cgroup/hugetlb",
		"cgroup",
		commonCgroupMntFlags,
		"hugetlb",
	); err != nil {
		log.Fatal(err)
	}

	log.Println("Mounting /sys/fs/cgroup/pids")
	if err := os.Mkdir("/sys/fs/cgroup/pids", chmod0555); err != nil {
		if !errors.Is(err, os.ErrExist) {
			log.Fatal(err)
		}
	}

	if err := mount("cgroup",
		"/sys/fs/cgroup/pids",
		"cgroup",
		commonCgroupMntFlags,
		"pids",
	); err != nil {
		log.Fatal(err)
	}

	log.Println("Mounting /sys/fs/cgroup/freezer")
	if err := os.Mkdir("/sys/fs/cgroup/freezer", chmod0555); err != nil {
		if !errors.Is(err, os.ErrExist) {
			log.Fatal(err)
		}
	}

	if err := mount("cgroup",
		"/sys/fs/cgroup/freezer",
		"cgroup",
		commonCgroupMntFlags,
		"freezer",
	); err != nil {
		log.Fatal(err)
	}

	log.Println("Mounting /sys/fs/cgroup/cpu,cpuacct")
	if err := os.Mkdir("/sys/fs/cgroup/cpuacct", chmod0555); err != nil {
		if !errors.Is(err, os.ErrExist) {
			log.Fatal(err)
		}
	}

	if err := os.Mkdir("/sys/fs/cgroup/cpu", chmod0555); err != nil {
		if !errors.Is(err, os.ErrExist) {
			log.Fatal(err)
		}
	}

	if err := mount("cgroup",
		"/sys/fs/cgroup/cpu",
		"cgroup",
		commonCgroupMntFlags,
		"pids",
	); err != nil {
		log.Fatal(err)
	}

	if err := mount("cgroup",
		"/sys/fs/cgroup/cpuacct",
		"cgroup",
		commonCgroupMntFlags,
		"pids",
	); err != nil {
		log.Fatal(err)
	}

	log.Println("Mounting /sys/fs/cgroup/devices")
	if err := os.Mkdir("/sys/fs/cgroup/devices", chmod0555); err != nil {
		if !errors.Is(err, os.ErrExist) {
			log.Fatal(err)
		}
	}

	if err := mount("cgroup",
		"/sys/fs/cgroup/devices",
		"cgroup",
		commonCgroupMntFlags,
		"devices",
	); err != nil {
		log.Fatal(err)
	}

	log.Println("Mounting /sys/fs/cgroup/blkio")
	if err := os.Mkdir("/sys/fs/cgroup/blkio", chmod0555); err != nil {
		if !errors.Is(err, os.ErrExist) {
			log.Fatal(err)
		}
	}

	if err := mount("cgroup",
		"/sys/fs/cgroup/blkio",
		"cgroup",
		commonCgroupMntFlags,
		"blkio",
	); err != nil {
		log.Fatal(err)
	}

	log.Println("Mounting /sys/fs/cgroup/memory")
	if err := os.Mkdir("/sys/fs/cgroup/memory", chmod0555); err != nil {
		if !errors.Is(err, os.ErrExist) {
			log.Fatal(err)
		}
	}

	if err := mount("cgroup",
		"/sys/fs/cgroup/memory",
		"cgroup",
		commonCgroupMntFlags,
		"memory",
	); err != nil {
		log.Fatal(err)
	}

	log.Println("Mounting /sys/fs/cgroup/perf_event")
	if err := os.Mkdir("/sys/fs/cgroup/perf_event", chmod0555); err != nil {
		if !errors.Is(err, os.ErrExist) {
			log.Fatal(err)
		}
	}

	if err := mount("cgroup",
		"/sys/fs/cgroup/perf_event",
		"cgroup",
		commonCgroupMntFlags,
		"perf_event",
	); err != nil {
		log.Fatal(err)
	}

	log.Println("Mounting /sys/fs/cgroup/cpuset")
	if err := os.Mkdir("/sys/fs/cgroup/cpuset", chmod0555); err != nil {
		if !errors.Is(err, os.ErrExist) {
			log.Fatal(err)
		}
	}

	if err := mount("cgroup",
		"/sys/fs/cgroup/cpuset",
		"cgroup",
		commonCgroupMntFlags,
		"cpuset",
	); err != nil {
		log.Fatal(err)
	}

	if err := unix.Setrlimit(unix.RLIMIT_NOFILE, &unix.Rlimit{Cur: 10240, Max: 10240}); err != nil {
		log.Fatal(err)
	}

}
