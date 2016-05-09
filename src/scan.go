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
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"../localdeps/github.com/BurntSushi/toml"
)

//Scan returns a slice of all the defined entities.
func Scan() ([]*Entity, []error) {
	//call into migration code
	err := migrateOldRegistry()
	if err != nil {
		return nil, []error{err}
	}

	//open resource directory
	dirPath := os.Getenv("HOLO_RESOURCE_DIR")
	dir, err := os.Open(dirPath)
	if err != nil {
		return nil, []error{err}
	}
	fis, err := dir.Readdir(-1)
	if err != nil {
		return nil, []error{err}
	}

	//find entity definitions
	var paths []string
	for _, fi := range fis {
		if fi.Mode().IsRegular() && strings.HasSuffix(fi.Name(), ".toml") {
			paths = append(paths, filepath.Join(dirPath, fi.Name()))
		}
	}
	sort.Strings(paths)

	//parse entity definitions
	entities := make(map[string]*Entity)
	var errors []error
	for _, definitionPath := range paths {
		err := readDefinitionFile(definitionPath, &entities)
		if err != nil {
			errors = append(errors, err)
		}
	}

	//find orphaned entities (invalid entities are considered "existing" here,
	//so that we don't remove entities that are still needed just because their
	//definition file is broken)
	ids, err := ProvisionedEntityIDs()
	if err != nil {
		return nil, []error{err}
	}
	for _, id := range ids {
		var def EntityDefinition
		switch {
		case strings.HasPrefix(id, "group:"):
			def = &GroupDefinition{Name: strings.TrimPrefix(id, "group:")}
		case strings.HasPrefix(id, "user:"):
			def = &UserDefinition{Name: strings.TrimPrefix(id, "user:")}
		}
		if _, ok := entities[def.EntityID()]; !ok {
			entities[def.EntityID()] = &Entity{Definition: def}
		}
	}

	//flatten result into a list sorted by EntityID and filter invalid entities
	result := make([]*Entity, 0, len(entities))
	for _, entity := range entities {
		if !entity.IsBroken {
			result = append(result, entity)
		}
	}
	sort.Sort(entitiesByName(result))

	return result, errors
}

type entitiesByName []*Entity

func (e entitiesByName) Len() int      { return len(e) }
func (e entitiesByName) Swap(i, j int) { e[i], e[j] = e[j], e[i] }
func (e entitiesByName) Less(i, j int) bool {
	return e[i].Definition.EntityID() < e[j].Definition.EntityID()
}

//FileInvalidError contains the set of errors that were encountered
//while parsing a file.
type FileInvalidError struct {
	path   string
	errors []error
}

//Error implements the error interface.
func (e *FileInvalidError) Error() string {
	str := fmt.Sprintf("File %s is invalid:", e.path)
	for _, suberr := range e.errors {
		str += fmt.Sprintf("\n>> %s", suberr.Error())
	}
	return str
}

func readDefinitionFile(definitionPath string, entities *map[string]*Entity) error {
	//unmarshal contents of definitionPath into this struct
	var contents struct {
		Group []*GroupDefinition
		User  []*UserDefinition
	}
	blob, err := ioutil.ReadFile(definitionPath)
	if err != nil {
		return &FileInvalidError{definitionPath, []error{err}}
	}
	_, err = toml.Decode(string(blob), &contents)
	if err != nil {
		return &FileInvalidError{definitionPath, []error{err}}
	}

	//when checking the entity definitions, report all errors at once
	var errors []error

	//collect the definitions in this file
	defs := make([]EntityDefinition, 0, len(contents.Group)+len(contents.User))
	for idx, group := range contents.Group {
		if group.Name == "" {
			errors = append(errors, fmt.Errorf("groups[%d] is missing required 'name' attribute", idx))
		} else {
			defs = append(defs, group)
		}
	}
	for idx, user := range contents.User {
		if user.Name == "" {
			errors = append(errors, fmt.Errorf("users[%d] is missing required 'name' attribute", idx))
			continue
		} else {
			defs = append(defs, user)
		}
	}

	//merge definitions into existing entities where appropriate
	for _, def := range defs {
		id := def.EntityID()
		entity, exists := (*entities)[id]
		if exists {
			//stacked definition for this entity -> merge into existing entity
			mergedDef, mergeErrors := def.Merge(entity.Definition, MergeWhereCompatible)
			if len(mergeErrors) == 0 {
				entity.Definition = mergedDef
			} else {
				errors = append(errors, mergeErrors...)
				entity.IsBroken = true
			}
		} else {
			//first definition for this entity -> wrap into an Entity
			entity = &Entity{Definition: def}
			(*entities)[id] = entity
		}
		entity.DefinitionFiles = append(entity.DefinitionFiles, definitionPath)
	}

	if len(errors) > 0 {
		return &FileInvalidError{definitionPath, errors}
	}
	return nil
}

//Migration path for the old registry at `/var/lib/holo/users-groups/state.toml`.
func migrateOldRegistry() error {
	//read state.toml (if it exists)
	statePath := filepath.Join(os.Getenv("HOLO_STATE_DIR"), "state.toml")
	blob, err := ioutil.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			//file is gone already, no upgrade necessary
			return nil
		}
		return err
	}

	//parse state.toml
	var state struct {
		ProvisionedGroups []string
		ProvisionedUsers  []string
	}
	_, err = toml.Decode(string(blob), &state)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, ">> Migrating %s...\n", statePath)
	fmt.Fprintf(os.Stderr, "!! This might require manual intervention! Please find instructions at:\n")
	fmt.Fprintf(os.Stderr, "!!   <https://github.com/holocm/holo-users-groups/wiki/Migrating-to-v2.0>\n")

	//migrate each provisioned entity
	for _, groupName := range state.ProvisionedGroups {
		err = migrateEntity(&GroupDefinition{Name: groupName})
		if err != nil {
			return err
		}
	}
	for _, userName := range state.ProvisionedUsers {
		err = migrateEntity(&UserDefinition{Name: userName})
		if err != nil {
			return err
		}
	}

	//all went well - drop state.toml
	fmt.Fprintf(os.Stderr, ">> All entities migrated. Removing %s...\n", statePath)
	return os.Remove(statePath)
}

func migrateEntity(emptyBaseImage EntityDefinition) error {
	//don't steam-roll over existing base images
	baseImage, err := BaseImageDir.LoadImageFor(emptyBaseImage)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if baseImage != nil {
		fmt.Fprintf(os.Stderr, ">> Skipping %s (found existing base image)\n", emptyBaseImage.EntityID())
		return nil
	}

	//write the empty base image
	fmt.Fprintf(os.Stderr, ">> Writing empty base image for %s\n", emptyBaseImage.EntityID())
	return BaseImageDir.SaveImage(emptyBaseImage)
}
