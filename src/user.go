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
	"os"
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

//User represents a UNIX user account (as registered in /etc/passwd). It
//implements the Entity interface and is handled accordingly.
type User struct {
	UserDefinition
	DefinitionFiles []string //paths to the files defining this entity

	Orphaned bool //whether entity definition have been deleted (default: false)
	broken   bool //whether the entity definition is invalid (default: false)
}

//Definition implements the Entity interface.
func (u User) Definition() EntityDefinition { return &u.UserDefinition }

//IsOrphaned implements the Entity interface.
func (u User) IsOrphaned() bool { return u.Orphaned }

//isValid is used inside the scanning algorithm to filter entities with
//broken definitions, which shall be skipped during `holo apply`.
func (u *User) isValid() bool { return !u.broken }

//setInvalid is used inside the scnaning algorithm to mark entities with
//broken definitions, which shall be skipped during `holo apply`.
func (u *User) setInvalid() { u.broken = true }

//PrintReport implements the Entity interface for User.
func (u User) PrintReport() {
	fmt.Printf("ENTITY: %s\n", u.EntityID())
	if u.Orphaned {
		fmt.Println("ACTION: Scrubbing (all definition files have been deleted)")
	} else {
		for _, defFile := range u.DefinitionFiles {
			fmt.Printf("found in: %s\n", defFile)
			fmt.Printf("SOURCE: %s\n", defFile)
		}
		if attributes := u.Attributes(); attributes != "" {
			fmt.Printf("with: %s\n", attributes)
		}
	}
}

type userDiff struct {
	field    string
	actual   string
	expected string
}

//Apply performs the complete application algorithm for the given Entity.
//If the user does not exist yet, it is created. If it does exist, but some
//attributes do not match, it will be updated, but only if withForce is given.
func (u User) Apply(withForce bool) (entityHasChanged bool) {
	//special handling for orphaned users
	if u.Orphaned {
		return u.applyOrphaned(withForce)
	}

	//check if we have that group already
	actualDef, err := u.GetProvisionedState()
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! Cannot read user database: %s\n", err.Error())
		return false
	}

	//check if the actual properties diverge from our definition
	if actualDef != nil {
		actualStr, err := SerializeDefinition(actualDef)
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		}
		expectedDef, _ := u.UserDefinition.Merge(actualDef, MergeEmptyOnly)
		expectedStr, err := SerializeDefinition(expectedDef)
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		}

		if string(actualStr) != string(expectedStr) {
			if withForce {
				err := u.UserDefinition.Apply(actualDef)
				if err != nil {
					fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
					return false
				}
				return true
			}
			_, err := os.NewFile(3, "file descriptor 3").Write([]byte("requires --force to overwrite\n"))
			if err != nil {
				fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
			}
		}
		return false
	}

	//create the user if it does not exist
	err = u.UserDefinition.Apply(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		return false
	}
	return true
}

func (u User) applyOrphaned(withForce bool) (entityHasChanged bool) {
	if !withForce {
		fmt.Fprintf(os.Stderr, "!! Won't do this without --force.\n")
		fmt.Fprintf(os.Stderr, ">> Call `holo apply --force user:%s` to delete this user.\n", u.Name)
		fmt.Fprintf(os.Stderr, ">> Or remove the user name from %s to keep the user.\n", RegistryPath())
		return false
	}

	//call userdel and remove user from our registry
	err := u.UserDefinition.Cleanup()
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		return false
	}
	return true
}
