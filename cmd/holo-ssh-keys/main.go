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

package entrypoint

import (
	"fmt"
	"os"

	"github.com/holocm/holo/cmd/holo-ssh-keys/impl"
)

// Main is the main entry point, but returns the exit code rather than
// calling os.Exit().  This distinction is useful for testing purposes.
func Main() (exitCode int) {
	if version := os.Getenv("HOLO_API_VERSION"); version != "3" {
		fmt.Fprintf(os.Stderr, "!! holo-users-groups plugin called with unknown HOLO_API_VERSION %s\n", version)
		return 1
	}

	//operations that do not require any arguments
	switch os.Args[1] {
	case "info":
		os.Stdout.Write([]byte("MIN_API_VERSION=3\nMAX_API_VERSION=3\n"))
		return
	case "scan":
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
		return 1
	}

	switch os.Args[1] {
	case "apply", "force-apply":
		err := entity.Apply()
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		}
	case "diff":
		expectedStateFile, actualStateFile, err := entity.PrepareDiff()
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		}
		out := fmt.Sprintf("%s\000%s\000", expectedStateFile, actualStateFile)
		_, err = os.NewFile(3, "file descriptor 3").Write([]byte(out))
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		}
	}

	return 0
}
