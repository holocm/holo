/*******************************************************************************
*
* Copyright 2015-2016 Stefan Majewsky <majewsky@gmx.net>
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

package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

//SerializeDefinitionIntoFile writes the given EntityDefinition as a TOML file.
func SerializeDefinitionIntoFile(def EntityDefinition, path string) error {
	bytes, err := SerializeDefinition(def)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, bytes, 0644)
}

//PrepareDiffFor creates temporary files that the frontend can use to generate
//a diff.
func PrepareDiffFor(def EntityDefinition, isOrphaned bool) error {
	//make sure that the directory for these files does exist
	dirPath := filepath.Join(os.Getenv("HOLO_CACHE_DIR"), def.EntityID())
	err := os.Mkdir(dirPath, 0755)
	if err != nil {
		return err
	}

	provisionedPath := filepath.Join(dirPath, "provisioned.toml")
	actualPath := filepath.Join(dirPath, "actual.toml")

	//get provisioned state of entity
	provisionedState, err := ProvisionedImageDir.LoadImageFor(def)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if provisionedState != nil {
		//TODO: skip serialization and return path to ProvisionedImageDir directly?
		err = SerializeDefinitionIntoFile(provisionedState, actualPath)
		if err != nil {
			return err
		}
	}

	//does this entity exist already?
	actualDef, err := def.GetProvisionedState()
	if err != nil {
		return err
	}

	//write actual state
	if actualDef.IsProvisioned() {
		err = SerializeDefinitionIntoFile(actualDef, actualPath)
		if err != nil {
			return err
		}
	}

	PrintCommandMessage("%s\000%s\000", provisionedPath, actualPath)
	return nil
}
