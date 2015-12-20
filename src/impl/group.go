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
	"strconv"
	"strings"
)

//Group represents a UNIX group (as registered in /etc/group). It implements
//the Entity interface and is handled accordingly.
type Group struct {
	Name            string   //the group name (the first field in /etc/group)
	GID             int      //the GID (the third field in /etc/group), or 0 if no specific GID is enforced
	System          bool     //whether the group is a system group (this influences the GID selection if GID = 0)
	DefinitionFiles []string //paths to the files defining this entity

	Orphaned bool //whether entity definition have been deleted (default: false)
	broken   bool //whether the entity definition is invalid (default: false)
}

//isValid is used inside the scanning algorithm to filter entities with
//broken definitions, which shall be skipped during `holo apply`.
func (g *Group) isValid() bool { return !g.broken }

//setInvalid is used inside the scnaning algorithm to mark entities with
//broken definitions, which shall be skipped during `holo apply`.
func (g *Group) setInvalid() { g.broken = true }

//EntityID implements the Entity interface for Group.
func (g Group) EntityID() string { return "group:" + g.Name }

//PrintReport implements the Entity interface for Group.
func (g Group) PrintReport() {
	fmt.Printf("ENTITY: %s\n", g.EntityID())
	if g.Orphaned {
		fmt.Println("ACTION: Scrubbing (all definition files have been deleted)")
	} else {
		for _, defFile := range g.DefinitionFiles {
			fmt.Printf("found in: %s\n", defFile)
			fmt.Printf("SOURCE: %s\n", defFile)
		}
		if attributes := g.attributes(); attributes != "" {
			fmt.Printf("with: %s\n", attributes)
		}
	}
}

func (g Group) attributes() string {
	attrs := []string{}
	if g.System {
		attrs = append(attrs, "type: system")
	}
	if g.GID > 0 {
		attrs = append(attrs, fmt.Sprintf("GID: %d", g.GID))
	}
	return strings.Join(attrs, ", ")
}

type groupDiff struct {
	field    string
	actual   string
	expected string
}

//Apply performs the complete application algorithm for the given Entity.
//If the group does not exist yet, it is created. If it does exist, but some
//attributes do not match, it will be updated, but only if withForce is given.
func (g Group) Apply(withForce bool) (entityHasChanged bool) {
	//special handling for orphaned groups
	if g.Orphaned {
		return g.applyOrphaned(withForce)
	}

	//check if we have that group already
	groupExists, actualGid, err := g.checkExists()
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! Cannot read group database: %s\n", err.Error())
		return false
	}

	//check if the actual properties diverge from our definition
	if groupExists {
		differences := []groupDiff{}
		if g.GID > 0 && g.GID != actualGid {
			differences = append(differences, groupDiff{"GID", strconv.Itoa(actualGid), strconv.Itoa(g.GID)})
		}

		if len(differences) != 0 {
			if withForce {
				for _, diff := range differences {
					fmt.Printf(">> fixing %s (was: %s)\n", diff.field, diff.actual)
				}
				err := g.callGroupmod()
				if err != nil {
					fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
					return false
				}
				return true
			}
			for _, diff := range differences {
				fmt.Fprintf(os.Stderr, "!! Group has %s: %s, expected %s (use --force to overwrite)\n", diff.field, diff.actual, diff.expected)
			}
		}
		return false
	}

	//create the group if it does not exist
	err = g.callGroupadd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		return false
	}
	return true
}

func (g Group) applyOrphaned(withForce bool) (entityHasChanged bool) {
	if !withForce {
		fmt.Fprintf(os.Stderr, "!! Won't do this without --force.\n")
		fmt.Fprintf(os.Stderr, ">> Call `holo apply --force group:%s` to delete this group.\n", g.Name)
		fmt.Fprintf(os.Stderr, ">> Or remove the group name from %s to keep the group.\n", RegistryPath())
		return false
	}

	//call groupdel and remove group from our registry
	err := ExecProgramOrMock("groupdel", g.Name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		return false
	}
	err = RemoveProvisionedGroup(g.Name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
	}
	return true
}

func (g Group) checkExists() (exists bool, gid int, e error) {
	groupFile := GetPath("etc/group")

	//fetch entry from /etc/group
	fields, err := Getent(groupFile, func(fields []string) bool { return fields[0] == g.Name })
	if err != nil {
		return false, 0, err
	}
	//is there such a group?
	if fields == nil {
		return false, 0, nil
	}
	//is the group entry intact?
	if len(fields) < 4 {
		return true, 0, errors.New("invalid entry in /etc/group (not enough fields)")
	}

	//read fields in entry
	actualGid, err := strconv.Atoi(fields[2])
	return true, actualGid, err
}

func (g Group) callGroupadd() error {
	//assemble arguments for groupadd call
	args := []string{}
	if g.System {
		args = append(args, "--system")
	}
	if g.GID > 0 {
		args = append(args, "--gid", strconv.Itoa(g.GID))
	}
	args = append(args, g.Name)

	//call groupadd
	err := ExecProgramOrMock("groupadd", args...)
	if err != nil {
		return err
	}
	return AddProvisionedGroup(g.Name)
}

func (g Group) callGroupmod() error {
	//assemble arguments for groupmod call
	args := []string{}
	if g.GID > 0 {
		args = append(args, "--gid", strconv.Itoa(g.GID))
	}
	args = append(args, g.Name)

	//call groupmod
	err := ExecProgramOrMock("groupmod", args...)
	if err != nil {
		return err
	}
	return AddProvisionedGroup(g.Name)
}
