package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

const (
	FIBMAP   = 1 // bmap access
	FIGETBSZ = 2 // get the block size used for bmap
)

type FibmapFile struct {
	*os.File
}

func NewFibmapFile(f *os.File) FibmapFile {
	return FibmapFile{f}
}

func (f FibmapFile) Fibmap(block uint) (uint, syscall.Errno) {
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), FIBMAP, uintptr(unsafe.Pointer(&block)))
	return block, err
}

func (f FibmapFile) Figetbsz() (int, syscall.Errno) {
	bsz := int(0)
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), FIGETBSZ, uintptr(unsafe.Pointer(&bsz)))
	return bsz, err
}

func (f FibmapFile) FibmapExtents() ([]uint, syscall.Errno) {
	result := make([]uint, 0)

	bsz, err := f.Figetbsz()
	if err != 0 {
		return nil, err
	}

	stat, _ := f.Stat()
	size := stat.Size()
	if size == 0 {
		return result, syscall.Errno(0)
	}

	blocks := uint((size-1)/int64(bsz)) + 1
	var block uint

	fmt.Println("blocks = ", blocks)
	for i := uint(0); i < blocks; i++ {
		block = i
		_, _, err = syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), FIBMAP, uintptr(unsafe.Pointer(&block)))
		result = append(result, block)
	}

	return result, err
}

func main() {
	file := os.Args[1]

	fd, err := os.OpenFile(file, os.O_RDWR, 0777)
	if err != nil {
		fmt.Printf("Open file %v failed, %v", file, err)
	}
	filemap := NewFibmapFile(fd)
	blocks, _ := filemap.FibmapExtents()
	for _, block := range blocks {
		fmt.Println(block)
	}
}
