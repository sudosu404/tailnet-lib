// SPDX-FileCopyrightText: 2024 Paulo Almeida <almeidapaulopt@gmail.com>
// SPDX-License-Identifier: MIT
package consts

const (
	PermNone         = 0000 // No permissions
	PermOwnerRead    = 0400 // Owner: Read
	PermOwnerWrite   = 0200 // Owner: Write
	PermOwnerExecute = 0100 // Owner: Execute
	PermOwnerAll     = 0700 // Owner: Read, Write, Execute

	PermGroupRead    = 0040 // Group: Read
	PermGroupWrite   = 0020 // Group: Write
	PermGroupExecute = 0010 // Group: Execute
	PermGroupAll     = 0070 // Group: Read, Write, Execute

	PermOthersRead    = 0004 // Others: Read
	PermOthersWrite   = 0002 // Others: Write
	PermOthersExecute = 0001 // Others: Execute
	PermOthersAll     = 0007 // Others: Read, Write, Execute

	PermOwnerGroupRead = 0440 // Owner and Group: Read
	PermOwnerGroupAll  = 0770 // Owner and Group: All

	PermAllRead    = 0444 // Everyone: Read
	PermAllWrite   = 0222 // Everyone: Write
	PermAllExecute = 0111 // Everyone: Execute
	PermAll        = 0777 // Everyone: Read, Write, Execute
)
