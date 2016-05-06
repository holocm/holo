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
func (e *Entity) Apply(withForce bool) error {
	def := e.Definition

	//check if this entity exists already
	actualDef, err := def.GetProvisionedState()
	if err != nil {
		return fmt.Errorf("Cannot read %s database: %s\n", def.TypeName(), err.Error())
	}

	//special handling for orphaned entities
	if e.IsOrphaned() {
		preImage, err := LoadPreImageFor(e.Definition)
		if err != nil {
			return err
		}

		if preImage.IsProvisioned() {
			err = preImage.Apply(actualDef)
		} else {
			err = def.Cleanup()
		}
		if err != nil {
			return err
		}

		return DeletePreImageFor(def)
	}

	//load pre-image
	preImage, err := LoadPreImageFor(e.Definition)
	if err != nil {
		if os.IsNotExist(err) {
			//write pre-image on first `apply`
			err = SavePreImage(actualDef)
			if err != nil {
				return err
			}
			preImage = actualDef
		} else {
			return err
		}
	}

	//check for manual changes
	desiredState, _ := e.Definition.Merge(preImage, MergeWhereCompatible)
	desiredStr, err := SerializeDefinition(desiredState)
	if err != nil {
		return err
	}
	compatibleState, _ := actualDef.Merge(e.Definition, MergeEmptyOnly)
	compatibleStr, err := SerializeDefinition(compatibleState)
	if err != nil {
		return err
	}
	if string(desiredStr) != string(compatibleStr) {
		PrintCommandMessage("requires --force to overwrite\n")
		return nil
	}

	//check if changes are necessary
	actualStr, err := SerializeDefinition(actualDef)
	if err != nil {
		return err
	}
	if string(desiredStr) == string(actualStr) {
		PrintCommandMessage("not changed\n")
		return nil
	}
	return desiredState.Apply(actualDef)
}

//PrintCommandMessage formats and prints a message on file descriptor 3.
func PrintCommandMessage(msg string, arguments ...interface{}) {
	if len(arguments) > 0 {
		msg = fmt.Sprintf(msg, arguments...)
	}
	_, err := os.NewFile(3, "file descriptor 3").Write([]byte(msg))
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
	}
}
