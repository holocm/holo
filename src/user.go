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
	"fmt"
	"strings"
)

//UserDefinition represents a UNIX user account (as registered in /etc/passwd).
type UserDefinition struct {
	Name    string   `toml:"name"`              //the user name (the first field in /etc/passwd)
	Comment string   `toml:"comment,omitempty"` //the full name (sometimes also called "comment"; the fifth field in /etc/passwd)
	UID     int      `toml:"uid,omitzero"`      //the user ID (the third field in /etc/passwd), or 0 if no specific UID is enforced
	System  bool     `toml:"system,omitempty"`  //whether the group is a system group (this influences the GID selection if gid = 0)
	Home    string   `toml:"home,omitempty"`    //path to the user's home directory (or empty to use the default)
	Group   string   `toml:"group,omitempty"`   //the name of the user's initial login group (or empty to use the default)
	Groups  []string `toml:"groups,omitempty"`  //the names of supplementary groups which the user is also a member of
	Shell   string   `toml:"shell,omitempty"`   //path to the user's login shell (or empty to use the default)
}

//TypeName implements the EntityDefinition interface.
func (u *UserDefinition) TypeName() string { return "user" }

//EntityID implements the EntityDefinition interface.
func (u *UserDefinition) EntityID() string { return "user:" + u.Name }

//Attributes implements the EntityDefinition interface.
func (u *UserDefinition) Attributes() string {
	var attrs []string
	if u.System {
		attrs = append(attrs, "type: system")
	}
	if u.UID > 0 {
		attrs = append(attrs, fmt.Sprintf("UID: %d", u.UID))
	}
	if u.Home != "" {
		attrs = append(attrs, "home: "+u.Home)
	}
	if u.Group != "" {
		attrs = append(attrs, "login group: "+u.Group)
	}
	if len(u.Groups) > 0 {
		attrs = append(attrs, "groups: "+strings.Join(u.Groups, ","))
	}
	if u.Shell != "" {
		attrs = append(attrs, "login shell: "+u.Shell)
	}
	if u.Comment != "" {
		attrs = append(attrs, "comment: "+u.Comment)
	}
	return strings.Join(attrs, ", ")
}

//WithSerializableState implements the EntityDefinition interface.
func (u *UserDefinition) WithSerializableState(callback func(EntityDefinition)) {
	//we don't want to serialize the `system` attribute in diffs etc.
	system := u.System
	u.System = false
	callback(u)
	u.System = system
}
