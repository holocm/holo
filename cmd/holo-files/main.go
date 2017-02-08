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
	"os"

	"github.com/holocm/holo/lib/holo-files/common"
	"github.com/holocm/holo/lib/holo-files/impl"
)

func main() {
	os.Exit(Main())
}

// Main is the main entry point, but returns the exit code rather than
// calling os.Exit().  This distinction is useful for testing purposes.
func Main() (exitCode int) {
	//the "info" action does not require any scanning
	if os.Args[1] == "info" {
		os.Stdout.Write([]byte("MIN_API_VERSION=3\nMAX_API_VERSION=3\n"))
		return 0
	}

	//scan for entities
	entities := impl.ScanRepo()
	if entities == nil {
		//some fatal error occurred - it was already reported, so just exit
		return 1
	}

	//scan action requires no arguments
	if os.Args[1] == "scan" {
		for _, entity := range entities {
			entity.PrintReport()
		}
		return 0
	}

	//all other actions require an entity selection
	entityID := os.Args[2]
	var selectedEntity *impl.TargetFile
	for _, entity := range entities {
		if entity.EntityID() == entityID {
			selectedEntity = entity
			break
		}
	}
	if selectedEntity == nil {
		fmt.Fprintf(os.Stderr, "!! unknown entity ID \"%s\"\n", entityID)
		return 1
	}

	switch os.Args[1] {
	case "apply":
		applyEntity(selectedEntity, false)
	case "force-apply":
		applyEntity(selectedEntity, true)
	case "diff":
		output := fmt.Sprintf("%s\000%s\000",
			selectedEntity.PathIn(common.ProvisionedDirectory()),
			selectedEntity.PathIn(common.TargetDirectory()),
		)
		_, err := os.NewFile(3, "file descriptor 3").Write([]byte(output))
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		}
	}

	return 0
}

func applyEntity(entity *impl.TargetFile, withForce bool) {
	skipReport, needForceToOverwrite, needForceToRestore := entity.Apply(withForce)

	if skipReport {
		_, err := os.NewFile(3, "file descriptor 3").Write([]byte("not changed\n"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		}
	}

	if needForceToOverwrite {
		_, err := os.NewFile(3, "file descriptor 3").Write([]byte("requires --force to overwrite\n"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		}
	}

	if needForceToRestore {
		_, err := os.NewFile(3, "file descriptor 3").Write([]byte("requires --force to restore\n"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		}
	}
}
