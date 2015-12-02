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

package plugins

import (
	"os"
	"path/filepath"
)

var cachePath string

func init() {
	cachePath = filepath.Join(RootDirectory(), "tmp/holo-cache")
	err := doInit()
	if err != nil {
		r := Report{Action: "Errors occurred during", Target: "startup"}
		r.AddError(err.Error())
		r.Print()
		panic("startup failed")
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

//CleanupRuntimeCache tries to cleanup /tmp/holo-cache.
func CleanupRuntimeCache() {
	_ = os.RemoveAll(cachePath) //fail silently
}
