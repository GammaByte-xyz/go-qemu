package qemu

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"
)

const (
	ImageFormatRAW   = "raw"
	ImageFormatCLOOP = "cloop"
	ImageFormatCOW   = "cow"
	ImageFormatQCOW  = "qcow"
	ImageFormatQCOW2 = "qcow2"
	ImageFormatVDMK  = "vdmk"
	ImageFormatVDI   = "vdi"
	ImageFormatVHDX  = "vhdx"
	ImageFormatVPC   = "vpc"
)

// Image represents a QEMU disk image
type Image struct {
	Path        string     // Image location (filepath)
	Format      string     // Image format
	Size        uint64     // Image size in bytes
	Secret      string     // Image secret, this enabled encryption
	BackingFile string     // Image backing file (filepath)
	Encrypted   bool       // Image encryption value (readonly)
	snapshots   []Snapshot // Image snapshot array
}

// Snapshot represents a QEMU image snapshot
// Snapshots are snapshots of the complete virtual machine including CPU state
// RAM, device state and the content of all the writable disks
type Snapshot struct {
	ID      int       // Snapshot numerical ID
	Name    string    // Snapshot Name
	Date    time.Time // Snapshot creation Date
	VMClock time.Time
}

// NewImage constructs a new Image data structure based
// on the specified parameters
func NewImage(path, format string, size uint64) Image {
	var img Image
	img.Path = path
	img.Format = format
	img.Size = size
	return img
}

// NewEncryptedImage constructs a new Image data structure based
// on the specified parameters
func NewEncryptedImage(path, format, secret string, size uint64) (Image, error) {
	var img Image
	img.Path = path
	img.Format = format
	img.Size = size
	img.Secret = secret
	img.Encrypted = true

	if format != ImageFormatQCOW2 {
		return img, fmt.Errorf("encrypted volumes must be of the type 'ImageFormatQCOW2'")
	}

	return img, nil
}

// OpenImage retrieves the information of the specified image
// file into an Image data structure
func OpenImage(path string) (Image, error) {
	var img Image

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return img, err
	}

	img.Path = path

	img, err := img.retreiveInfos()
	if err != nil {
		return img, err
	}

	if img.Encrypted {
		return img, fmt.Errorf("image is encrypted but secret was not provided")
	}

	return img, nil
}

// OpenEncryptedImage retrieves the information of the specified image
// file into an Image data structure
func OpenEncryptedImage(path, secret string) (Image, error) {
	var img Image

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return img, err
	}

	img.Path = path
	img.Encrypted = true
	img.Secret = secret

	img, err := img.retreiveInfos()
	if err != nil {
		return img, err
	}

	if secret == "" {
		return img, fmt.Errorf("cannot open encrypted image without secret")
	}
	if !img.Encrypted {
		return img, fmt.Errorf("image is not encrypted")
	}

	return img, nil
}

func (i *Image) retreiveInfos() (Image, error) {
	type snapshotInfo struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		DateSec   int64  `json:"date-sec"`
		DateNsec  int64  `json:"date-nsec"`
		ClockSec  int64  `json:"vm-clock-sec"`
		ClockNsec int64  `json:"vm-clock-nsec"`
	}

	type imgInfo struct {
		Snapshots []snapshotInfo `json:"snapshots"`

		Format    string `json:"format"`
		Size      uint64 `json:"virtual-size"`
		Encrypted bool   `json:"encrypted,omitempty"`
	}

	var info imgInfo

	cmd := exec.Command("qemu-img", "info", "--output=json", i.Path)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return *i, fmt.Errorf("'qemu-img info' output: %s", oneLine(out))
	}

	err = json.Unmarshal(out, &info)
	if err != nil {
		return *i, fmt.Errorf("'qemu-img info' invalid json output")
	}

	i.Format = info.Format
	i.Size = info.Size
	if i.Secret != "" {
		i.Encrypted = true
	} else {
		i.Encrypted = info.Encrypted
	}

	i.snapshots = make([]Snapshot, 0)
	for _, snap := range info.Snapshots {
		var s Snapshot

		id, err := strconv.Atoi(snap.ID)
		if err != nil {
			continue
		}

		s.ID = id
		s.Name = snap.Name
		s.Date = time.Unix(snap.DateSec, snap.DateNsec)
		s.VMClock = time.Unix(snap.ClockSec, snap.ClockNsec)

		i.snapshots = append(i.snapshots, s)
	}

	return *i, nil
}

// Snapshots returns the snapshots contained
// within the image
func (i Image) Snapshots() ([]Snapshot, error) {
	_, err := i.retreiveInfos()
	if err != nil {
		return nil, err
	}

	if len(i.snapshots) == 0 {
		return make([]Snapshot, 0), nil
	}

	return i.snapshots, nil
}

