# init

This is a simple init system that makes it easy to run docker images as firecracker vms. It is based on fly.io's init and is written in Go. thi/init can boot docker imaged based root filesystem drives without the need to pre configure rc-init or some alternative init system before hand.

This project is part of an educational endeavor to build tooling around firecracker to make running micro vms trivial for the average user. We're looking forward to adding more features along the way as we learn more about init systems and linux as a whole.

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
