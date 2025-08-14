package util

import (
	"os"
	"syscall"
)

func OpenFileSecure(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag|syscall.O_NOFOLLOW, perm)
}
