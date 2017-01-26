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

//Scan returns a slice of all the Entities.
func Scan() []*Entity {
	entities := make(map[string]*Entity)
	//walk over the resource directory to find resources (and thus the corresponding entities)
	resourceDir := common.ResourceDirectory()
	filepath.Walk(resourceDir, func(resourcePath string, resourceFileInfo os.FileInfo, err error) error {
		//skip over unaccessible stuff
		if err != nil {
			return err
		}
		//only look at manageable files (regular files or symlinks)
		if !(resourceFileInfo.Mode().IsRegular() || common.IsFileInfoASymbolicLink(resourceFileInfo)) {
			return nil
		}
		// don't consider resourceDir itself to be a resource (it might be a symlink)
		if resourcePath == resourceDir {
			return nil
		}
		//only look at files within subdirectories (files in the resource directory
		//itself are skipped)
		relPath, _ := filepath.Rel(resourceDir, resourcePath)
		if !strings.ContainsRune(relPath, filepath.Separator) {
			return nil
		}

		//create new Entity if necessary and store the resource in it
		resource := NewResource(resourcePath)
		entityPath := resource.EntityPath()
		if entities[entityPath] == nil {
			entities[entityPath] = NewEntity(entityPath)
		}
		entities[entityPath].AddResource(resource)
		return nil
	})

	//walk over the base directory to find orphaned entities
	baseDir := common.BaseDirectory()
	filepath.Walk(baseDir, func(basePath string, baseFileInfo os.FileInfo, err error) error {
		//skip over unaccessible stuff
		if err != nil {
			return err
		}
		//only look at manageable files (regular files or symlinks)
		if !(baseFileInfo.Mode().IsRegular() || common.IsFileInfoASymbolicLink(baseFileInfo)) {
			return nil
		}
		//don't consider baseDir itself to be a base (it might be a symlink)
		if basePath == baseDir {
			return nil
		}

		//ensure that there is an Entity for this base
		//(it could be orphaned)
		entityPath, _ := filepath.Rel(baseDir, basePath)
		entity := NewEntity(entityPath)
		if entities[entityPath] == nil {
			entities[entityPath] = entity
		}
		return nil
	})

	//flatten result into list
	result := make([]*Entity, 0, len(entities))
	for _, entity := range entities {
		result = append(result, entity)
	}

	sort.Sort(entityList(result))
	return result
}

type entityList []*Entity

func (f entityList) Len() int           { return len(f) }
func (f entityList) Less(i, j int) bool { return f[i].relPath < f[j].relPath }
func (f entityList) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
