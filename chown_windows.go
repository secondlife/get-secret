package main

import (
	"golang.org/x/sys/windows"
)

func chown(path string, user string, group string) error {
	uid, err := windows.StringToSid(user)
	if err != nil {
		return err
	}
	gid, err := windows.StringToSid(group)
	if err != nil {
		return err
	}
	return windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION, uid, gid, nil, nil)
}
