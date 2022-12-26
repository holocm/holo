/*******************************************************************************
*
* Copyright 2015 Stefan Majewsky <majewsky@gmx.net>
*
* This file is part of Holo.
*
* Holo is free software: you can redistribute it and/or modify it under the
* terms of the GNU General Public License as published by the Free Software
* Foundation, either version 3 of the License, or (at your option) any later
* version.
*
* Holo is distributed in the hope that it will be useful, but WITHOUT ANY
* WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
* A PARTICULAR PURPOSE. See the GNU General Public License for more details.
*
* You should have received a copy of the GNU General Public License along with
* Holo. If not, see <http://www.gnu.org/licenses/>.
*
*******************************************************************************/

package fs

import (
	"os"
	"syscall"
)

// IsManageableFile returns whether the file can be managed by Holo (i.e. is a
// regular file or a symlink).
func IsManageableFile(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular() || IsFileInfoASymbolicLink(info)
}

// IsManageableFileInfo returns whether the given FileInfo refers to a
// manageable file (i.e. a regular file or a symlink).
func IsManageableFileInfo(info os.FileInfo) bool {
	return info.Mode().IsRegular() || IsFileInfoASymbolicLink(info)
}

// IsFileInfoASymbolicLink returns whether the given FileInfo describes a symlink.
func IsFileInfoASymbolicLink(fileInfo os.FileInfo) bool {
	return (fileInfo.Mode() & os.ModeType) == os.ModeSymlink
}

// CopyMode is an enumeration for the different modes of behavior for CopyFile().
type CopyMode int

const (
	//CopyContentsFileModeAndOwnership is a CopyMode.
	CopyContentsFileModeAndOwnership CopyMode = iota
	//CopyContentsAndExecutableBitOnly is a CopyMode.
	CopyContentsAndExecutableBitOnly
)

// CopyFile copies a regular file or symlink, including the file metadata.
func CopyFile(fromPath, toPath string, mode CopyMode) error {
	info, err := os.Lstat(fromPath)
	if err != nil {
		return err
	}
	if info.Mode().IsRegular() {
		return copyFileImpl(fromPath, toPath, info, mode)
	}
	return copySymlinkImpl(fromPath, toPath)
}

func copyFileImpl(fromPath, toPath string, fromInfo os.FileInfo, mode CopyMode) error {
	//copy contents
	data, err := os.ReadFile(fromPath)
	if err != nil {
		return err
	}
	err = os.WriteFile(toPath, data, 0600)
	if err != nil {
		return err
	}

	//apply permissions, ownership, modification date from source file to target file
	//NOTE: We cannot just pass the FileMode in os.WriteFile(), because its
	//FileMode argument is only applied when a new file is created, not when
	//an existing one is truncated.
	switch mode {
	case CopyContentsFileModeAndOwnership:
		//apply permissions
		err = os.Chmod(toPath, fromInfo.Mode())
		if err != nil {
			return err
		}

		//apply ownership
		stat := fromInfo.Sys().(*syscall.Stat_t) // UGLY
		return os.Chown(toPath, int(stat.Uid), int(stat.Gid))
	case CopyContentsAndExecutableBitOnly:
		//apply executable bit if set on source
		return os.Chmod(toPath, 0600|(fromInfo.Mode()&0100))
	default:
		//unreachable, but we don't care
		return nil
	}
}

func copySymlinkImpl(fromPath, toPath string) error {
	//read link target
	target, err := os.Readlink(fromPath)
	if err != nil {
		return err
	}
	//remove old file or link if it exists
	if IsManageableFile(toPath) {
		err = os.Remove(toPath)
		if err != nil {
			return err
		}
	}
	//create new link
	err = os.Symlink(target, toPath)
	if err != nil {
		return err
	}

	return nil
}

// MoveFile is like CopyFile, but it removes the fromPath after successful
// copying.
func MoveFile(fromPath, toPath string) error {
	err := CopyFile(fromPath, toPath, CopyContentsFileModeAndOwnership)
	if err != nil {
		return err
	}
	return os.Remove(fromPath)
}
