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
func Scan() []Entity {
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
	groups := make(map[string]*Group)
	users := make(map[string]*User)
	for _, definitionPath := range paths {
		err := readDefinitionFile(definitionPath, &groups, &users)
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
		if _, ok := groups[name]; !ok {
			groups[name] = &Group{GroupDefinition: GroupDefinition{Name: name}, Orphaned: true}
		}
	}
	for _, name := range KnownUserNames() {
		if _, ok := users[name]; !ok {
			users[name] = &User{UserDefinition: UserDefinition{Name: name}, Orphaned: true}
		}
	}

	//flatten result into a list sorted by EntityID and filter invalid entities
	result := make([]Entity, 0, len(groups)+len(users))
	for _, group := range groups {
		if group.isValid() {
			result = append(result, *group)
		}
	}
	for _, user := range users {
		if user.isValid() {
			result = append(result, *user)
		}
	}
	sort.Sort(entitiesByName(result))

	return result
}

type entitiesByName []Entity

func (e entitiesByName) Len() int      { return len(e) }
func (e entitiesByName) Swap(i, j int) { e[i], e[j] = e[j], e[i] }
func (e entitiesByName) Less(i, j int) bool {
	return e[i].Definition().EntityID() < e[j].Definition().EntityID()
}

func readDefinitionFile(definitionPath string, groups *map[string]*Group, users *map[string]*User) []error {
	//unmarshal contents of definitionPath into this struct
	var contents struct {
		Group []Group
		User  []User
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

	//convert the definitions read into entities, or extend existing entities if
	//the definition is stacked on an earlier one (BUT: we only allow changes
	//that are compatible with the original definition; for example, users may
	//be extended with additional groups, but its UID may not be changed)
	for idx, group := range contents.Group {
		if group.Name == "" {
			errors = append(errors, fmt.Errorf("groups[%d] is missing required 'name' attribute", idx))
			continue
		}
		existingGroup, exists := (*groups)[group.Name]
		if exists {
			//stacked definition for this group - extend existing Group entity
			groupErrors := mergeGroupDefinition(&group, existingGroup)
			if len(groupErrors) > 0 {
				errors = append(errors, groupErrors...)
				existingGroup.setInvalid()
			}
		} else {
			//first definition for this group - create new Group entity
			copyOfGroup := group
			(*groups)[group.Name] = &copyOfGroup
			existingGroup = &copyOfGroup
		}
		existingGroup.DefinitionFiles = append(existingGroup.DefinitionFiles, definitionPath)
	}

	for idx, user := range contents.User {
		if user.Name == "" {
			errors = append(errors, fmt.Errorf("users[%d] is missing required 'name' attribute", idx))
			continue
		}
		existingUser, exists := (*users)[user.Name]
		if exists {
			//stacked definition for this user - extend existing User entity
			userErrors := mergeUserDefinition(&user, existingUser)
			if len(userErrors) > 0 {
				errors = append(errors, userErrors...)
				existingUser.setInvalid()
			}
		} else {
			//first definition for this user - create new User entity
			copyOfUser := user
			(*users)[user.Name] = &copyOfUser
			existingUser = &copyOfUser
		}
		existingUser.DefinitionFiles = append(existingUser.DefinitionFiles, definitionPath)
	}

	return errors
}

//Merges `def` into `group` if possible, returns errors if merge conflicts arise.
func mergeGroupDefinition(group *Group, existingGroup *Group) []error {
	def, errors := group.GroupDefinition.Merge(&existingGroup.GroupDefinition, MergeWhereCompatible)
	if len(errors) == 0 {
		existingGroup.GroupDefinition = *(def.(*GroupDefinition))
	}
	return errors
}

//Merges `def` into `user` if possible, returns errors if merge conflicts arise.
func mergeUserDefinition(user *User, existingUser *User) []error {
	def, errors := user.UserDefinition.Merge(&existingUser.UserDefinition, MergeWhereCompatible)
	if len(errors) == 0 {
		existingUser.UserDefinition = *(def.(*UserDefinition))
	}
	return errors
}
