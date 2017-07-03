package ydfs

import (
	"log"
	"time"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

type yundunfs struct{}

func (fs *yundunfs) SetDebug(debug bool) {}

// yundunfs
func (fs *yundunfs) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	log.Printf("GetAttr: name[%v]\n", name)
	return nil, fuse.ENOSYS
}

func (fs *yundunfs) GetXAttr(name string, attr string, context *fuse.Context) ([]byte, fuse.Status) {
	log.Printf("GetXAttr: name[%s], attr[%s]\n", name, attr)
	return nil, fuse.ENOATTR
}

func (fs *yundunfs) SetXAttr(name string, attr string, data []byte, flags int, context *fuse.Context) fuse.Status {
	log.Printf("SetXAttr: name[%s], attr[%s], data[%v], flags[%v]\n", name, attr, data, flags)
	return fuse.ENOSYS
}

func (fs *yundunfs) ListXAttr(name string, context *fuse.Context) ([]string, fuse.Status) {
	log.Printf("ListXAttr: name[%v]\n", name)
	return nil, fuse.ENOSYS
}

func (fs *yundunfs) RemoveXAttr(name string, attr string, context *fuse.Context) fuse.Status {
	log.Printf("RemoveXAttr: name[%v], attr[%v]\n", name, attr)
	return fuse.ENOSYS
}

func (fs *yundunfs) Readlink(name string, context *fuse.Context) (string, fuse.Status) {
	log.Printf("Readlink: name[%v]\n", name)
	return "", fuse.ENOSYS
}

func (fs *yundunfs) Mknod(name string, mode uint32, dev uint32, context *fuse.Context) fuse.Status {
	log.Printf("Mknod: name[%v], mode[%v], dev[%v]\n", name, mode, dev)
	return fuse.ENOSYS
}

func (fs *yundunfs) Mkdir(name string, mode uint32, context *fuse.Context) fuse.Status {
	log.Printf("Mkdir: name[%v], mode[%v]\n", name, mode)
	return fuse.ENOSYS
}

func (fs *yundunfs) Unlink(name string, context *fuse.Context) (code fuse.Status) {
	log.Printf("Unlink: name[%v]\n", name)
	return fuse.ENOSYS
}

func (fs *yundunfs) Rmdir(name string, context *fuse.Context) (code fuse.Status) {
	log.Printf("Rmdir: name[%v]\n", name)
	return fuse.ENOSYS
}

func (fs *yundunfs) Symlink(value string, linkName string, context *fuse.Context) (code fuse.Status) {
	log.Printf("Symlink: value[%v], linkName[%v]\n", value, linkName)
	return fuse.ENOSYS
}

func (fs *yundunfs) Rename(oldName string, newName string, context *fuse.Context) (code fuse.Status) {
	log.Printf("Rename: oldName[%v], newName[%v]\n", oldName, newName)
	return fuse.ENOSYS
}

func (fs *yundunfs) Link(oldName string, newName string, context *fuse.Context) (code fuse.Status) {
	log.Printf("Link: oldName[%v], newName[%v]\n", oldName, newName)
	return fuse.ENOSYS
}

func (fs *yundunfs) Chmod(name string, mode uint32, context *fuse.Context) (code fuse.Status) {
	log.Printf("Chmod: name[%v], mode[%v]\n", name, mode)
	return fuse.ENOSYS
}

func (fs *yundunfs) Chown(name string, uid uint32, gid uint32, context *fuse.Context) (code fuse.Status) {
	log.Printf("Chown: name[%v], uid[%v], gid[%v]\n", name, uid, gid)
	return fuse.ENOSYS
}

func (fs *yundunfs) Truncate(name string, offset uint64, context *fuse.Context) (code fuse.Status) {
	log.Printf("Truncate: name[%v], offset[%v]\n", name, offset)
	return fuse.ENOSYS
}

func (fs *yundunfs) Open(name string, flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	log.Printf("Open: name[%v], flag[%v]\n", name, flags)
	return nil, fuse.ENOSYS
}

func (fs *yundunfs) OpenDir(name string, context *fuse.Context) (stream []fuse.DirEntry, status fuse.Status) {
	log.Printf("OpenDir: name[%v]\n", name)
	return nil, fuse.ENOSYS
}

func (fs *yundunfs) OnMount(nodeFs *pathfs.PathNodeFs) {
	log.Printf("OnMount\n")
}

func (fs *yundunfs) OnUnmount() {
	log.Printf("OnUnmount\n")
}

func (fs *yundunfs) Access(name string, mode uint32, context *fuse.Context) (code fuse.Status) {
	log.Printf("Access: name[%v], mode[%v]\n", name, mode)
	return fuse.ENOSYS
}

func (fs *yundunfs) Create(name string, flags uint32, mode uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	log.Printf("Create: name[%v], flags[%v], mode[%v]\n", name, flags, mode)
	return nil, fuse.ENOSYS
}

func (fs *yundunfs) Utimens(name string, Atime *time.Time, Mtime *time.Time, context *fuse.Context) (code fuse.Status) {
	log.Printf("Utimens: name[%v], Atime[%v], Mtime[%v]\n", name, Atime, Mtime)
	return fuse.ENOSYS
}

func (fs *yundunfs) String() string {
	return "yundunfs"
}

func (fs *yundunfs) StatFs(name string) *fuse.StatfsOut {
	log.Printf("StatFs: name[%v]\n", name)
	return nil
}

func initYDFS() error {
	const mountRootPath = "/mnt/ydfs"

	nfs := pathfs.NewPathNodeFs(&yundunfs{}, nil)
	server, _, err := nodefs.MountRoot(mountRootPath, nfs.Root(), nil)
	if err != nil {
		return err
	}

	go server.Serve()
	return nil
}
