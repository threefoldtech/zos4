package filesystem

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"

	"github.com/threefoldtech/zosv2/modules"
)

var (
	_ Filesystem = (*btrfs)(nil)

	DeviceAlreadyMountedError = fmt.Errorf("device is already mounted")
	DeviceNotMountedError     = fmt.Errorf("device is not mounted")
)

// btrfs is the filesystem implementation for btrfs
type btrfs struct {
	devices DeviceManager
}

// NewBtrfs creates a new filesystem that implements btrfs
func NewBtrfs(manager DeviceManager) Filesystem {
	return &btrfs{manager}
}

func run(ctx context.Context, name string, args ...string) ([]byte, error) {
	output, err := exec.CommandContext(ctx, name, args...).Output()
	if err != nil {
		if err, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("%s", string(err.Stderr))
		}
	}

	return output, nil
}

func (b *btrfs) btrfs(ctx context.Context, args ...string) ([]byte, error) {
	return run(ctx, "btrfs", args...)
}

func (b *btrfs) Create(ctx context.Context, name string, devices []string, policy modules.RaidProfile) (Pool, error) {
	name = strings.TrimSpace(name)
	if len(name) == 0 {
		return nil, fmt.Errorf("invalid name")
	}

	block, err := b.devices.WithLabel(ctx, name)
	if err != nil {
		return nil, err
	}

	if len(block) != 0 {
		return nil, fmt.Errorf("unique name is required")
	}

	for _, device := range devices {
		dev, err := b.devices.Device(ctx, device)
		if err != nil {
			return nil, err
		}

		if dev.Used() {
			return nil, fmt.Errorf("device '%s' is already used", dev)
		}
	}

	args := []string{
		"-L", name,
		"-d", string(policy),
		"-m", string(policy),
	}

	args = append(args, devices...)
	if _, err := run(ctx, "mkfs.btrfs", args...); err != nil {
		return nil, err
	}

	return btrfsPool(name), nil
}

func (b *btrfs) List(ctx context.Context) ([]Pool, error) {
	var pools []Pool
	available, err := BtrfsList(ctx, "", false)
	if err != nil {
		return nil, err
	}

	for _, fs := range available {
		if len(fs.Label) == 0 {
			// we only assume labeled devices are managed
			continue
		}

		pools = append(pools, btrfsPool(fs.Label))
	}

	return pools, nil
}

type btrfsPool string

// Mounted checks is the pool is mounted
// It doesn't check the default mount location of the pool
// but instead check if any of the pool devices is mounted
// under any location
func (p btrfsPool) Mounted() (string, bool) {
	ctx := context.Background()
	list, _ := BtrfsList(ctx, string(p), true)
	if len(list) != 1 {
		return "", false
	}

	return p.mounted(&list[0])
}

func (p btrfsPool) mounted(fs *Btrfs) (string, bool) {
	for _, device := range fs.Devices {
		if target, ok := getMountTarget(device.Path); ok {
			return target, true
		}
	}

	return "", false
}

func (p btrfsPool) Name() string {
	return string(p)
}

func (p btrfsPool) Path() string {
	return path.Join("/mnt", string(p))
}

// Mount mounts the pool in it's default mount location under /mnt/name
func (p btrfsPool) Mount() (string, error) {
	ctx := context.Background()
	list, _ := BtrfsList(ctx, string(p), false)
	if len(list) != 1 {
		return "", fmt.Errorf("unknown pool '%s'", p)
	}

	fs := list[0]

	if mnt, mounted := p.mounted(&fs); mounted {
		return mnt, nil
	}

	mnt := p.Path()
	if err := os.MkdirAll(mnt, 0755); err != nil {
		return "", err
	}

	return mnt, syscall.Mount(fs.Devices[0].Path, mnt, "btrfs", 0, "")
}

func (p btrfsPool) UnMount() error {
	mnt, ok := p.Mounted()
	if !ok {
		return nil
	}

	return syscall.Unmount(mnt, syscall.MNT_DETACH)
}

func (p btrfsPool) AddDevice(device string) error {
	mnt, ok := p.Mounted()
	if !ok {
		return DeviceNotMountedError
	}
	ctx := context.Background()

	_, err := run(ctx, "btrfs", "device", "add", device, mnt)
	return err
}

func (p btrfsPool) RemoveDevice(device string) error {
	mnt, ok := p.Mounted()
	if !ok {
		return DeviceNotMountedError
	}
	ctx := context.Background()

	_, err := run(ctx, "btrfs", "device", "remove", device, mnt)
	return err
}

func (p btrfsPool) Volumes() ([]Volume, error) {
	mnt, ok := p.Mounted()
	if !ok {
		return nil, DeviceNotMountedError
	}

	var volumes []Volume

	subs, err := BtrfsSubvolumeList(context.Background(), mnt)
	if err != nil {
		return nil, err
	}

	for _, sub := range subs {
		volumes = append(volumes, btrfsVolume(path.Join(mnt, sub.Path)))
	}

	return volumes, nil
}

func (p btrfsPool) AddVolume(name string, size uint64) (Volume, error) {
	mnt, ok := p.Mounted()
	if !ok {
		return nil, DeviceNotMountedError
	}

	mnt = path.Join(mnt, name)
	if _, err := run(context.Background(), "btrfs", "subvolume", "create", mnt); err != nil {
		return nil, err
	}

	return btrfsVolume(mnt), nil
}

func (p btrfsPool) RemoveVolume(name string) error {
	mnt, ok := p.Mounted()
	if !ok {
		return DeviceNotMountedError
	}

	mnt = path.Join(mnt, name)
	_, err := run(context.Background(), "btrfs", "subvolume", "delete", mnt)

	return err
}

type btrfsVolume string

func (v btrfsVolume) Path() string {
	return string(v)
}

func (v btrfsVolume) Volumes() ([]Volume, error) {
	var volumes []Volume

	subs, err := BtrfsSubvolumeList(context.Background(), string(v))
	if err != nil {
		return nil, err
	}

	for _, sub := range subs {
		volumes = append(volumes, btrfsVolume(path.Join(string(v), sub.Path)))
	}

	return volumes, nil
}

func (v btrfsVolume) AddVolume(name string, size uint64) (Volume, error) {
	mnt := path.Join(string(v), name)
	if _, err := run(context.Background(), "btrfs", "subvolume", "create", mnt); err != nil {
		return nil, err
	}

	return btrfsVolume(mnt), nil
}

func (v btrfsVolume) RemoveVolume(name string) error {
	mnt := path.Join(string(v), name)
	_, err := run(context.Background(), "btrfs", "subvolume", "remove", mnt)

	return err
}
