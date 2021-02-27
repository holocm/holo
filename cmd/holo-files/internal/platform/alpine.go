/*******************************************************************************
*
* Copyright 2017 Stefan Majewsky <majewsky@gmx.net>
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

import "github.com/holocm/holo/internal/fs"

//apkImpl provides the platform.Impl for Alpine Linux and derivatives.
type apkImpl struct{}

func (p apkImpl) FindUpdatedTargetBase(targetPath string) (actualPath, reportedPath string, err error) {
	apknewPath := targetPath + ".apk-new"
	if fs.IsManageableFile(apknewPath) {
		return apknewPath, apknewPath, nil
	}
	return "", "", nil
}

func (p apkImpl) AdditionalCleanupTargets(targetPath string) (ret []string) {
	return nil
}
