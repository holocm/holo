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
	"path/filepath"
	"strings"

	"../common"
)

//RepoFile represents a single file in the configuration repository. The string
//stored in it is the path to the repo file (also accessible as Path()).
type RepoFile string

//NewRepoFile creates a RepoFile instance when its path in the file system is
//known.
func NewRepoFile(path string) RepoFile {
	return RepoFile(path)
}

//Path returns the path to this repo file in the file system.
func (file RepoFile) Path() string {
	return string(file)
}

//TargetPath returns the path to the corresponding target file.
func (file RepoFile) TargetPath() string {
	//the optional ".holoscript" suffix appears only on repo files
	repoFile := file.Path()
	if strings.HasSuffix(repoFile, ".holoscript") {
		repoFile = strings.TrimSuffix(repoFile, ".holoscript")
	}

	//make path relative
	relPath, _ := filepath.Rel(common.ResourceDirectory(), repoFile)
	//remove the disambiguation path element to get to the relPath for the ConfigFile
	//e.g. repoFile = '/usr/share/holo/files/23-foo/etc/foo.conf'
	//  -> relPath  = '23-foo/etc/foo.conf'
	//  -> relPath  = 'etc/foo.conf'
	segments := strings.SplitN(relPath, fmt.Sprintf("%c", filepath.Separator), 2)
	relPath = segments[1]

	return filepath.Join(common.TargetDirectory(), relPath)
}

//Disambiguator returns the disambiguator, i.e. the Path() element before the
//TargetPath() that disambiguates multiple repo entries for the same target file.
func (file RepoFile) Disambiguator() string {
	//make path relative to ResourceDirectory()
	relPath, _ := filepath.Rel(common.ResourceDirectory(), file.Path())
	//the disambiguator is the first path element in there
	segments := strings.SplitN(relPath, fmt.Sprintf("%c", filepath.Separator), 2)
	return segments[0]
}

//ApplicationStrategy returns the human-readable name for the strategy that
//will be employed to apply this repo file.
func (file RepoFile) ApplicationStrategy() string {
	if strings.HasSuffix(file.Path(), ".holoscript") {
		return "passthru"
	}
	return "apply"
}

//DiscardsPreviousBuffer indicates whether applying this file will discard the
//previous file buffer (and thus the effect of all previous application steps).
//This is used as a hint by the application algorithm to decide whether
//application steps can be skipped completely.
func (file RepoFile) DiscardsPreviousBuffer() bool {
	return file.ApplicationStrategy() == "apply"
}

//RepoFiles holds a slice of RepoFile instances, and implements some methods
//to satisfy the sort.Interface interface.
type RepoFiles []RepoFile

func (f RepoFiles) Len() int           { return len(f) }
func (f RepoFiles) Less(i, j int) bool { return f[i].Disambiguator() < f[j].Disambiguator() }
func (f RepoFiles) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
