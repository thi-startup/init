KERNEL_OPTS = 'ro console=ttyS0,115200n8 noapic reboot=k panicOD=1  pci=off nomodules init=/thi/init'

build/init:
	@echo "building init drive"
	sudo spitfire mkroot --name tmpinit --fs ext2 --size 100M --init --build-from .
	mv tmpinit scratch/tmpinit

build/rootfs: build/init
	@echo "building rootfs from image"
	sudo spitfire mkroot --fs ext4 --image 'alpine:latest' --size 100M --name alpine.ext4
	mv alpine.ext4 scratch/alpine.ext4
	sudo ./scratch/firectl --cni-net fcnet --kernel ./scratch/vmlinux --root-drive ./scratch/tmpinit --add-drive ./scratch/alpine.ext4:rw --kernel-opts ${KERNEL_OPTS}

kill/firecracker:
	sudo kill -HUP $$(pgrep 'firecracker')

run:
	sudo spitfire mkroot --fs ext4 --image 'alpine:latest' --size 100M --name ./scratch/alpine.ext4


image ?= ubuntu.ext4
run/firecracker:
	sudo ./scratch/firectl --cni-net fcnet --kernel ./scratch/vmlinux --root-drive ./scratch/tmpinit --add-drive ./scratch/${image}:rw --add-drive ./scratch/add.ext4:rw --kernel-opts ${KERNEL_OPTS}

clean:
	rm alpine.ext4 tmpinit
