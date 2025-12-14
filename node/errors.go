package node

import (
	"syscall"

	"github.com/pkg/errors"
)

var (
	ErrDataDirUsed     = errors.New("dataDir already used by another process")
	ErrNodeStopped     = errors.New("node not started")
)

func convertFileLockError(err error) error {
	if errno, ok := err.(syscall.Errno); ok && datadirInUseErrnos[uint(errno)] {
		return ErrDataDirUsed
	}
	return err
}
