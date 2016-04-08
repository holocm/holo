/*******************************************************************************
*
* Copyright 2016 Stefan Majewsky <majewsky@gmx.net>
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

package impl

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

var (
	lockPath string
	lockFile *os.File
)

//AcquireLockfile will create a lock file to ensure that only one instance of
//Holo is running at the same time. If the operation fails, it will os.Exit().
func AcquireLockfile() {
	//where to store the lock file?
	if RootDirectory() == "/" {
		lockPath = "/run/holo.pid"
	} else {
		lockPath = filepath.Join(RootDirectory(), "holo.pid")
	}

	var err error
	lockFile, err = os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		Errorf(Stderr, "Cannot create lock file %s: %s", lockPath, err.Error())
		//is this the "file exists" error that indicates another running instance?
		suberr := err.(*os.PathError).Err
		if errno, ok := suberr.(syscall.Errno); ok {
			if errno == syscall.EEXIST {
				fmt.Fprintln(Stderr, "This usually means that another instance of Holo is currently running.")
				fmt.Fprintln(Stderr, "If not, you can try to delete the lock file manually.")
			}
		}
		os.Exit(255)
	}
	fmt.Fprintf(lockFile, "%d\n", os.Getpid())
	lockFile.Sync()
}

//ReleaseLockfile removes the lock file created by AcquireLockfile.
func ReleaseLockfile() {
	err := lockFile.Close()
	if err != nil {
		Errorf(Stderr, err.Error())
	}
	err = os.Remove(lockPath)
	if err != nil {
		Errorf(Stderr, err.Error())
	}
}
