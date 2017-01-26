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

	"github.com/holocm/holo/cmd/holo-files/internal/common"
	"github.com/holocm/holo/cmd/holo-files/internal/platform"
)

//applyNonOrphan performs the complete application algorithm for the given Entity.
//This includes taking a copy of the base if necessary, applying all
//resources, and saving the result in the target path with the correct
//file metadata.
func (entity *Entity) applyNonOrphan(withForce bool) (skipReport bool, err error) {
	//determine the related paths
	var (
		targetPath      = entity.PathIn(common.TargetDirectory())
		basePath        = entity.PathIn(common.BaseDirectory())
		provisionedPath = entity.PathIn(common.ProvisionedDirectory())
	)

	//step 1: entities may only be applied if:
	//option 1: there is a manageable file in the target location (this target
	//file is either the base from the application package or the
	//product of a previous Apply run)
	//option 2: the target file was deleted, but we have a base that we
	//can start from
	needForcefulReprovision := false
	targetExists := common.IsManageableFile(targetPath)
	if !targetExists {
		if !common.IsManageableFile(basePath) {
			return false, errors.New("skipping target: not a manageable file")
		}
		if withForce {
			needForcefulReprovision = true
		} else {
			return false, ErrNeedForceToRestore
		}
	}

	//step 2: if we don't have a base yet, the file at targetPath *is*
	//the base which we have to copy now
	if !common.IsManageableFile(basePath) {
		baseDir := filepath.Dir(basePath)
		err := os.MkdirAll(baseDir, 0755)
		if err != nil {
			return false, fmt.Errorf("Cannot create directory %s: %s", baseDir, err.Error())
		}

		err = common.CopyFile(targetPath, basePath)
		if err != nil {
			return false, fmt.Errorf("Cannot copy %s to %s: %s", targetPath, basePath, err.Error())
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
		fmt.Printf(">> found updated target base: %s -> %s\n", reportedTBPath, basePath)
		err := common.CopyFile(updatedTBPath, basePath)
		if err != nil {
			return false, fmt.Errorf("Cannot copy %s to %s: %s", updatedTBPath, basePath, err.Error())
		}
		_ = os.Remove(updatedTBPath) //this can fail silently
	}

	//step 4: apply the resources *if* the version at targetPath is the one
	//installed by the package (which can be found at basePath); complain if
	//the user made any changes to config files governed by holo (this check is
	//overridden by the --force option)

	//load the last provisioned version
	var provisionedBuffer *common.FileBuffer
	if common.IsManageableFile(provisionedPath) {
		provisionedBuffer, err = common.NewFileBuffer(provisionedPath, targetPath)
		if err != nil {
			return false, err
		}
	}

	//render desired state of entity
	buffer, err := entity.Render()
	if err != nil {
		return false, err
	}

	//compare it against the last provisioned version (which must exist at this point
	//unless we are using --force)
	needToWriteTarget := true
	if targetExists && provisionedBuffer != nil {
		targetBuffer, err := common.NewFileBuffer(targetPath, targetPath)
		if err != nil {
			return false, err
		}
		needToWriteTarget = !targetBuffer.EqualTo(buffer)
		if !targetBuffer.EqualTo(provisionedBuffer) {
			if withForce {
				needForcefulReprovision = true
			} else {
				if needToWriteTarget {
					return false, ErrNeedForceToOverwrite
				}
			}
		}
	}

	//don't do anything more if nothing has changed and the target file has not been touched
	if !needForcefulReprovision && provisionedBuffer != nil {
		if buffer.EqualTo(provisionedBuffer) {
			//since we did not do anything, don't report this
			return true, nil
		}
	}

	//save a copy of the provisioned config file to check for manual
	//modifications in the next Apply() run
	if provisionedBuffer == nil || !buffer.EqualTo(provisionedBuffer) {
		provisionedDir := filepath.Dir(provisionedPath)
		err = os.MkdirAll(provisionedDir, 0755)
		if err != nil {
			return false, fmt.Errorf("Cannot write %s: %s", provisionedPath, err.Error())
		}
		err = buffer.Write(provisionedPath)
		if err != nil {
			return false, err
		}
		err = common.ApplyFilePermissions(basePath, provisionedPath)
		if err != nil {
			return false, err
		}
	}

	//we're done now if the target already has the correct contents
	if !needToWriteTarget {
		//just check ownership again
		return true, common.ApplyFilePermissions(basePath, targetPath)
	}

	//write the result buffer to the target and copy
	//owners/permissions from base file to target file
	newTargetPath := targetPath + ".holonew"
	err = buffer.Write(newTargetPath)
	if err != nil {
		return false, err
	}
	err = common.ApplyFilePermissions(basePath, newTargetPath)
	if err != nil {
		return false, err
	}
	//move $target.holonew -> $target atomically (to ensure that there is
	//always a valid file at $target)
	return false, os.Rename(newTargetPath, targetPath)
}

//Render applies all the resources for this Entity onto the base.
func (entity *Entity) Render() (*common.FileBuffer, error) {
	//check if we can skip any application steps (firstStep = -1 means: start
	//with loading the base and apply all steps, firstStep >= 0 means:
	//start at that application step with an empty buffer)
	firstStep := -1
	resources := entity.Resources()
	for idx, resource := range resources {
		if resource.DiscardsPreviousBuffer() {
			firstStep = idx
		}
	}

	//load the base into a buffer as the start for the application
	//algorithm, unless it will be discarded by an application step
	basePath := entity.PathIn(common.BaseDirectory())
	targetPath := entity.PathIn(common.TargetDirectory())
	var (
		buffer *common.FileBuffer
		err    error
	)
	if firstStep == -1 {
		buffer, err = common.NewFileBuffer(basePath, targetPath)
		if err != nil {
			return nil, err
		}
	} else {
		buffer = common.NewFileBufferFromContents([]byte(nil), targetPath)
	}

	//apply all the applicable resources in order (starting from the first one
	//that matters)
	if firstStep > 0 {
		resources = resources[firstStep:]
	}
	for _, resource := range resources {
		buffer, err = resource.ApplyTo(buffer)
		if err != nil {
			return nil, err
		}
	}

	return buffer, nil
}
