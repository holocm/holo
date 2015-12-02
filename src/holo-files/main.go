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

	"./impl"
)

// #include <locale.h>
import "C"

func init() {
	//Holo requires a neutral locale, esp. for deterministic sorting of file paths
	lcAll := C.int(0)
	C.setlocale(lcAll, C.CString("C"))
}

func main() {
	if version := os.Getenv("HOLO_API_VERSION"); version != "1" {
		fmt.Fprintf(os.Stderr, "!! holo-users-groups plugin called with unknown HOLO_API_VERSION %s\n", version)
	}

	//scan for entities
	entities := impl.ScanRepo()
	if entities == nil {
		//some fatal error occurred - it was already reported, so just exit
		os.Exit(1)
	}

	//scan action requires no arguments
	if os.Args[1] == "scan" {
		for _, entity := range entities {
			entity.PrintReport()
		}
		return
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

func applyEntity(entity *impl.TargetFile, withForce bool) {
	skipReport := entity.Apply(withForce)
	if skipReport {
		_, err := os.NewFile(3, "file descriptor 3").Write([]byte("not changed\n"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		}
	}
}
