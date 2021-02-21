package main

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

const (
	FS_IOC_GET_ENCRYPTION_POLICY    = 0x400c6615
	FS_IOC_GET_ENCRYPTION_POLICY_EX = 0xc0096616
)

const (
	FSCRYPT_POLICY_V1           = 0
	FSCRYPT_KEY_DESCRIPTOR_SIZE = 8
	FSCRYPT_POLICY_V2           = 2
	FSCRYPT_KEY_IDENTIFIER_SIZE = 16
)

const (
	FSCRYPT_MODE_AES_256_XTS = 1
	FSCRYPT_MODE_AES_256_CTS = 4
	FSCRYPT_MODE_AES_128_CBC = 5
	FSCRYPT_MODE_AES_128_CTS = 6
	FSCRYPT_MODE_ADIANTUM    = 9
)

type fscrypt_get_policy_ex_arg struct {
	policy_size uint64
	policy      [100]byte
}

const (
	FSCRYPT_POLICY_V1_SIZE = 4 + FSCRYPT_KEY_DESCRIPTOR_SIZE
	FSCRYPT_POLICY_V2_SIZE = 8 + FSCRYPT_KEY_IDENTIFIER_SIZE
)

type fscrypt_policy_v1 struct {
	version                   uint8
	contents_encryption_mode  uint8
	filenames_encryption_mode uint8
	flags                     uint8
	master_key_descriptor     [FSCRYPT_KEY_DESCRIPTOR_SIZE]byte
}

func (policy *fscrypt_policy_v1) setValue(buf []byte) {
	var readbuf [FSCRYPT_POLICY_V1_SIZE]byte
	copy(readbuf[:], buf[:FSCRYPT_POLICY_V1_SIZE])
	*policy = *(*fscrypt_policy_v1)(unsafe.Pointer(&readbuf))
}

type fscrypt_policy_v2 struct {
	version                   uint8
	contents_encryption_mode  uint8
	filenames_encryption_mode uint8
	flags                     uint8
	__reserved                [4]byte
	master_key_identifier     [FSCRYPT_KEY_IDENTIFIER_SIZE]byte
}

func (policy *fscrypt_policy_v2) setValue(buf []byte) {
	var readbuf [FSCRYPT_POLICY_V2_SIZE]byte
	copy(readbuf[:], buf[:FSCRYPT_POLICY_V2_SIZE])
	*policy = *(*fscrypt_policy_v2)(unsafe.Pointer(&readbuf))
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func encryption_mode_to_string(mode uint8) string {
	switch mode {
	case FSCRYPT_MODE_AES_256_XTS:
		return "AES_256_XTS"
	case FSCRYPT_MODE_AES_256_CTS:
		return "AES_256_CTS"
	case FSCRYPT_MODE_AES_128_CBC:
		return "AES_128_CBC"
	case FSCRYPT_MODE_AES_128_CTS:
		return "AES_128_CTS"
	case FSCRYPT_MODE_ADIANTUM:
		return "ADIANTUM"
	default:
		return "UNKNOW"
	}
}

func GetEncryptPolicy(path string) (interface{}, error) {
	var policy interface{}
	if is_exist, _ := PathExists(path); is_exist == false {
		err := errors.New(fmt.Sprintf("path: %v not exist\n", path))
		return policy, err
	}
	f, err := os.Open(path)
	if err != nil {
		return policy, err
	}
	var buf fscrypt_get_policy_ex_arg
	buf.policy_size = 100

	_, _, err = syscall.Syscall(syscall.SYS_IOCTL, uintptr(f.Fd()), uintptr(FS_IOC_GET_ENCRYPTION_POLICY_EX), uintptr(unsafe.Pointer(&buf)))
	if err.Error() != "errno 0" {
		return policy, err
	}

	f.Close()

	//fmt.Printf("policy_size = %d\n", buf.policy_size)
	if buf.policy[0] == FSCRYPT_POLICY_V1 {
		var policy_v1 fscrypt_policy_v1
		policy_v1.setValue(buf.policy[:])
		return policy_v1, nil
	} else if buf.policy[0] == FSCRYPT_POLICY_V2 {
		var policy_v2 fscrypt_policy_v2
		policy_v2.setValue(buf.policy[:])
		return policy_v2, nil
	}
	return policy, errors.New("encrypt policy version is unknow.")
}

func main() {
	for _, path := range os.Args[1:] {

		policy, err := GetEncryptPolicy(path)
		if err != nil {
			fmt.Printf("[%s]:\n", path)
			fmt.Printf("\t%s\n", err)
			continue
		}

		if policy_v1, ok := policy.(fscrypt_policy_v1); ok {
			fmt.Printf("[%s]:\n", path)
			fmt.Printf("\tversion = %d\n", policy_v1.version)
			fmt.Printf("\tcontents_encryption_mode = %d (%s)\n", policy_v1.contents_encryption_mode, encryption_mode_to_string(policy_v1.contents_encryption_mode))
			fmt.Printf("\tfilenames_encryption_mode = %d (%s)\n", policy_v1.filenames_encryption_mode, encryption_mode_to_string(policy_v1.filenames_encryption_mode))
			fmt.Printf("\tflags = 0x%x\n", policy_v1.flags)
			fmt.Printf("\tmaster_key_descriptor[%d] =  0x%02x\n", FSCRYPT_KEY_DESCRIPTOR_SIZE, policy_v1.master_key_descriptor)
		} else if policy_v2, ok := policy.(fscrypt_policy_v2); ok {
			fmt.Printf("[%s]:\n", path)
			fmt.Printf("\tversion = %d\n", policy_v2.version)
			fmt.Printf("\tcontents_encryption_mode = %d (%s)\n", policy_v2.contents_encryption_mode, encryption_mode_to_string(policy_v2.contents_encryption_mode))
			fmt.Printf("\tfilenames_encryption_mode = %d (%s)\n", policy_v2.filenames_encryption_mode, encryption_mode_to_string(policy_v2.filenames_encryption_mode))
			fmt.Printf("\tflags = 0x%x\n", policy_v2.flags)
			fmt.Printf("\tmaster_key_identifier[%d] =  0x%02x\n", FSCRYPT_KEY_IDENTIFIER_SIZE, policy_v2.master_key_identifier)
		} else {
			fmt.Println("encrypt policy version is unknow.")
		}
	}
}
