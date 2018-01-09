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

package entrypoint

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
)

// Main is the main entry point, but returns the exit code rather than
// calling os.Exit().  This distinction is useful for monobinary and
// testing purposes.
func Main() (exitCode int) {
	if version := os.Getenv("HOLO_API_VERSION"); version != "3" {
		fmt.Fprintf(os.Stderr, "!! holo-users-groups plugin called with unknown HOLO_API_VERSION %s\n", version)
		return 1
	}

	if os.Getenv("HOLO_STATE_DIR") != "" {
		for _, dir := range []string{string(BaseImageDir), string(ProvisionedImageDir)} {
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
				os.Exit(1)
			}
		}
	}

	gob.Register(&GroupDefinition{})
	gob.Register(&UserDefinition{})
	gob.Register(Entity{})

	var err error
	switch os.Args[1] {
	case "info":
		os.Stdout.Write([]byte("MIN_API_VERSION=3\nMAX_API_VERSION=3\n"))
	case "scan":
		err = executeScanCommand()
	default:
		err = executeNonScanCommand()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		return 1
	}

	return 0
}

func pathToCacheFile() string {
	return filepath.Join(os.Getenv("HOLO_CACHE_DIR"), "entities.toml")
}

func executeScanCommand() error {
	//scan for entities
	entities, errors := Scan()
	for _, err := range errors {
		fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
	}
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
		return err
	}
	err = gob.NewEncoder(file).Encode(entities)
	if err != nil {
		return err
	}
	return file.Close()
}

func executeNonScanCommand() error {
	//retrieve entities from cache
	file, err := os.Open(pathToCacheFile())
	if err != nil {
		return err
	}
	var entities []*Entity
	err = gob.NewDecoder(file).Decode(&entities)
	if err != nil {
		return err
	}
	err = file.Close()
	if err != nil {
		return err
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
		return fmt.Errorf("unknown entity ID \"%s\"", entityID)
	}

	switch os.Args[1] {
	case "apply":
		return selectedEntity.Apply(false)
	case "force-apply":
		return selectedEntity.Apply(true)
	case "diff":
		return selectedEntity.PrepareDiff()
	default:
		return fmt.Errorf("unknown command '%s'", os.Args[1])
	}
}
