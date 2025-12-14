//go:build !windows

package node

var datadirInUseErrnos = map[uint]bool{11: true, 32: true, 35: true}