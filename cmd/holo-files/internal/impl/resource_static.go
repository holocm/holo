/*******************************************************************************
*
* Copyright 2015 Stefan Majewsky <majewsky@gmx.net>
* Copyright 2017 Luke Shumaker <lukeshu@parabola.nu>
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

	"github.com/holocm/holo/cmd/holo-files/internal/common"
)

// StaticResource is a Resource that is a plain static file that
// replaces the current version of the entity.
type StaticResource struct{ rawResource }

// ApplicationStrategy implements the Resource interface.
func (resource StaticResource) ApplicationStrategy() string { return "apply" }

// DiscardsPreviousBuffer implements the Resource interface.
func (resource StaticResource) DiscardsPreviousBuffer() bool { return true }

// ApplyTo implements the Resource interface.
func (resource StaticResource) ApplyTo(entityBuffer common.FileBuffer) (common.FileBuffer, error) {
	resourceBuffer, err := common.NewFileBuffer(resource.Path())
	if err != nil {
		return common.FileBuffer{}, err
	}
	entityBuffer.Contents = resourceBuffer.Contents
	entityBuffer.Mode = (entityBuffer.Mode &^ os.ModeType) | (resourceBuffer.Mode & os.ModeType)

	//since Linux disregards mode flags on symlinks and always reports 0777 perms,
	//normalize the mode thusly to make FileBuffer.EqualTo() work reliably
	if entityBuffer.Mode&os.ModeSymlink != 0 {
		entityBuffer.Mode = os.ModeSymlink | os.ModePerm
	}
	return entityBuffer, nil
}
