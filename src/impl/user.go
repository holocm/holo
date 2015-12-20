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
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

//User represents a UNIX user account (as registered in /etc/passwd). It
//implements the Entity interface and is handled accordingly.
type User struct {
	Name            string   //the user name (the first field in /etc/passwd)
	Comment         string   //the full name (sometimes also called "comment"; the fifth field in /etc/passwd)
	UID             int      //the user ID (the third field in /etc/passwd), or 0 if no specific UID is enforced
	System          bool     //whether the group is a system group (this influences the GID selection if gid = 0)
	HomeDirectory   string   `toml:"home"` //path to the user's home directory (or empty to use the default)
	Group           string   //the name of the user's initial login group (or empty to use the default)
	Groups          []string //the names of supplementary groups which the user is also a member of
	Shell           string   //path to the user's login shell (or empty to use the default)
	DefinitionFiles []string //paths to the files defining this entity

	broken bool //whether the entity definition is invalid (default: false)
}

//isValid is used inside the scanning algorithm to filter entities with
//broken definitions, which shall be skipped during `holo apply`.
func (u *User) isValid() bool { return !u.broken }

//setInvalid is used inside the scnaning algorithm to mark entities with
//broken definitions, which shall be skipped during `holo apply`.
func (u *User) setInvalid() { u.broken = true }

//EntityID implements the Entity interface for User.
func (u User) EntityID() string { return "user:" + u.Name }

//PrintReport implements the Entity interface for User.
func (u User) PrintReport() {
	fmt.Printf("ENTITY: %s\n", u.EntityID())
	for _, defFile := range u.DefinitionFiles {
		fmt.Printf("found in: %s\n", defFile)
		fmt.Printf("SOURCE: %s\n", defFile)
	}
	if attributes := u.attributes(); attributes != "" {
		fmt.Printf("with: %s\n", attributes)
	}
}

func (u User) attributes() string {
	attrs := []string{}
	if u.System {
		attrs = append(attrs, "type: system")
	}
	if u.UID > 0 {
		attrs = append(attrs, fmt.Sprintf("UID: %d", u.UID))
	}
	if u.HomeDirectory != "" {
		attrs = append(attrs, "home: "+u.HomeDirectory)
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

type userDiff struct {
	field    string
	actual   string
	expected string
}

//Apply performs the complete application algorithm for the given Entity.
//If the user does not exist yet, it is created. If it does exist, but some
//attributes do not match, it will be updated, but only if withForce is given.
func (u User) Apply(withForce bool) (entityHasChanged bool) {
	//check if we have that group already
	userExists, actualUser, err := u.checkExists()
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! Cannot read user database: %s\n", err.Error())
		return false
	}

	//check if the actual properties diverge from our definition
	if userExists {
		differences := []userDiff{}
		if u.Comment != "" && u.Comment != actualUser.Comment {
			differences = append(differences, userDiff{"comment", actualUser.Comment, u.Comment})
		}
		if u.UID > 0 && u.UID != actualUser.UID {
			differences = append(differences, userDiff{"UID", strconv.Itoa(actualUser.UID), strconv.Itoa(u.UID)})
		}
		if u.HomeDirectory != "" && u.HomeDirectory != actualUser.HomeDirectory {
			differences = append(differences, userDiff{"home directory", actualUser.HomeDirectory, u.HomeDirectory})
		}
		if u.Shell != "" && u.Shell != actualUser.Shell {
			differences = append(differences, userDiff{"login shell", actualUser.Shell, u.Shell})
		}
		if u.Group != "" && u.Group != actualUser.Group {
			differences = append(differences, userDiff{"login group", actualUser.Group, u.Group})
		}
		//to detect changes in u.Groups <-> actualUser.Groups, we sort and join both slices
		expectedGroupsSlice := append([]string(nil), u.Groups...) //take a copy of the slice
		sort.Strings(expectedGroupsSlice)
		expectedGroups := strings.Join(expectedGroupsSlice, ", ")
		actualGroupsSlice := append([]string(nil), actualUser.Groups...)
		sort.Strings(actualGroupsSlice)
		actualGroups := strings.Join(actualGroupsSlice, ", ")
		if expectedGroups != actualGroups {
			differences = append(differences, userDiff{"groups", actualGroups, expectedGroups})
		}

		if len(differences) != 0 {
			if withForce {
				for _, diff := range differences {
					fmt.Printf(">> fixing %s (was: %s)\n", diff.field, diff.actual)
				}
				err := u.callUsermod()
				if err != nil {
					fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
					return false
				}
				return true
			}
			for _, diff := range differences {
				fmt.Fprintf(os.Stderr, "!! User has %s: %s, expected %s (use --force to overwrite)\n", diff.field, diff.actual, diff.expected)
			}
		}
		return false
	}

	//create the user if it does not exist
	err = u.callUseradd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		return false
	}
	return true
}

//checkExists checks if the user exists in /etc/passwd. If it does, its actual
//properties will be returned in the second return argument.
func (u User) checkExists() (exists bool, currentUser *User, e error) {
	passwdFile := GetPath("etc/passwd")
	groupFile := GetPath("etc/group")

	//fetch entry from /etc/passwd
	fields, err := Getent(passwdFile, func(fields []string) bool { return fields[0] == u.Name })
	if err != nil {
		return false, nil, err
	}
	//is there such a user?
	if fields == nil {
		return false, nil, nil
	}
	//is the passwd entry intact?
	if len(fields) < 4 {
		return true, nil, errors.New("invalid entry in /etc/passwd (not enough fields)")
	}

	//read fields in passwd entry
	actualUID, err := strconv.Atoi(fields[2])
	if err != nil {
		return true, nil, err
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
		return true, nil, err
	}
	if groupFields == nil {
		return true, nil, errors.New("invalid entry in /etc/passwd (login group does not exist)")
	}
	groupName := groupFields[0]

	//check /etc/group for the supplementary group memberships of this user
	groupNames := []string{}
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
		return true, nil, err
	}

	return true, &User{
		//NOTE: Some fields (name, system, definitionFile) are not set because
		//they are not relevant for the algorithm.
		Comment:       fields[4],
		UID:           actualUID,
		HomeDirectory: fields[5],
		Group:         groupName,
		Groups:        groupNames,
		Shell:         fields[6],
	}, nil
}

func (u User) callUseradd() error {
	//assemble arguments for useradd call
	args := []string{}
	if u.System {
		args = append(args, "--system")
	}
	if u.UID > 0 {
		args = append(args, "--uid", strconv.Itoa(u.UID))
	}
	if u.Comment != "" {
		args = append(args, "--comment", u.Comment)
	}
	if u.HomeDirectory != "" {
		args = append(args, "--home-dir", u.HomeDirectory)
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

	//call useradd
	err := ExecProgramOrMock("useradd", args...)
	if err != nil {
		return err
	}
	return AddProvisionedUser(u.Name)
}

func (u User) callUsermod() error {
	//assemble arguments for usermod call
	args := []string{}
	if u.UID > 0 {
		args = append(args, "--uid", strconv.Itoa(u.UID))
	}
	if u.Comment != "" {
		args = append(args, "--comment", u.Comment)
	}
	if u.HomeDirectory != "" {
		args = append(args, "--home", u.HomeDirectory)
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

	//call usermod
	err := ExecProgramOrMock("usermod", args...)
	if err != nil {
		return err
	}
	return AddProvisionedUser(u.Name)
}
