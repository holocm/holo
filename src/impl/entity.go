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
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

//Entity represents a key file in the source directory, and the keys
//provisioned by it.
type Entity struct {
	FilePath string //e.g. "/usr/share/holo/ssh-keys/john-doe/login.pub"
	Name     string //e.g. "ssh-keyset:john-doe/login"
	UserName string //e.g. "john-doe"
	BaseName string //e.g. "login"
}

var resourceDirPath = os.Getenv("HOLO_RESOURCE_DIR")
var userNameRxStr = `([a-z_][a-z0-9_-]*\$?)` //from man:useradd(8)
var fileNameRxStr = `([^/]+)`                //forbid unexpected subdirectories
var entityNameRx = regexp.MustCompile(fmt.Sprintf(`^ssh-keyset:%s/%s$`, userNameRxStr, fileNameRxStr))

//filePathRx is for the part below HOLO_RESOURCE_DIR
var filePathRx = regexp.MustCompile(fmt.Sprintf(`^%s/%s.pub$`, userNameRxStr, fileNameRxStr))

func makeEntity(userName, baseName string) *Entity {
	return &Entity{
		FilePath: fmt.Sprintf("%s/%s/%s.pub", resourceDirPath, userName, baseName),
		Name:     fmt.Sprintf("ssh-keyset:%s/%s", userName, baseName),
		UserName: userName,
		BaseName: baseName,
	}
}

//NewEntityFromName constructs a new Entity from the entity name.
func NewEntityFromName(entityName string) (*Entity, error) {
	//check entity name format and deparse into userName + fileName
	match := entityNameRx.FindStringSubmatch(entityName)
	if match == nil {
		return nil, fmt.Errorf("unacceptable entity name: '%s'", entityName)
	}
	return makeEntity(match[0], match[1]), nil
}

//NewEntityFromKeyfilePath constructs a new Entity from the path to the key file.
func NewEntityFromKeyfilePath(path string) (*Entity, error) {
	//make path relative to resourceDirPath
	relPath, err := filepath.Rel(resourceDirPath, path)
	if err != nil {
		return nil, err
	}

	//check file path format and deparse into userName + fileName
	match := filePathRx.FindStringSubmatch(relPath)
	if match == nil {
		return nil, fmt.Errorf("unacceptable source file path: '%s'", path)
	}
	return makeEntity(match[0], match[1]), nil
}

//Keys lists the keys in the key file for this entity.
func (e *Entity) Keys() ([]string, error) {
	//read file
	contents, err := ioutil.ReadFile(e.FilePath)
	if err != nil {
		return nil, err
	}
	//split lines, filter empty and comments
	lines := strings.Split(string(contents), "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if line != "" && line[0] != '#' {
			result = append(result, line)
		}
	}
	return result, nil
}
