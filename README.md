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
	vol, err := qemu.NewEncryptedImage("rockyLinux.qcow2", qemu.ImageFormatQCOW2, "yourVolumeSecret", 100*GiB)
	if err != nil {
		panic(err.Error())
	}

	// Create the volume after applying the configuration options
	err = vol.Create()
	if err != nil {
		panic(err.Error())
	}

}
```


### Open an existing volume

```go
img, err := qemu.OpenImage("rockyLinux.qcow2")
if err != nil {
	panic(err.Error())
}

fmt.Println("image", img.Path, "format", img.Format, "size", img.Size)
```


### Open an existing *encrypted* volume

```go
img, err := qemu.OpenEncryptedImage("rockyLinux.qcow2", "yourVolumeSecret")
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

## License

WTFPL (Public Domain)
