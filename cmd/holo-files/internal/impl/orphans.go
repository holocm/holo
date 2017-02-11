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
)

//scanOrphanedTargetBase locates a target file for a given orphaned target base
//and assesses the situation. This logic is grouped in one function because
//it's used by both `holo scan` and `holo apply`.
func (target *TargetFile) scanOrphanedTargetBase() (theTargetPath, strategy, assessment string) {
	targetPath := target.PathIn(common.TargetDirectory())
	if common.IsManageableFile(targetPath) {
		return targetPath, "restore", "all repository files were deleted"
	}
	return targetPath, "delete", "target was deleted"
}

//handleOrphanedTargetBase cleans up an orphaned target base.
func (target *TargetFile) handleOrphanedTargetBase() []error {
	targetPath, strategy, _ := target.scanOrphanedTargetBase()
	targetBasePath := target.PathIn(common.TargetBaseDirectory())

	var errs []error
	appendError := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}

	switch strategy {
	case "delete":
		//if the package management left behind additional cleanup targets
		//(most likely a backup of our custom configuration), we can delete
		//these too
		cleanupTargets := platform.Implementation().AdditionalCleanupTargets(targetPath)
		for _, otherFile := range cleanupTargets {
			fmt.Printf(">> also deleting %s\n", otherFile)
			appendError(os.Remove(otherFile))
		}
	case "restore":
		//target is still there - restore the target base, *but* before that,
		//check if there is an updated target base
		updatedTBPath, reportedTBPath, err := platform.Implementation().FindUpdatedTargetBase(targetPath)
		appendError(err)
		if updatedTBPath != "" {
			fmt.Printf(">> found updated target base: %s -> %s", reportedTBPath, targetPath)
			//use this target base instead of the one in the TargetBaseDirectory
			appendError(os.Remove(targetBasePath))
			targetBasePath = updatedTBPath
		}

		//now really restore the target base
		appendError(common.CopyFile(targetBasePath, targetPath))
	}

	//target is not managed by Holo anymore, so delete the provisioned target and the target base
	lastProvisionedPath := target.PathIn(common.ProvisionedDirectory())
	err := os.Remove(lastProvisionedPath)
	if err != nil && !os.IsNotExist(err) {
		appendError(err)
	}
	appendError(os.Remove(targetBasePath))

	//TODO: cleanup empty directories below TargetBaseDirectory() and ProvisionedDirectory()
	return errs
}
