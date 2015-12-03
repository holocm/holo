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

	"./impl"
	"./internal/toml"
)

func main() {
	if version := os.Getenv("HOLO_API_VERSION"); version != "1" {
		fmt.Fprintf(os.Stderr, "!! holo-users-groups plugin called with unknown HOLO_API_VERSION %s\n", version)
	}

	if os.Args[1] == "scan" {
		executeScanCommand()
	} else {
		executeNonScanCommand()
	}
}

type cache struct {
	Groups []impl.Group
	Users  []impl.User
}

func pathToCacheFile() string {
	return filepath.Join(os.Getenv("HOLO_CACHE_DIR"), "entities.toml")
}

func executeScanCommand() {
	//scan for entities
	groups, users := impl.Scan()
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
	var selectedEntity impl.Entity
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
		output, err := selectedEntity.RenderDiff()
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		}
		os.Stdout.Write(output)
	}
}

func applyEntity(entity impl.Entity, withForce bool) {
	entityHasChanged := entity.Apply(withForce)
	if !entityHasChanged {
		_, err := os.NewFile(3, "file descriptor 3").Write([]byte("not changed\n"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		}
	}
}
