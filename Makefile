build/init:
	@echo "building init drive"
	sudo spitfire mkroot --name tmpinit --fs ext2 --size 100M --init --build-from .
	cp tmpinit scratch/tmpinit

kill/firecracker:
	sudo kill -HUP $$(pgrep 'firecracker')

image ?= ubuntu.ext4
run/firecracker:
	sudo ./scratch/firectl --kernel ./scratch/vmlinux --root-drive ./scratch/tmpinit --add-drive ./scratch/${image}:rw --add-drive ./scratch/add.ext4:rw --kernel-opts "ro console=ttyS0,115200n8 noapic reboot=k panicOD=1  pci=off nomodules init=/thi/init"
