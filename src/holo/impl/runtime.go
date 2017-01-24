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
	"io/ioutil"
	"os"
)

var cachePath string

//WithCacheDirectory executes the worker function after having set up a cache
//directory, and ensures that the cache directory is cleaned up afterwards.
func WithCacheDirectory(worker func() (exitCode int)) (exitCode int) {
	var err error
	cachePath, err = ioutil.TempDir(os.TempDir(), "holo.")
	if err != nil {
		Errorf(Stderr, err.Error())
		return 255
	}

	//ensure that the cache is removed even if worker() panics
	defer func() {
		_ = os.RemoveAll(cachePath) //failure to cleanup is non-fatal
		cachePath = ""
	}()

	return worker()
}

//CachePath returns the path below which plugin cache directories can be allocated.
func CachePath() string {
	if cachePath == "" {
		panic("Tried to use cachePath outside WithCacheDirectory() call!")
	}
	return cachePath
}
