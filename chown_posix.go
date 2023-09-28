//go:build linux || darwin

package main

import (
	"os"
	"strconv"
)

func chown(path string, user string, group string) error {
	uid, err := strconv.Atoi(user)
	if err != nil {
		return err
	}

	gid, err := strconv.Atoi(group)
	if err != nil {
		return err
	}

	return os.Chown(path, uid, gid)
}
