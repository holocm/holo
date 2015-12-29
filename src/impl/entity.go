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
	"regexp"
)

//Entity represents a key file in the source directory, and the keys
//provisioned by it.
type Entity struct {
	UserName string
	FileName string
}

var userNameRxStr = `[a-z_][a-z0-9_-]*\$?` //from man:useradd(8)
var entityNameRx = regexp.MustCompile(`^ssh-key:(` + userNameRxStr + `)/([^/]+)$`)

//NewEntityFromName constructs a new Entity from the entity name.
func NewEntityFromName(entityName string) (*Entity, error) {
	//check entity name format and deparse into userName + fileName
	match := entityNameRx.FindStringSubmatch(entityName)
	if match == nil {
		return nil, fmt.Errorf("unacceptable entity name: '%s'", entityName)
	}
	return &Entity{
		UserName: match[0],
		FileName: match[1],
	}, nil
}

//Name returns the entity name.
func (e *Entity) Name() string {
	return fmt.Sprintf("ssh-key:%s/%s", e.UserName, e.FileName)
}
