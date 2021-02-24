package main

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"
)

var testdirpath string = "/data/system/aa/"

var testfilenum int64 = 100

const (
	BLOCK_SIZE = 4096
)

var testbuf []byte = make([]byte, BLOCK_SIZE)

func testbuf_init() {
	var i int64
	for i = 0; i < BLOCK_SIZE; i++ {
		testbuf[i] = 'a'
	}
}

func check_file(filepath string) error {
	fd, err := os.Open(filepath)
	if err != nil {
		//fmt.Println("Open file failed, ", err)
		return nil
	}
	defer fd.Close()

	buf := make([]byte, BLOCK_SIZE)
	for {
		count, err := fd.Read(buf)
		if err != nil {
			//fmt.Println("Read failed, ", err)
			break
		}
		if count < BLOCK_SIZE {
			fmt.Printf("Read less than %d bytes\n", BLOCK_SIZE)
			break
		}

		if bytes.Equal(testbuf, buf) == false {
			fmt.Printf("check file %v failed.\n", filepath)
			fmt.Println(buf)
			os.Exit(-1)
			return errors.New(fmt.Sprintf("check file %v failed.", filepath))
		}
	}
	return nil
}

func write_file(filepath string) {
	fd, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		//fmt.Println("OpenFile failed, ", err)
		return
	}
	defer fd.Close()

	for block_count := 500 + rand.Int63()%1000; block_count > 0; block_count-- {
		count, err := fd.Write(testbuf)
		if err != nil {
			//fmt.Println("Write failed, ", err)
			break
		}
		if count < BLOCK_SIZE {
			fmt.Printf("Write %v less than %d bytes\n", filepath, BLOCK_SIZE)
			break
		}
	}
}

func rename_file(filepath string) error {
	tmpfile := filepath + "_backup"
	write_file(tmpfile)
	check_file(tmpfile)
	err := os.Rename(tmpfile, filepath)
	if err != nil {
		//fmt.Printf("Rename file %v->%v failed.\n", tmpfile, filepath)
		return nil
	}
	check_file(filepath)
	return nil
}

func write_file_check(filepath string) {
	for {
		check_file(filepath)
		write_file(filepath)
		check_file(filepath)
		//time.Sleep(1 * time.Millisecond)
	}
}
func rename_file_check(filepath string) {
	for {
		check_file(filepath)
		rename_file(filepath)
		check_file(filepath)
		//time.Sleep(1 * time.Millisecond)
	}
}

func del_file(filepath string) {
	for {
		err := os.Remove(filepath)
		if err != nil {
			//fmt.Printf("Remove file %v faild\n", err)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func testfile(filepath string) {
	check_file(filepath)
	go del_file(filepath)
	time.Sleep(time.Second)
	go rename_file_check(filepath)
	go write_file_check(filepath)
}

func main() {
	testbuf_init()
	fmt.Println("file test started.")

	var i int64
	for i = 0; i < testfilenum; i++ {
		check_file(testdirpath + fmt.Sprintf("a%d.txt", i))
	}

	os.RemoveAll(testdirpath)
	os.Mkdir(testdirpath, 0777)

	for i = 0; i < testfilenum; i++ {
		go testfile(testdirpath + fmt.Sprintf("a%d.txt", i))
	}
	for {
		time.Sleep(time.Second)
	}
}
