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

//Scan returns a slice of all the defined entities. If an error is encountered
//during the scan, it will be reported on stderr, and nil is returned.
func Scan() []*Entity {
	//open resource directory
	dirPath := os.Getenv("HOLO_RESOURCE_DIR")
	dir, err := os.Open(dirPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return nil
	}
	fis, err := dir.Readdir(-1)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return nil
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
	for _, definitionPath := range paths {
		err := readDefinitionFile(definitionPath, &entities)
		if len(err) > 0 {
			fmt.Fprintf(os.Stderr, "!! File %s is invalid:\n", definitionPath)
			for _, suberr := range err {
				fmt.Fprintf(os.Stderr, ">> %s\n", suberr.Error())
			}
		}
	}

	//find orphaned entities (invalid entities are considered "existing" here,
	//so that we don't remove entities that are still needed just because their
	//definition file is broken)
	for _, name := range KnownGroupNames() {
		def := &GroupDefinition{Name: name}
		if _, ok := entities[def.EntityID()]; !ok {
			entities[def.EntityID()] = &Entity{Definition: def}
		}
	}
	for _, name := range KnownUserNames() {
		def := &UserDefinition{Name: name}
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

	return result
}

type entitiesByName []*Entity

func (e entitiesByName) Len() int      { return len(e) }
func (e entitiesByName) Swap(i, j int) { e[i], e[j] = e[j], e[i] }
func (e entitiesByName) Less(i, j int) bool {
	return e[i].Definition.EntityID() < e[j].Definition.EntityID()
}

func readDefinitionFile(definitionPath string, entities *map[string]*Entity) []error {
	//unmarshal contents of definitionPath into this struct
	var contents struct {
		Group []*GroupDefinition
		User  []*UserDefinition
	}
	blob, err := ioutil.ReadFile(definitionPath)
	if err != nil {
		return []error{err}
	}
	_, err = toml.Decode(string(blob), &contents)
	if err != nil {
		return []error{err}
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

	return errors
}
