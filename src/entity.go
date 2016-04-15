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
	"os"

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
	//there, nil is returned.
	GetProvisionedState() (EntityDefinition, error)
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

//Entity contains attributes and logic that are shared between entity types.
type Entity struct {
	Definition      EntityDefinition
	DefinitionFiles []string //paths to the files defining this entity
	IsBroken        bool     //whether any of these are invalid (default: false)
}

//IsOrphaned returns whether all definitions for this entity have been deleted.
func (e *Entity) IsOrphaned() bool {
	return len(e.DefinitionFiles) == 0
}

//PrintReport prints the scan report for this entity on stdout.
func (e *Entity) PrintReport() {
	fmt.Printf("ENTITY: %s\n", e.Definition.EntityID())
	if e.IsOrphaned() {
		fmt.Println("ACTION: Scrubbing (all definition files have been deleted)")
	} else {
		for _, defFile := range e.DefinitionFiles {
			fmt.Printf("found in: %s\n", defFile)
			fmt.Printf("SOURCE: %s\n", defFile)
		}
		if attributes := e.Definition.Attributes(); attributes != "" {
			fmt.Printf("with: %s\n", attributes)
		}
	}
}

//Apply performs the complete application algorithm for the given Entity.
//If the entity does not exist yet, it is created. If it does exist, but some
//attributes do not match, it will be updated, but only if withForce is given.
func (e *Entity) Apply(withForce bool) (entityHasChanged bool) {
	//special handling for orphaned entities
	if e.IsOrphaned() {
		return e.applyOrphaned(withForce)
	}

	//check if this entity exists already
	def := e.Definition
	actualDef, err := def.GetProvisionedState()
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! Cannot read %s database: %s\n", def.TypeName(), err.Error())
		return false
	}

	//check if the actual properties diverge from our definition
	if actualDef != nil {
		actualStr, err := SerializeDefinition(actualDef)
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
			return false
		}
		expectedDef, _ := def.Merge(actualDef, MergeEmptyOnly)
		expectedStr, err := SerializeDefinition(expectedDef)
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
			return false
		}

		if string(actualStr) != string(expectedStr) {
			if withForce {
				err := def.Apply(actualDef)
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

	//create the entity if it does not exist
	err = def.Apply(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		return false
	}
	return true
}

func (e *Entity) applyOrphaned(withForce bool) (entityHasChanged bool) {
	def := e.Definition

	if !withForce {
		typeName := def.TypeName()
		entityID := def.EntityID()
		fmt.Fprintf(os.Stderr, "!! Won't do this without --force.\n")
		fmt.Fprintf(os.Stderr, ">> Call `holo apply --force %s` to delete this %s.\n", entityID, typeName)
		fmt.Fprintf(os.Stderr, ">> Or remove the %s name from %s to keep the %s.\n", typeName, RegistryPath(), typeName)
		return false
	}

	//call groupdel and remove group from our registry
	err := def.Cleanup()
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		return false
	}
	return true
}
