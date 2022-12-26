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

	"github.com/holocm/holo/cmd/holo-files/internal/common"
	"github.com/holocm/holo/cmd/holo-files/internal/platform"
	"github.com/holocm/holo/internal/fs"
)

// scanOrphan locates an entity for a given orphaned entity
// and assesses the situation. This logic is grouped in one function because
// it's used by both `holo scan` and `holo apply`.
func (entity *Entity) scanOrphan() (targetPath, strategy, assessment string) {
	targetPath = entity.PathIn(common.TargetDirectory())
	if fs.IsManageableFile(targetPath) {
		return targetPath, "restore", "all repository files were deleted"
	}
	return targetPath, "delete", "target was deleted"
}

// applyOrphan cleans up an orphaned entity.
func (entity *Entity) applyOrphan() []error {
	var errs []error
	appendError := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}

	current, err := entity.GetCurrent()
	if !os.IsNotExist(err) {
		appendError(err)
	}

	provisioned, err := entity.GetProvisioned()
	appendError(err)

	basePath := entity.PathIn(common.BaseDirectory())

	if !current.Manageable { // delete
		//if the package management left behind additional cleanup targets
		//(most likely a backup of our custom configuration), we can delete
		//these too
		cleanupTargets := platform.Implementation().AdditionalCleanupTargets(current.Path)
		for _, path := range cleanupTargets {
			otherFile, err := common.NewFileBuffer(path)
			if err != nil {
				continue
			}
			if otherFile.EqualTo(provisioned) {
				fmt.Printf(">> also deleting %s\n", otherFile.Path)
				appendError(os.Remove(otherFile.Path))
			}
		}

		appendError(os.Remove(provisioned.Path))
		appendError(os.Remove(basePath))
	} else { // restore
		//target is still there - restore the target base, *but* before that,
		//check if there is an updated target base
		updatedTBPath, reportedTBPath, err := platform.Implementation().FindUpdatedTargetBase(current.Path)
		appendError(err)
		if updatedTBPath != "" {
			fmt.Printf(">> found updated target base: %s -> %s", reportedTBPath, current.Path)
			//use this target base instead of the one in the BaseDirectory
			appendError(os.Remove(basePath))
			basePath = updatedTBPath
		}

		appendError(os.Remove(provisioned.Path))
		appendError(fs.MoveFile(basePath, current.Path))
	}

	//TODO: cleanup empty directories below BaseDirectory() and ProvisionedDirectory()
	return errs
}
