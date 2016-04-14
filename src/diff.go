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
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

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

func (group *GroupDefinition) serializeForDiff(path string) error {
	var buf bytes.Buffer
	buf.Write([]byte("[[group]]\n"))

	err := appendField(&buf, "name", group.Name)
	if err != nil {
		return err
	}

	if group.GID != 0 {
		err := appendField(&buf, "gid", group.GID)
		if err != nil {
			return err
		}
	}

	return ioutil.WriteFile(path, buf.Bytes(), 0644)
}

func (user *UserDefinition) serializeForDiff(path string) error {
	var buf bytes.Buffer
	buf.Write([]byte("[[user]]\n"))

	err := appendField(&buf, "name", user.Name)
	if err != nil {
		return err
	}

	if user.Comment != "" {
		err := appendField(&buf, "comment", user.Comment)
		if err != nil {
			return err
		}
	}

	if user.UID != 0 {
		err := appendField(&buf, "uid", user.UID)
		if err != nil {
			return err
		}
	}

	if user.Home != "" {
		err := appendField(&buf, "home", user.Home)
		if err != nil {
			return err
		}
	}

	if user.Group != "" {
		err := appendField(&buf, "group", user.Group)
		if err != nil {
			return err
		}
	}

	if len(user.Groups) > 0 {
		err := appendField(&buf, "groups", user.Groups)
		if err != nil {
			return err
		}
	}

	if user.Shell != "" {
		err := appendField(&buf, "shell", user.Shell)
		if err != nil {
			return err
		}
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
		err := actualDef.(*GroupDefinition).serializeForDiff(actualPath)
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

		err := g.serializeForDiff(expectedPath)
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
		actualUser := actualDef.(*UserDefinition)
		err := actualUser.serializeForDiff(actualPath)
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

		err := u.serializeForDiff(expectedPath)
		if err != nil {
			return "", "", err
		}
	}

	return expectedPath, actualPath, nil
}

func encodeField(field string, value interface{}) (string, error) {
	var buf bytes.Buffer
	err := toml.NewEncoder(&buf).Encode(map[string]interface{}{field: value})
	return strings.TrimSpace(buf.String()), err
}

func appendField(w io.Writer, field string, value interface{}) error {
	str, err := encodeField(field, value)
	if err != nil {
		return err
	}
	w.Write([]byte(str + "\n"))
	return nil
}
