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
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"../common"
	"../platform"
)

//Apply performs the complete application algorithm for the given TargetFile.
//This includes taking a copy of the target base if necessary, applying all
//repository entries, and saving the result in the target path with the correct
//file metadata.
func apply(target *TargetFile, withForce bool) (skipReport bool, err error) {
	//determine the related paths
	targetPath := target.PathIn(common.TargetDirectory())
	targetBasePath := target.PathIn(common.TargetBaseDirectory())

	//step 1: will only apply targets if:
	//option 1: there is a manageable file in the target location (this target
	//file is either the target base from the application package or the
	//product of a previous Apply run)
	//option 2: the target file was deleted, but we have a target base that we
	//can start from
	needForcefulReprovision := false
	targetExists := common.IsManageableFile(targetPath)
	if !targetExists {
		if !common.IsManageableFile(targetBasePath) {
			return false, errors.New("skipping target: not a manageable file")
		}
		if withForce {
			needForcefulReprovision = true
		} else {
			return false, errors.New("skipping target: file has been deleted by user (use --force to restore)")
		}
	}

	//step 2: if we don't have a target base yet, the file at targetPath *is*
	//the targetBase which we have to copy now
	if !common.IsManageableFile(targetBasePath) {
		targetBaseDir := filepath.Dir(targetBasePath)
		err := os.MkdirAll(targetBaseDir, 0755)
		if err != nil {
			return false, fmt.Errorf("Cannot create directory %s: %s", targetBaseDir, err.Error())
		}

		err = common.CopyFile(targetPath, targetBasePath)
		if err != nil {
			return false, fmt.Errorf("Cannot copy %s to %s: %s", targetPath, targetBasePath, err.Error())
		}
	}

	//step 3: check if a system update installed a new version of the stock
	//configuration
	updatedTBPath, reportedTBPath, err := platform.Implementation().FindUpdatedTargetBase(targetPath)
	if err != nil {
		return false, err
	}
	if updatedTBPath != "" {
		//an updated stock configuration is available at updatedTBPath
		fmt.Printf(">> found updated target base: %s -> %s", reportedTBPath, targetBasePath)
		err := common.CopyFile(updatedTBPath, targetBasePath)
		if err != nil {
			return false, fmt.Errorf("Cannot copy %s to %s: %s", updatedTBPath, targetBasePath, err.Error())
		}
		_ = os.Remove(updatedTBPath) //this can fail silently
	}

	//step 4: apply the repo files *if* the version at targetPath is the one
	//installed by the package (which can be found at targetBasePath); complain if
	//the user made any changes to config files governed by holo (this check is
	//overridden by the --force option)

	//load the last provisioned version
	var lastProvisionedBuffer *FileBuffer
	lastProvisionedPath := target.PathIn(common.ProvisionedDirectory())
	if common.IsManageableFile(lastProvisionedPath) {
		lastProvisionedBuffer, err = NewFileBuffer(lastProvisionedPath, targetPath)
		if err != nil {
			return false, err
		}
	}

	//compare it against the target version (which must exist at this point
	//unless we are using --force)
	if targetExists && lastProvisionedBuffer != nil {
		targetBuffer, err := NewFileBuffer(targetPath, targetPath)
		if err != nil {
			return false, err
		}
		if !targetBuffer.EqualTo(lastProvisionedBuffer) {
			if withForce {
				needForcefulReprovision = true
			} else {
				return false, errors.New("skipping target: file has been modified by user (use --force to overwrite)")
			}
		}
	}

	//check if we can skip any application steps (firstStep = -1 means: start
	//with loading the target base and apply all steps, firstStep >= 0 means:
	//start at that application step with an empty buffer)
	firstStep := -1
	repoEntries := target.RepoEntries()
	for idx, repoFile := range repoEntries {
		if repoFile.DiscardsPreviousBuffer() {
			firstStep = idx
		}
	}

	//load the target base into a buffer as the start for the application
	//algorithm, unless it will be discarded by an application step
	var buffer *FileBuffer
	if firstStep == -1 {
		buffer, err = NewFileBuffer(targetBasePath, targetPath)
		if err != nil {
			return false, err
		}
	} else {
		buffer = NewFileBufferFromContents([]byte(nil), targetPath)
	}

	//apply all the applicable repo files in order (starting from the first one
	//that matters)
	if firstStep > 0 {
		repoEntries = repoEntries[firstStep:]
	}
	for _, repoFile := range repoEntries {
		buffer, err = GetApplyImpl(repoFile)(buffer)
		if err != nil {
			return false, err
		}
	}

	//don't do anything more if nothing has changed and the target file has not been touched
	if !needForcefulReprovision && lastProvisionedBuffer != nil {
		if buffer.EqualTo(lastProvisionedBuffer) {
			//since we did not do anything, don't report this
			return true, nil
		}
	}

	//save a copy of the provisioned config file to check for manual
	//modifications in the next Apply() run
	provisionedDir := filepath.Dir(lastProvisionedPath)
	err = os.MkdirAll(provisionedDir, 0755)
	if err != nil {
		return false, fmt.Errorf("Cannot write %s: %s", lastProvisionedPath, err.Error())
	}
	err = buffer.Write(lastProvisionedPath)
	if err != nil {
		return false, err
	}
	err = common.ApplyFilePermissions(targetBasePath, lastProvisionedPath)
	if err != nil {
		return false, err
	}

	//write the result buffer to the target location and copy
	//owners/permissions from target base to target file
	newTargetPath := targetPath + ".holonew"
	err = buffer.Write(newTargetPath)
	if err != nil {
		return false, err
	}
	err = common.ApplyFilePermissions(targetBasePath, newTargetPath)
	if err != nil {
		return false, err
	}
	//move $target.holonew -> $target atomically (to ensure that there is
	//always a valid file at $target)
	return false, os.Rename(newTargetPath, targetPath)
}
