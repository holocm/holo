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

//GroupDefinition represents a UNIX group (as registered in /etc/group).
type GroupDefinition struct {
	Name   string `toml:"name"`             //the group name (the first field in /etc/group)
	GID    int    `toml:"gid,omitzero"`     //the GID (the third field in /etc/group), or 0 if no specific GID is enforced
	System bool   `toml:"system,omitempty"` //whether the group is a system group (this influences the GID selection if GID = 0)
}

//TypeName implements the EntityDefinition interface.
func (g *GroupDefinition) TypeName() string { return "group" }

//EntityID implements the EntityDefinition interface.
func (g *GroupDefinition) EntityID() string { return "group:" + g.Name }

//Attributes implements the EntityDefinition interface.
func (g *GroupDefinition) Attributes() string {
	var attrs []string
	if g.System {
		attrs = append(attrs, "type: system")
	}
	if g.GID > 0 {
		attrs = append(attrs, fmt.Sprintf("GID: %d", g.GID))
	}
	return strings.Join(attrs, ", ")
}

//WithSerializableState implements the EntityDefinition interface.
func (g *GroupDefinition) WithSerializableState(callback func(EntityDefinition)) {
	//we don't want to serialize the `system` attribute in diffs etc.
	system := g.System
	g.System = false
	callback(g)
	g.System = system
}
