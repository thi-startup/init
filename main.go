package main

import (
	"log"
	"os"

	"golang.org/x/sys/unix"
)

func main() {
	log.Println("started init")

	const (
		chmod0755      = unix.S_IRWXU | unix.S_IRGRP | unix.S_IXGRP | unix.S_IROTH
		chmod0555      = unix.S_IXUSR | unix.S_IRGRP | unix.S_IXGRP | unix.S_IROTH | unix.S_IXOTH
		commonMntFlags = unix.MS_NODEV | unix.MS_NOEXEC | unix.MS_NOSUID
	)

	log.Println("mounting /dev")
	err := mount("devtmpfs", "/dev", "devtmpfs", unix.MS_NOSUID, "mode=0755")
	if err != nil {
		log.Fatal(err)
	}

	err = os.Mkdir("/newroot", chmod0755)
	if err != nil {
		log.Fatal(err)
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
		"mode=0755",
	); err != nil {
		log.Fatal(err)
	}

}
