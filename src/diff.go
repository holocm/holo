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

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"../localdeps/github.com/BurntSushi/toml"
)

//SerializeDefinitionIntoFile writes the given EntityDefinition as a TOML file.
func SerializeDefinitionIntoFile(def EntityDefinition, path string) error {
	//write header
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "[[%s]]\n", def.TypeName())

	//write attributes
	var err error
	def.WithSerializableState(func(def EntityDefinition) {
		err = toml.NewEncoder(&buf).Encode(def)
	})
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, buf.Bytes(), 0644)
}

//PrepareDiffFor creates temporary files that the frontend can use to generate
//a diff.
func PrepareDiffFor(def EntityDefinition, isOrphaned bool) (expectedPath string, actualPath string, err error) {
	//does this entity exist already?
	actualDef, err := def.GetProvisionedState()
	if err != nil {
		return
	}

	//make sure that the directory for these files does exist
	dirPath := filepath.Join(os.Getenv("HOLO_CACHE_DIR"), def.EntityID())
	err = os.Mkdir(dirPath, 0755)
	if err != nil {
		return
	}

	expectedPath = filepath.Join(dirPath, "expected.toml")
	actualPath = filepath.Join(dirPath, "actual.toml")

	//write actual state
	if actualDef != nil {
		err = SerializeDefinitionIntoFile(actualDef, actualPath)
		if err != nil {
			return
		}
	}

	//write expected state
	if !isOrphaned {
		//merge actual state into definition where definition does not define anything
		serializable := def
		if actualDef != nil {
			serializable, _ = def.Merge(actualDef, MergeEmptyOnly)
		}

		err = SerializeDefinitionIntoFile(serializable, expectedPath)
		if err != nil {
			return
		}
	}

	return
}
