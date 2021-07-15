# go-qemu

Golang interface to the QEMU hypervisor (particularly volume operations), forked from quadrifoglio/go-qemu.

This fork adds additional functionality such as NBD connections, volume encryption, cluster size adjustments,
compatibility options, and other optimizations that could be potentially useful in certain scenarios.
## Installation

```
go get github.com/GammaByte-xyz/go-qemu
```

You obviously need QEMU to use this tool.

## Usage

### Create a 100 GiB volume

```go
package main

import (
	"github.com/GammaByte-xyz/go-qemu"
)
const (
    GiB = 1073741824 // 1 GiB = 2^30 bytes
)

func main() {
    volume := qemu.NewImage("myVolume.qcow2", qemu.ImageFormatQCOW2, 100*GiB)
    
    // Create the volume after applying the configuration options 
    err := volume.Create()
    if err != nil {
        panic(err.Error()) // Never invoke a panic if your application 
    }                      // runs as a daemon!

}
```

### Create a 100 GiB *encrypted* volume

```go
package main

import (
	"github.com/GammaByte-xyz/go-qemu"
)
const (
    GiB = 1073741824 // 1 GiB = 2^30 bytes
)

func main() {
    volume := qemu.NewImage("myVolume.qcow2", qemu.ImageFormatQCOW2, 100*GiB)
    
    // Create the volume after applying the configuration options 
    err := volume.Create()
    if err != nil {
        panic(err.Error()) // Never invoke a panic if your application 
    }                      // runs as a daemon!

}
```


### Open an existing volume

```go
img, err := qemu.OpenImage("rockylinux.qcow2")
if err != nil {
	panic(err.Error())
}

fmt.Println("image", img.Path, "format", img.Format, "size", img.Size)
```

### Snapshot creation and deletion

```go
err = img.CreateSnapshot("mySnapshot")
if err != nil {
	panic(err.Error())
}

snaps, err := img.Snapshots()
if err != nil {
	panic(err.Error())
}

for _, snapshot := range snaps {
	fmt.Println(snapshot.Name, snapshot.Date)
}
```

### Start a virtual machine

```go
img, err := qemu.OpenImage("debian.qcow2")
if err != nil {
	log.Fatal(err)
}

m := qemu.NewMachine(1, 512) // 1 CPU, 512MiB RAM
m.AddDrive(img)

// x86_64 arch (using qemu-system-x86_64), with kvm
pid, err := m.Start("x86_64", true, func(stderr string) {
        log.Println("QEMU stderr:", stderr)
})

if err != nil {
	log.Fatal(err)
}

fmt.Println("QEMU started on PID", pid)
```

## License

WTFPL (Public Domain)
