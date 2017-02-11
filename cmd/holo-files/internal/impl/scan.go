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
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/holocm/holo/cmd/holo-files/internal/common"
)

//ScanRepo returns a slice of all the TargetFile entities.
func ScanRepo() []*TargetFile {
	//walk over the repo to find repo files (and thus the corresponding target files)
	targets := make(map[string]*TargetFile)
	repoDir := common.ResourceDirectory()
	filepath.Walk(repoDir, func(repoPath string, repoFileInfo os.FileInfo, err error) error {
		//skip over unaccessible stuff
		if err != nil {
			return err
		}
		//only look at manageable files (regular files or symlinks)
		if !(repoFileInfo.Mode().IsRegular() || common.IsFileInfoASymbolicLink(repoFileInfo)) {
			return nil
		}
		//don't consider repoDir itself to be a repo entry (it might be a symlink)
		if repoPath == repoDir {
			return nil
		}
		//only look at files within subdirectories (files in the repo directory
		//itself are skipped)
		relPath, _ := filepath.Rel(repoDir, repoPath)
		if !strings.ContainsRune(relPath, filepath.Separator) {
			return nil
		}

		//create new TargetFile if necessary and store the repo entry in it
		repoEntry := NewRepoFile(repoPath)
		targetPath := repoEntry.TargetPath()
		if targets[targetPath] == nil {
			targets[targetPath] = NewTargetFileFromPathIn(common.TargetDirectory(), targetPath)
		}
		targets[targetPath].AddRepoEntry(repoEntry)
		return nil
	})

	//walk over the target base directory to find orphaned target bases
	targetBaseDir := common.TargetBaseDirectory()
	filepath.Walk(targetBaseDir, func(targetBasePath string, targetBaseFileInfo os.FileInfo, err error) error {
		//skip over unaccessible stuff
		if err != nil {
			return err
		}
		//only look at manageable files (regular files or symlinks)
		if !(targetBaseFileInfo.Mode().IsRegular() || common.IsFileInfoASymbolicLink(targetBaseFileInfo)) {
			return nil
		}
		//don't consider targetBaseDir itself to be a target base (it might be a symlink)
		if targetBasePath == targetBaseDir {
			return nil
		}

		//check if we have seen the config file for this target base
		//(if not, it's orphaned)
		//TODO: s/(targetBase)Path/\1Dir/g and s/(targetBase)File/Path/g
		target := NewTargetFileFromPathIn(targetBaseDir, targetBasePath)
		targetPath := target.PathIn(common.TargetDirectory())
		if targets[targetPath] == nil {
			target.orphaned = true
			targets[targetPath] = target
		}
		return nil
	})

	//flatten result into list
	result := make([]*TargetFile, 0, len(targets))
	for _, target := range targets {
		result = append(result, target)
	}

	sort.Sort(filesByPath(result))
	return result
}

type filesByPath []*TargetFile

func (f filesByPath) Len() int           { return len(f) }
func (f filesByPath) Less(i, j int) bool { return f[i].relTargetPath < f[j].relTargetPath }
func (f filesByPath) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
