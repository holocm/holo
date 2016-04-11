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

package impl

import (
	"fmt"
	"os"
	"path/filepath"
)

var cachePath string

func init() {
	rootDir := RootDirectory()
	if rootDir == "/" {
		//in productive mode, honor the TMPDIR variable (through os.TempDir)
		//and include the PID to avoid collisions between parallel runs of
		//"holo scan" (which is not protected by the /run/holo.pid lockfile)
		cachePath = filepath.Join(os.TempDir(), fmt.Sprintf("holo-%d", os.Getpid()))
	} else {
		//during unit tests, we are free to choose a simple, reproducible cache
		//location
		cachePath = filepath.Join(rootDir, "tmp/holo")
	}

	err := doInit()
	if err != nil {
		Errorf(Stderr, err.Error())
		os.Exit(255)
	}
}

func doInit() error {
	//if the cache directory exists from a previous run, remove it recursively
	err := os.RemoveAll(cachePath)
	if err != nil {
		return err
	}

	//create the cache directory
	return os.MkdirAll(cachePath, 0700)
}

//CachePath returns the path below which plugin cache directories can be allocated.
func CachePath() string {
	return cachePath
}

//CleanupRuntimeCache tries to cleanup the CachePath().
func CleanupRuntimeCache() {
	_ = os.RemoveAll(cachePath) //fail silently
}
