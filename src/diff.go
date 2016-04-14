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

func pathsForDiffOf(e Entity) (string, string, error) {
	//make sure that the directory for these files does exist
	dirPath := filepath.Join(os.Getenv("HOLO_CACHE_DIR"), e.EntityID())
	err := os.Mkdir(dirPath, 0755)
	if err != nil {
		return "", "", err
	}

	return filepath.Join(dirPath, "expected.toml"), filepath.Join(dirPath, "actual.toml"), nil
}

//SerializeDefinitionIntoFile writes the given EntityDefinition as a TOML file.
func SerializeDefinitionIntoFile(def EntityDefinition, path string) error {
	//reset "system" flag (which we don't want to serialize)
	var isSystem bool
	switch def := def.(type) {
	case *GroupDefinition:
		isSystem = def.System
		def.System = false
	case *UserDefinition:
		isSystem = def.System
		def.System = false
	default:
		panic("unreachable")
	}

	//serialize attributes
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "[[%s]]\n", def.TypeName())
	err := toml.NewEncoder(&buf).Encode(def)
	if err != nil {
		return err
	}

	//restore "system" flag
	switch def := def.(type) {
	case *GroupDefinition:
		def.System = isSystem
	case *UserDefinition:
		def.System = isSystem
	default:
		panic("unreachable")
	}

	return ioutil.WriteFile(path, buf.Bytes(), 0644)
}

//PrepareDiff implements the Entity interface.
func (group Group) PrepareDiff() (string, string, error) {
	//does this group exist already?
	actualDef, err := group.GetProvisionedState()
	if err != nil {
		return "", "", err
	}

	//prepare paths
	expectedPath, actualPath, err := pathsForDiffOf(group)
	if err != nil {
		return "", "", err
	}

	//write actual state
	if actualDef != nil {
		err := SerializeDefinitionIntoFile(actualDef, actualPath)
		if err != nil {
			return "", "", err
		}
	}

	//write expected state
	if !group.Orphaned {
		//merge actual state into definition where definition does not define anything
		g := group
		if g.GID == 0 && actualDef != nil {
			g.GID = actualDef.(*GroupDefinition).GID
		}

		err := SerializeDefinitionIntoFile(&g.GroupDefinition, expectedPath)
		if err != nil {
			return "", "", err
		}
	}

	return expectedPath, actualPath, nil
}

//PrepareDiff implements the Entity interface.
func (user User) PrepareDiff() (string, string, error) {
	//does this user exist already?
	actualDef, err := user.GetProvisionedState()
	if err != nil {
		return "", "", err
	}

	//prepare paths
	expectedPath, actualPath, err := pathsForDiffOf(user)
	if err != nil {
		return "", "", err
	}

	//write actual state
	if actualDef != nil {
		err := SerializeDefinitionIntoFile(actualDef, actualPath)
		if err != nil {
			return "", "", err
		}
	}

	//write expected state
	if !user.Orphaned {
		//merge actual state into definition where definition does not define anything
		u := user
		if actualDef != nil {
			actualUser := actualDef.(*UserDefinition)
			if u.UID == 0 {
				u.UID = actualUser.UID
			}
			if u.Home == "" {
				u.Home = actualUser.Home
			}
			if u.Group == "" {
				u.Group = actualUser.Group
			}
			//TODO: u.Groups
			if u.Shell == "" {
				u.Shell = actualUser.Shell
			}
		}

		err := SerializeDefinitionIntoFile(&u.UserDefinition, expectedPath)
		if err != nil {
			return "", "", err
		}
	}

	return expectedPath, actualPath, nil
}
