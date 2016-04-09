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

	"../localdeps/github.com/BurntSushi/toml"
)

func main() {
	if version := os.Getenv("HOLO_API_VERSION"); version != "3" {
		fmt.Fprintf(os.Stderr, "!! holo-users-groups plugin called with unknown HOLO_API_VERSION %s\n", version)
	}

	switch os.Args[1] {
	case "info":
		os.Stdout.Write([]byte("MIN_API_VERSION=3\nMAX_API_VERSION=3\n"))
	case "scan":
		executeScanCommand()
	default:
		executeNonScanCommand()
	}
}

type cache struct {
	Groups []Group
	Users  []User
}

func pathToCacheFile() string {
	return filepath.Join(os.Getenv("HOLO_CACHE_DIR"), "entities.toml")
}

func executeScanCommand() {
	//scan for entities
	groups, users := Scan()
	if groups == nil && users == nil {
		//some fatal error occurred - it was already reported, so just exit
		os.Exit(1)
	}

	//print reports
	for _, group := range groups {
		group.PrintReport()
	}
	for _, user := range users {
		user.PrintReport()
	}

	//store scan result in cache
	file, err := os.Create(pathToCacheFile())
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		os.Exit(1)
	}
	err = toml.NewEncoder(file).Encode(&cache{groups, users})
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		os.Exit(1)
	}
	file.Close()
}

func executeNonScanCommand() {
	//retrieve entities from cache
	blob, err := ioutil.ReadFile(pathToCacheFile())
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		os.Exit(1)
	}
	var cacheData cache
	_, err = toml.Decode(string(blob), &cacheData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		os.Exit(1)
	}

	//all other actions require an entity selection
	entityID := os.Args[2]
	var selectedEntity Entity
	for _, group := range cacheData.Groups {
		if group.EntityID() == entityID {
			selectedEntity = group
			break
		}
	}
	for _, user := range cacheData.Users {
		if user.EntityID() == entityID {
			selectedEntity = user
			break
		}
	}
	if selectedEntity == nil {
		fmt.Fprintf(os.Stderr, "!! unknown entity ID \"%s\"\n", entityID)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "apply":
		applyEntity(selectedEntity, false)
	case "force-apply":
		applyEntity(selectedEntity, true)
	case "diff":
		expectedStateFile, actualStateFile, err := selectedEntity.PrepareDiff()
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		}
		out := fmt.Sprintf("%s\000%s\000", expectedStateFile, actualStateFile)
		_, err = os.NewFile(3, "file descriptor 3").Write([]byte(out))
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		}
	}
}

func applyEntity(entity Entity, withForce bool) {
	entityHasChanged := entity.Apply(withForce)
	if !entityHasChanged {
		_, err := os.NewFile(3, "file descriptor 3").Write([]byte("not changed\n"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		}
	}
}
