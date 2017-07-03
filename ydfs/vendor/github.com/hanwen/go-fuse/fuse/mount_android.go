package fuse

import (
	"fmt"
	"os"
	"syscall"
)

func mount(mountPoint string, opts *MountOptions, ready chan<- error) (int, error) {
	f, err := os.OpenFile("/dev/fuse", os.O_RDWR, 0644)
	if err != nil {
		return 0, err
	}
	fd := int(f.Fd())

	option := fmt.Sprintf("fd=%d,rootmode=40000,default_permissions,allow_other,user_id=0,group_id=0", fd)

	err = syscall.Mount("/dev/fuse", mountPoint, "fuse", 0, option)
	if err != nil {
		f.Close()
		return 0, err
	}

	close(ready)
	return fd, nil
}

func unmount(mountPoint string) (err error) {
	return syscall.Unmount(mountPoint, 2)
}
