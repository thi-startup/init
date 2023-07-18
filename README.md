# init

Thi/init aimed at simplifying the process of running Docker images as Firecracker VMs. Built upon fly.io's init and written in Go, it provides a straightforward solution for booting Docker image-based root filesystem drives without the need for pre-configuring rc-init or any alternative init system.

Our primary goal is to support educational endeavors centered around Firecracker, empowering average users to effortlessly run micro VMs. As we delve deeper into the world of init systems and Linux, we are excited to expand thi/init's feature set and offer even more functionality.

### demo
- spawning a simple bash shell in firecracker

  ![Made with VHS](https://vhs.charm.sh/vhs-6MF5u4Tsar87Ryp1r7yIk6.gif)
## Usage

### build it yourself

- Clone this repo and build this with `CGO_ENABLED=0 go build .`
- Create a device for the init
```shell
fallocate -l 50M tmpinit
mkfs.ext2 tmpinit
mkdir initm
mount -o loop,noatime tmpinit initm
mkdir initm/thi
cp init initm/thi
cp run.json initm/thi/run.json
umount initm
```
- attach your init drive as `/dev/vda`
- attach your rootfs as `/dev/vdb`

### run.json file

This file contains the basic config needed by the init to run your image. Here is a demo version that spawns a bash shell, essentially the one used in the demo

```json
{
    "RootDevice": "/dev/vdb",
    "Hostname": "yikes",
    "CmdOverride": ["bash"],
    "ImageConfig": {
        "Env": ["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"],
        "WorkingDir": "/out",
        "User": "root"
    },
    "ExtraEnv": ["TERM=xterm"],
    "Mounts": [
      {
        "DevicePath": "/dev/vdc",
        "MountPath": "/out"
      }
    ],
    "EtcResolv": {
      "Nameservers": ["8.8.8.8", "8.8.4.4"]
   }
}
```
