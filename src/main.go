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
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	if version := os.Getenv("HOLO_API_VERSION"); version != "3" {
		fmt.Fprintf(os.Stderr, "!! holo-users-groups plugin called with unknown HOLO_API_VERSION %s\n", version)
		os.Exit(1)
	}

	gob.Register(&GroupDefinition{})
	gob.Register(&UserDefinition{})
	gob.Register(Entity{})

	switch os.Args[1] {
	case "info":
		os.Stdout.Write([]byte("MIN_API_VERSION=3\nMAX_API_VERSION=3\n"))
	case "scan":
		executeScanCommand()
	default:
		executeNonScanCommand()
	}
}

func pathToCacheFile() string {
	return filepath.Join(os.Getenv("HOLO_CACHE_DIR"), "entities.toml")
}

func executeScanCommand() {
	//scan for entities
	entities := Scan()
	if entities == nil {
		//some fatal error occurred - it was already reported, so just exit
		os.Exit(1)
	}

	//print reports
	for _, entity := range entities {
		entity.PrintReport()
	}

	//store scan result in cache
	file, err := os.Create(pathToCacheFile())
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		os.Exit(1)
	}
	err = gob.NewEncoder(file).Encode(entities)
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		os.Exit(1)
	}
	err = file.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		os.Exit(1)
	}
}

func executeNonScanCommand() {
	//retrieve entities from cache
	file, err := os.Open(pathToCacheFile())
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		os.Exit(1)
	}
	var entities []*Entity
	err = gob.NewDecoder(file).Decode(&entities)
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		os.Exit(1)
	}
	err = file.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		os.Exit(1)
	}

	//all other actions require an entity selection
	entityID := os.Args[2]
	var selectedEntity *Entity
	for _, entity := range entities {
		if entity.Definition.EntityID() == entityID {
			selectedEntity = entity
			break
		}
	}
	if selectedEntity == nil {
		fmt.Fprintf(os.Stderr, "!! unknown entity ID \"%s\"\n", entityID)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "apply":
		err = selectedEntity.Apply(false)
	case "force-apply":
		err = selectedEntity.Apply(true)
	case "diff":
		err = PrepareDiffFor(selectedEntity.Definition, selectedEntity.IsOrphaned())
	default:
		err = fmt.Errorf("unknown command '%s'", os.Args[1])
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
	}
}
