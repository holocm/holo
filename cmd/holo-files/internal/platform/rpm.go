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

package platform

import (
	"fmt"

	"github.com/holocm/holo/internal/fs"
)

// rpmImpl provides the platform.Impl for RPM-based distributions.
type rpmImpl struct{}

func (p rpmImpl) FindUpdatedTargetBase(targetPath string) (actualPath, reportedPath string, err error) {
	rpmnewPath := targetPath + ".rpmnew"   //may be an updated target base
	rpmsavePath := targetPath + ".rpmsave" //may be a backup of the last provisioned target when the updated target base is at targetPath

	//if "${target}.rpmsave" exists, move it back to $target and move the
	//updated target base to "${target}.rpmnew" so that the usual application
	//logic can continue
	if fs.IsManageableFile(rpmsavePath) {
		err := fs.MoveFile(targetPath, rpmnewPath)
		if err != nil {
			return "", "", err
		}
		err = fs.MoveFile(rpmsavePath, targetPath)
		if err != nil {
			return "", "", err
		}
		return rpmnewPath, fmt.Sprintf("%s (with .rpmsave)", targetPath), nil
	}

	if fs.IsManageableFile(rpmnewPath) {
		return rpmnewPath, rpmnewPath, nil
	}
	return "", "", nil
}

func (p rpmImpl) AdditionalCleanupTargets(targetPath string) []string {
	//not used by RPM
	return []string{}
}
