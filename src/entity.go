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
	"fmt"
	"os"
)

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
