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

	"github.com/holocm/holo/lib/holo-files/common"
)

//dpkgImpl provides the platform.Impl for dpkg-based distributions (Debian and derivatives).
type dpkgImpl struct{}

func (p dpkgImpl) FindUpdatedTargetBase(targetPath string) (actualPath, reportedPath string, err error) {
	dpkgDistPath := targetPath + ".dpkg-dist" //may be an updated target base
	dpkgOldPath := targetPath + ".dpkg-old"   //may be a backup of the last provisioned target when the updated target base is at targetPath

	//if "${target}.dpkg-old" exists, move it back to $target and move the
	//updated target base to "${target}.dpkg-dist" so that the usual application
	//logic can continue
	if common.IsManageableFile(dpkgOldPath) {
		err := common.MoveFile(targetPath, dpkgDistPath)
		if err != nil {
			return "", "", err
		}
		err = common.MoveFile(dpkgOldPath, targetPath)
		if err != nil {
			return "", "", err
		}
		return dpkgDistPath, fmt.Sprintf("%s (with .dpkg-old)", targetPath), nil
	}

	if common.IsManageableFile(dpkgDistPath) {
		return dpkgDistPath, dpkgDistPath, nil
	}
	return "", "", nil
}

func (p dpkgImpl) AdditionalCleanupTargets(targetPath string) []string {
	//not used by dpkg
	return []string{}
}
