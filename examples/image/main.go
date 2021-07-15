package main

import (
	"fmt"
	"github.com/GammaByte-xyz/go-qemu"
)

const (
	GiB = 1073741824 // 1 GiB = 2^30 bytes
)

func EncryptedVolume() {
	vol, err := qemu.NewEncryptedImage("rockyLinux.qcow2", qemu.ImageFormatQCOW2, "yourPrivateSecret", 100*GiB)
	if err != nil {
		panic(err.Error())
	}

	err = vol.Create()
	if err != nil {
		panic(err.Error())
	}

	snapshot, err := vol.CreateSnapshot("myTestSnapshot")
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Created snapshot %s with ID %d at %s\n", snapshot.Name, snapshot.ID, snapshot.Date.String())
}

func StandardVolume() {
	vol := qemu.NewImage("CentOS.qcow2", qemu.ImageFormatQCOW2, 100*GiB)

	err := vol.Create()
	if err != nil {
		panic(err.Error())
	}

	snapshot, err := vol.CreateSnapshot("myTestSnapshot")
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Created snapshot %s with ID %d at %s\n", snapshot.Name, snapshot.ID, snapshot.Date.String())
}

func main() {
	EncryptedVolume()
	StandardVolume()
}
