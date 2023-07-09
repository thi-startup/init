package main

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/opencontainers/runc/libcontainer/system"
	"github.com/opencontainers/runc/libcontainer/user"
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

	log.Println("decoding run.json file")
	config, err := DecodeMachine("/thi/run.json")
	if err != nil {
		log.Fatalf("could not parse run.json file %v", err)
	}

	if err := os.Mkdir("/dev", chmod0755); err != nil {
		log.Fatal(err)
	}

	log.Println("mounting /dev")
	err = mount("devtmpfs", "/dev", "devtmpfs", unix.MS_NOSUID, "mode=0755")
	if err != nil {
		log.Fatal(err)
	}

	err = os.Mkdir("/newroot", chmod0755)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Mounting newroot fs")
	if err := mount("/dev/vdb", "/newroot", "ext4", unix.MS_RELATIME, ""); err != nil {
		log.Fatal(err)
	}

	log.Println("Moving /dev")
	if err := mount("/dev", "/newroot/dev", "", unix.MS_MOVE, ""); err != nil {
		log.Fatal(err)
	}

	log.Println("Removing /thi to save space")
	if err := os.RemoveAll("/thi"); err != nil {
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
		log.Fatal(err)
	}

	if err := mount("mqueue",
		"/dev/mqueue",
		"mqueue",
		commonMntFlags,
		"",
	); err != nil {
		log.Fatal(err)
	}

	log.Println("Mounting /dev/shm")
	if err := os.Mkdir("/dev/shm", chmod1777); err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
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
		if !errors.Is(err, os.ErrExist) {
			log.Fatal(err)
		}
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
		if errors.Is(err, os.ErrExist) {
			log.Fatal(err)
		}
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

	if err := unix.Setrlimit(0, &unix.Rlimit{Cur: 10240, Max: 10240}); err != nil {
		log.Fatal(err)
	}

	// parse user and  group names
	username := config.ImageConfig.User
	if username == "" {
		username = "root"
	}
	usrSplit := strings.Split(username, ":")

	var group user.Group
	if len(usrSplit) < 1 {
		log.Fatal("no username set, something is terribly wrong!")
	} else if len(usrSplit) >= 2 {
		group, err = user.LookupGroup(usrSplit[1])
		if err != nil {
			log.Fatalf("group %s not found: %v", usrSplit[1], err)
		}
	}
	_ = group

	nixUser, err := user.LookupUser(usrSplit[0])
	if err != nil {
		log.Fatalf("user %s not found: %v", username, err)
	}

	if err := system.Setgid(nixUser.Gid); err != nil {
		log.Fatalf("unable to set group id: %v", err)
	}

	if err := system.Setuid(nixUser.Uid); err != nil {
		log.Fatalf("unable to set group id: %v", err)
	}

	// set environment variables
	for _, pair := range config.ImageConfig.Env {
		p := strings.SplitN(pair, "=", 2)
		if len(p) < 2 {
			log.Fatal("invalid env var: missing '='")
		}
		name, val := p[0], p[1]
		if name == "" {
			log.Fatal("invalid env var: name cannot be empty")
		}
		if strings.IndexByte(name, 0) >= 0 {
			log.Fatal("invalid env var: name contains null byte")
		}
		if strings.IndexByte(val, 0) >= 0 {
			log.Fatal("invalid env var: value contains null byte")
		}
		if err := os.Setenv(name, val); err != nil {
			log.Fatalf("could not set env var: system shit: %v", err)
		}
	}

	// set the home dir if not already set
	if envHome := os.Getenv("HOME"); envHome == "" {
		if err := os.Setenv("HOME", nixUser.Home); err != nil {
			log.Fatal("unable to set user home directory")
		}
	}

	if err := unix.Sethostname([]byte(config.Hostname)); err != nil {
		log.Fatalf("error setting hostname: %v", err)
	}

	if err := os.Mkdir("/etc", chmod0755); err != nil {
		if !errors.Is(err, os.ErrExist) {
			log.Fatalf("could not create /etc dir: %v", err)
		}
	}

	if err := os.WriteFile("/etc/hostname", []byte(config.Hostname), chmod0755); err != nil {
		log.Fatalf("error writing /etc/hostname: %v", err)
	}

	if len(config.ImageConfig.Cmd) < 1 {
		log.Fatal("no command to execute, exiting now!")
	}

	stdin, err := os.OpenFile("/proc/1/fd/0", os.O_RDWR, 0755)
	if err != nil {
		log.Fatal(err)
	}

	stdout, err := os.OpenFile("/proc/1/fd/1", os.O_RDWR, 0755)
	if err != nil {
		log.Fatal(err)
	}

	stderr, err := os.OpenFile("/proc/1/fd/2", os.O_RDWR, 0755)
	if err != nil {
		log.Fatal(err)
	}

	if err := unix.Fchown(int(stdin.Fd()), nixUser.Uid, nixUser.Gid); err != nil {
		log.Fatalf("could not fchown stdin file: %v", err)
	}

	if err := unix.Fchown(int(stdout.Fd()), nixUser.Uid, nixUser.Gid); err != nil {
		log.Fatalf("could not fchown stdin file: %v", err)
	}

	if err := unix.Fchown(int(stderr.Fd()), nixUser.Uid, nixUser.Gid); err != nil {
		log.Fatalf("could not fchown stdin file: %v", err)
	}

	args := func(cmd []string) []string {
		if len(cmd) == 1 {
			return nil
		}
		return cmd[1:]
	}(config.ImageConfig.Cmd)

	procAttr := &syscall.ProcAttr{
		Dir:   config.ImageConfig.WorkingDir,
		Env:   config.ImageConfig.Env,
		Files: []uintptr{stdin.Fd(), stdout.Fd(), stderr.Fd()},
		Sys: &syscall.SysProcAttr{
			Setpgid: true,
			Pgid:    nixUser.Gid,
		},
	}

	pid, err := syscall.ForkExec(config.ImageConfig.Cmd[0], args, procAttr)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("started process %d\n", pid)

	for {
	}

}
