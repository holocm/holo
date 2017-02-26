/*******************************************************************************
*
* Copyright 2015 Stefan Majewsky <majewsky@gmx.net>
* Copyright 2017 Luke Shumaker <lukeshu@sbcglobal.net>
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
	//step 1: check if a system update installed a new version of the stock
	//configuration
	//
	// This has to come first because it might shuffle some files
	// around, and if we do anything else first, we might end up
	// stat()ing the wrong file.
	newBasePath, newBase, err := entity.GetNewBase()
	if err != nil {
		return false, err
	}

	// step 2: Load our 3 versions into memory.
	current, err := entity.GetCurrent()
	if err != nil && !os.IsNotExist(err) {
		if pe, ok := err.(*os.PathError); ok {
			err = errors.New("skipping target: " + pe.Err.Error())
		}
		return false, err
	}

	base, err := entity.GetBase()
	if err != nil && !os.IsNotExist(err) {
		if pe, ok := err.(*os.PathError); ok {
			err = errors.New("skipping target: " + pe.Err.Error())
		}
		return false, err
	}

	provisioned, err := entity.GetProvisioned()
	if err != nil && !os.IsNotExist(err) {
		if pe, ok := err.(*os.PathError); ok {
			err = errors.New("skipping target: " + pe.Err.Error())
		}
		return false, err
	}

	////////////////////////////////////////////////////////////////////////

	//step 1: make sure there is a current file (unless --force)
	needForcefulReprovision := false
	if !current.Manageable {
		if !base.Manageable {
			return false, errors.New("skipping target: not a manageable file")
		}
		if withForce {
			needForcefulReprovision = true
		} else {
			return false, ErrNeedForceToRestore
		}
	}

	//step 2: if we don't have a base yet, the file at current *is*
	//the base which we have to copy now
	if !base.Manageable && current.Manageable {
		baseDir := filepath.Dir(base.Path)
		err := os.MkdirAll(baseDir, 0755)
		if err != nil {
			return false, fmt.Errorf("Cannot create directory %s: %s", baseDir, err.Error())
		}

		err = current.Write(base.Path)
		if err != nil {
			return false, fmt.Errorf("Cannot copy %s to %s: %s", current.Path, base.Path, err.Error())
		}
		tmp := current
		tmp.Path = base.Path
		base = current
	}

	if !base.Manageable {
		return false, errors.New("skipping target: not a manageable file")
	}

	//step 3: check if a system update installed a new version of the stock
	//configuration
	if newBase.Manageable {
		//an updated stock configuration is available at newBase.Path
		//(but show it to the user as newBasePath)
		fmt.Printf(">> found updated target base: %s -> %s\n", newBasePath, base.Path)
		err := newBase.Write(base.Path)
		if err != nil {
			return false, fmt.Errorf("Cannot copy %s to %s: %v", newBase.Path, base.Path, err)
		}
		_ = os.Remove(newBase.Path) //this can fail silently
		newBase.Path = base.Path
		base = newBase
	}

	//step 4: apply the resources *if* the version at current is the one
	//installed by the package (which can be found at base); complain if
	//the user made any changes to config files governed by holo (this check is
	//overridden by the --force option)

	//render desired state of entity
	desired, err := entity.GetDesired(base)
	if err != nil {
		return false, err
	}

	//compare it against the last provisioned version (which must exist at this point
	//unless we are using --force)
	needToWriteTarget := true
	if current.Manageable && provisioned.Manageable {
		needToWriteTarget = !current.EqualTo(desired)
		if !current.EqualTo(provisioned) {
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
	if !needForcefulReprovision && provisioned.Manageable {
		if desired.EqualTo(provisioned) {
			//since we did not do anything, don't report this
			return true, nil
		}
	}

	//save a copy of the provisioned config file to check for manual
	//modifications in the next Apply() run
	if !provisioned.Manageable || !desired.EqualTo(provisioned) {
		provisionedDir := filepath.Dir(provisioned.Path)
		err = os.MkdirAll(provisionedDir, 0755)
		if err != nil {
			return false, fmt.Errorf("Cannot write %s: %s", provisioned.Path, err.Error())
		}
		err = desired.Write(provisioned.Path)
		if err != nil {
			return false, err
		}
	}

	//we're done now if the target already has the correct contents
	if !needToWriteTarget {
		return true, nil
	}

	//write the result buffer to the target and copy
	//owners/permissions from base file to target file
	newTargetPath := current.Path + ".holonew"
	err = desired.Write(newTargetPath)
	if err != nil {
		return false, err
	}
	//move $target.holonew -> $target atomically (to ensure that there is
	//always a valid file at $target)
	return false, os.Rename(newTargetPath, current.Path)
}

//GetBase return the package manager-supplied base version of the
//entity, as recorded the last time it was provisioned.
func (entity *Entity) GetBase() (common.FileBuffer, error) {
	return common.NewFileBuffer(entity.PathIn(common.BaseDirectory()))
}

//GetProvisioned returns the recorded last-provisioned state of the
//entity.
func (entity *Entity) GetProvisioned() (common.FileBuffer, error) {
	return common.NewFileBuffer(entity.PathIn(common.ProvisionedDirectory()))
}

//GetCurrent returns the current version of the entity.
func (entity *Entity) GetCurrent() (common.FileBuffer, error) {
	return common.NewFileBuffer(entity.PathIn(common.TargetDirectory()))
}

//GetNewBase returns the base version of the entity, if it has been
//updated by the package manager since last applied.
func (entity *Entity) GetNewBase() (path string, buf common.FileBuffer, err error) {
	realPath, path, err := platform.Implementation().FindUpdatedTargetBase(entity.PathIn(common.TargetDirectory()))
	if err != nil {
		return
	}
	if realPath != "" {
		buf, err = common.NewFileBuffer(realPath)
		return
	}
	return
}

//GetDesired applies all the resources for this Entity onto the base.
func (entity *Entity) GetDesired(base common.FileBuffer) (common.FileBuffer, error) {
	resources := entity.Resources()

	// Optimization: check if we can skip any application steps
	firstStep := 0
	for idx, resource := range resources {
		if resource.DiscardsPreviousBuffer() {
			firstStep = idx
		}
	}
	resources = resources[firstStep:]

	//load the base into a buffer as the start for the application
	//algorithm
	buffer := base
	buffer.Path = entity.PathIn(common.TargetDirectory())

	//apply all the applicable resources in order
	var err error
	for _, resource := range resources {
		buffer, err = resource.ApplyTo(buffer)
		if err != nil {
			return common.FileBuffer{}, err
		}
	}

	return buffer, nil
}
