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
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
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

//Apply implements the EntityDefinition interface.
func (u *UserDefinition) Apply(provisioned EntityDefinition) error {
	//assemble arguments
	var args []string
	if provisioned == nil && u.System {
		args = append(args, "--system")
	}
	if u.UID > 0 {
		args = append(args, "--uid", strconv.Itoa(u.UID))
	}
	if u.Comment != "" {
		args = append(args, "--comment", u.Comment)
	}
	if u.Home != "" {
		//yay for consistency
		if provisioned == nil {
			args = append(args, "--home-dir", u.Home)
		} else {
			args = append(args, "--home", u.Home)
		}
	}
	if u.Group != "" {
		args = append(args, "--gid", u.Group)
	}
	if len(u.Groups) > 0 {
		args = append(args, "--groups", strings.Join(u.Groups, ","))
	}
	if u.Shell != "" {
		args = append(args, "--shell", u.Shell)
	}
	args = append(args, u.Name)

	//call useradd/usermod
	command := "usermod"
	if provisioned == nil {
		command = "useradd"
	}
	err := ExecProgramOrMock(command, args...)
	if err != nil {
		return err
	}
	return AddProvisionedUser(u.Name)
}

//Cleanup implements the EntityDefinition interface.
func (u *UserDefinition) Cleanup() error {
	err := ExecProgramOrMock("userdel", u.Name)
	if err != nil {
		return err
	}
	return RemoveProvisionedUser(u.Name)
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

//GetProvisionedState implements the EntityDefinition interface.
func (u *UserDefinition) GetProvisionedState() (EntityDefinition, error) {
	passwdFile := GetPath("etc/passwd")
	groupFile := GetPath("etc/group")

	//fetch entry from /etc/passwd
	fields, err := Getent(passwdFile, func(fields []string) bool { return fields[0] == u.Name })
	if err != nil {
		return nil, err
	}
	//is there such a user?
	if fields == nil {
		return nil, nil
	}
	//is the passwd entry intact?
	if len(fields) < 4 {
		return nil, errors.New("invalid entry in /etc/passwd (not enough fields)")
	}

	//read fields in passwd entry
	actualUID, err := strconv.Atoi(fields[2])
	if err != nil {
		return nil, err
	}

	//fetch entry for login group from /etc/group (to resolve actualGID into a
	//group name)
	actualGIDString := fields[3]
	groupFields, err := Getent(groupFile, func(fields []string) bool {
		if len(fields) <= 2 {
			return false
		}
		return fields[2] == actualGIDString
	})
	if err != nil {
		return nil, err
	}
	if groupFields == nil {
		return nil, errors.New("invalid entry in /etc/passwd (login group does not exist)")
	}
	groupName := groupFields[0]

	//check /etc/group for the supplementary group memberships of this user
	var groupNames []string
	_, err = Getent(groupFile, func(fields []string) bool {
		if len(fields) <= 3 {
			return false
		}
		//collect groups that contain this user
		users := strings.Split(fields[3], ",")
		for _, user := range users {
			if user == u.Name {
				groupNames = append(groupNames, fields[0])
			}
		}
		//keep going
		return false
	})
	if err != nil {
		return nil, err
	}

	//make sure that the groups list is always sorted (esp. for reproducible test output)
	sort.Strings(groupNames)

	return &UserDefinition{
		Name:    fields[0],
		Comment: fields[4],
		UID:     actualUID,
		Home:    fields[5],
		Group:   groupName,
		Groups:  groupNames,
		Shell:   fields[6],
	}, nil
}