// CreateSnapshot creates a snapshot of the image
// with the specified name
func (i *Image) CreateSnapshot(name string) (Snapshot, error) {
	var snap Snapshot
	// Handles normal volumes
	if i.Encrypted == false {
		cmd := exec.Command("qemu-img", "snapshot", "-c", name, i.Path)

		out, err := cmd.CombinedOutput()
		if err != nil {
			return snap, fmt.Errorf("'qemu-img snapshot' output: %s", oneLine(out))
		}
		snaps, err := i.Snapshots()
		if err != nil {
			return snap, err
		}

		var exists bool
		for _, s := range snaps {
			if s.Name == name {
				snap = s
				exists = true
				break
			}
		}

		if exists {
			return snap, nil
		} else {
			return snap, fmt.Errorf("couldn't find newly created snapshot")
		}
	}
	// Handles encrypted volumes
	cmd := exec.Command("qemu-img", "snapshot", "--object", "secret,id=sec0,data="+i.Secret, "--image-opts", "-c", name, "encrypt.format=luks,encrypt.key-secret=sec0,file.filename="+i.Path)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return snap, fmt.Errorf("'qemu-img snapshot' output: %s", oneLine(out))
	}

	snaps, err := i.Snapshots()
	if err != nil {
		return snap, err
	}

	var exists bool
	for _, s := range snaps {
		if s.Name == name {
			snap = s
			exists = true
			break
		}
	}

	if exists {
		return snap, nil
	} else {
		return snap, fmt.Errorf("couldn't find newly created snapshot")
	}
}

// RestoreSnapshot restores the the image to the
// specified snapshot name
func (i Image) RestoreSnapshot(name string) error {
	// Handles normal volumes
	if i.Encrypted == false {
		cmd := exec.Command("qemu-img", "snapshot", "-a", name, i.Path)

		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("'qemu-img snapshot' output: %s", oneLine(out))
		}

		return nil
	}
	// Handles encrypted volumes
	cmd := exec.Command("qemu-img", "snapshot", "--object", "secret,id=sec0,data="+i.Secret, "--image-opts", "-a", name, "encrypt.format=luks,encrypt.key-secret=sec0,file.filename="+i.Path)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("'qemu-img snapshot' output: %s", oneLine(out))
	}

	return nil
}

// DeleteSnapshot deletes the the corresponding
// snapshot from the image
func (i Image) DeleteSnapshot(name string) error {
	if i.Encrypted == false {
		cmd := exec.Command("qemu-img", "snapshot", "-d", name, i.Path)

		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("'qemu-img snapshot' output: %s", oneLine(out))
		}

		return nil
	}

	cmd := exec.Command("qemu-img", "snapshot", "--object", "secret,id=sec0,data="+i.Secret, "--image-opts", "-d", name, "encrypt.format=luks,encrypt.key-secret=sec0,file.filename="+i.Path)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("'qemu-img snapshot' output: %s", oneLine(out))
	}

	return nil
}

// SetBackingFile sets a backing file for the image
// If it is specified, the image will only record the
// differences from the backing file
func (i *Image) SetBackingFile(backingFile string) error {
	if _, err := os.Stat(backingFile); os.IsNotExist(err) {
		return err
	}

	i.BackingFile = backingFile
	return nil
}

// Create actually creates the image based on the Image structure
// using the 'qemu-img create' command. If a secret is set, the volume is provisioned
// with encryption enabled.
func (i Image) Create() error {
	if i.Encrypted == false {
		args := []string{"create", "-f", i.Format, "-o", "preallocation=metadata"}

		if len(i.BackingFile) > 0 {
			args = append(args, "-o")
			args = append(args, fmt.Sprintf("backing_file=%s", i.BackingFile))
		}

		args = append(args, i.Path)
		args = append(args, strconv.FormatUint(i.Size, 10))

		cmd := exec.Command("qemu-img", args...)

		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("'qemu-img create' output: %s", oneLine(out))
		}

		return nil
	}
	if i.Format != ImageFormatQCOW2 {
		return fmt.Errorf("encrypted volumes must be qcow2 format")
	}
	args := []string{"create", "--object", "secret,id=sec0,data=" + i.Secret, "-f", i.Format, "-o", "encrypt.format=luks,encrypt.key-secret=sec0", "-o", "preallocation=metadata"}
	if len(i.BackingFile) > 0 {
		args = append(args, "-o")
		args = append(args, fmt.Sprintf("backing_file=%s", i.BackingFile))
	}
	args = append(args, i.Path)
	args = append(args, strconv.FormatUint(i.Size, 10))

	cmd := exec.Command("qemu-img", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("'qemu-img create' output: %s", oneLine(out))
	}

	return nil

}

// Rebase changes the backing file of the image
// to the specified file path
func (i *Image) Rebase(backingFile string) error {
	i.BackingFile = backingFile

	cmd := exec.Command("qemu-img", "rebase", "-b", backingFile, i.Path)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("'qemu-img rebase' output: %s", oneLine(out))
	}

	return nil
}
