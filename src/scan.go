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
func Scan() ([]Group, []User) {
	//open resource directory
	dirPath := os.Getenv("HOLO_RESOURCE_DIR")
	dir, err := os.Open(dirPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return nil, nil
	}
	fis, err := dir.Readdir(-1)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return nil, nil
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
	groupsList := make([]Group, 0, len(groups))
	for _, group := range groups {
		if group.isValid() {
			groupsList = append(groupsList, *group)
		}
	}
	sort.Sort(groupsByName(groupsList))

	usersList := make([]User, 0, len(users))
	for _, user := range users {
		if user.isValid() {
			usersList = append(usersList, *user)
		}
	}
	sort.Sort(usersByName(usersList))

	return groupsList, usersList
}

type usersByName []User

func (u usersByName) Len() int           { return len(u) }
func (u usersByName) Less(i, j int) bool { return u[i].Name < u[j].Name }
func (u usersByName) Swap(i, j int)      { u[i], u[j] = u[j], u[i] }

type groupsByName []Group

func (g groupsByName) Len() int           { return len(g) }
func (g groupsByName) Less(i, j int) bool { return g[i].Name < g[j].Name }
func (g groupsByName) Swap(i, j int)      { g[i], g[j] = g[j], g[i] }

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
