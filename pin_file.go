package main

/*
#include <stdio.h>
#include <sys/ioctl.h>

typedef unsigned int __u32;

#define F2FS_IOCTL_MAGIC                0xf5
#define F2FS_IOC_SET_PIN_FILE           _IOW(F2FS_IOCTL_MAGIC, 13, __u32)
#define F2FS_IOC_GET_PIN_FILE           _IOR(F2FS_IOCTL_MAGIC, 14, __u32)

void print_value(void) {
	printf("F2FS_IOC_SET_PIN_FILE = 0x%x\n", F2FS_IOC_SET_PIN_FILE);
	printf("F2FS_IOC_GET_PIN_FILE = 0x%x\n", F2FS_IOC_GET_PIN_FILE);
}
*/
//import "C"
import (
	"flag"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

const (
	F2FS_IOC_SET_PIN_FILE = 0x4004f50d
	F2FS_IOC_GET_PIN_FILE = 0x8004f50e
)

func ispined(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}

	var buf uint
	_, _, err = syscall.Syscall(syscall.SYS_IOCTL, uintptr(f.Fd()), uintptr(F2FS_IOC_GET_PIN_FILE), uintptr(unsafe.Pointer(&buf)))
	if err.Error() != "errno 0" {
		return false, err
	}

	if buf == 0 {
		return false, nil
	}

	return true, nil
}

func pinfile(path string) error {
	f, err := os.OpenFile(path, os.O_RDWR, 0777)
	if err != nil {
		return err
	}

	var buf uint = 1
	_, _, err = syscall.Syscall(syscall.SYS_IOCTL, uintptr(f.Fd()), uintptr(F2FS_IOC_SET_PIN_FILE), uintptr(unsafe.Pointer(&buf)))
	if err.Error() != "errno 0" {
		return err
	}

	return nil
}

func main() {
	//C.print_value()

	var set_pin_flag bool
	flag.BoolVar(&set_pin_flag, "s", false, "set pin file")
	flag.Parse()

	fmt.Println(flag.Args())
	for _, path := range flag.Args() {
		if set_pin_flag {
			err := pinfile(path)
			if err != nil {
				fmt.Println(path, ": ", err)
			}
		} else {
			ret, err := ispined(path)
			if err != nil {
				fmt.Println(path, ": ", err)
				continue
			}
			fmt.Println(path, ret)
		}
	}
}
