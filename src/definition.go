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
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"

	"../localdeps/github.com/BurntSushi/toml"
)

const (
	//MergeWhereCompatible can be used as second argument for EntityDefinition.Merge.
	MergeWhereCompatible = false
	//MergeEmptyOnly can be used as second argument for EntityDefinition.Merge.
	MergeEmptyOnly = true
)

//EntityDefinition contains data from a definition file that describes an entity
//(a user account or group). Definitions can also be obtained by scanning the
//user/group databases.
type EntityDefinition interface {
	//TypeName returns the part of the entity ID before the ":", i.e. either
	//"group" or "user".
	TypeName() string
	//EntityID returns exactly that, e.g. "user:john".
	EntityID() string
	//Attributes returns a human-readable stringification of this definition.
	Attributes() string
	//GetProvisionedState reads the current state of this entity from the
	//system database (/etc/passwd or /etc/group). The return value has the same
	//concrete type as the callee. If no entity with the same ID exists in
	//there, a non-nil instance will be returned for which IsProvisioned()
	//yields false.
	GetProvisionedState() (EntityDefinition, error)
	//IsProvisioned must be called on an instance returned from
	//GetProvisionedState(), and will indicate whether this entity is present
	//in the system database (/etc/passwd or /etc/group).
	IsProvisioned() bool
	//WithSerializableState brings the definition into a safely serializable
	//state, executes the callback, and then restores the original state.
	WithSerializableState(callback func(EntityDefinition))
	//Merge constructs a new EntityDefinition of the same concrete type whose
	//attributes are merged from the callee and the argument. The argument's
	//concrete type must be identical to that of the callee. If both sources
	//have different values set for the same attribute, the callee's value
	//takes precedence, and an error is returned in the second argument.
	//If merge conflicts are not a problem, the error argument may be ignored.
	//
	//If `emptyOnly` is true, only empty arguments may be merged.
	Merge(other EntityDefinition, emptyOnly bool) (EntityDefinition, []error)
	//Apply provisions this entity. The argument indicates the currently
	//provisioned state. If the entity is not provisioned yet, it will be nil.
	//If not nil, the argument's concrete type must match the callee.
	Apply(provisioned EntityDefinition) error
	//Cleanup removes the entity from the system.
	Cleanup() error
}

//SerializeDefinition returns a TOML representation of this EntityDefinition.
func SerializeDefinition(def EntityDefinition) ([]byte, error) {
	//write header
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "[[%s]]\n", def.TypeName())

	//write attributes
	var err error
	def.WithSerializableState(func(def EntityDefinition) {
		err = toml.NewEncoder(&buf).Encode(def)
	})
	return buf.Bytes(), err
}

//SerializeDefinitionIntoFile writes the given EntityDefinition as a TOML file.
func SerializeDefinitionIntoFile(def EntityDefinition, path string) error {
	bytes, err := SerializeDefinition(def)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, bytes, 0644)
}

//GroupDefinition represents a UNIX group (as registered in /etc/group).
type GroupDefinition struct {
	Name   string `toml:"name"`             //the group name (the first field in /etc/group)
	GID    int    `toml:"gid,omitzero"`     //the GID (the third field in /etc/group), or 0 if no specific GID is enforced
	System bool   `toml:"system,omitempty"` //whether the group is a system group (this influences the GID selection if GID = 0)
}

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
func (g *GroupDefinition) TypeName() string { return "group" }

//TypeName implements the EntityDefinition interface.
func (u *UserDefinition) TypeName() string { return "user" }

//EntityID implements the EntityDefinition interface.
func (g *GroupDefinition) EntityID() string { return "group:" + g.Name }

//EntityID implements the EntityDefinition interface.
func (u *UserDefinition) EntityID() string { return "user:" + u.Name }

//IsProvisioned implements the EntityDefinition interface.
func (g *GroupDefinition) IsProvisioned() bool { return g.GID > 0 }

//IsProvisioned implements the EntityDefinition interface.
func (u *UserDefinition) IsProvisioned() bool { return u.UID > 0 }

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
func (g *GroupDefinition) WithSerializableState(callback func(EntityDefinition)) {
	//we don't want to serialize the `system` attribute in diffs etc.
	system := g.System
	g.System = false
	callback(g)
	g.System = system
}

//WithSerializableState implements the EntityDefinition interface.
func (u *UserDefinition) WithSerializableState(callback func(EntityDefinition)) {
	//we don't want to serialize the `system` attribute in diffs etc.
	system := u.System
	u.System = false
	callback(u)
	u.System = system
}
