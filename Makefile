KERNEL_OPTS = 'ro console=ttyS0,115200n8 noapic reboot=k panic=-1 pci=off nomodules init=/thi/init'

build/init:
	@echo "building init binary"
	@CGO_ENABLED=0 go build -o ./bin/init .

drive_path ?= assets/tmpinit.ext2
mount_dir ?= assets/initmount
init_dir ?= ${mount_dir}/thi
init_bin ?= bin/init
init_config ?= assets/run.json

build/init_drive:
	@echo "buildinig init drive"
	sudo init_bin=${init_bin} init_dir=${init_dir} init_config=${init_config} mount_dir=${mount_dir} drive_path=${drive_path} ./scripts/build.sh

kill/firecracker:
	sudo kill -HUP $$(pgrep 'firecracker')

run:
	sudo spitfire mkroot --fs ext4 --image 'alpine:latest' --size 100M --name ./scratch/alpine.ext4

image ?= ubuntu.ext4
run/firecracker:
	sudo ./assets/firectl --kernel ./assets/vmlinux-5.10 --cni-net fcnet --root-drive ./assets/tmpinit.ext2 --add-drive ./assets/${image}:rw --kernel-opts ${KERNEL_OPTS}

