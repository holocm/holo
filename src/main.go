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

func main() {
	if version := os.Getenv("HOLO_API_VERSION"); version != "2" {
		fmt.Fprintf(os.Stderr, "!! holo-users-groups plugin called with unknown HOLO_API_VERSION %s\n", version)
	}

	//the scan operation does not require any arguments
	if os.Args[1] == "scan" {
		errs := impl.Scan()
		for _, err := range errs {
			fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		}
		return
	}

	//all other operations work on an entity
	entity, err := impl.NewEntityFromName(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		os.Exit(1)
	}

	switch os.Args[1] {
	case "apply", "force-apply":
		err := entity.Apply()
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		}
	case "diff":
		output, err := entity.RenderDiff()
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		}
		os.Stdout.Write(output)
	}
}
