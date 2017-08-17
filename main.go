/*******************************************************************************
*
* Copyright 2017 Stefan Majewsky <majewsky@gmx.net>
*
* This program is free software: you can redistribute it and/or modify it under
* the terms of the GNU General Public License as published by the Free Software
* Foundation, either version 3 of the License, or (at your option) any later
* version.
*
* This program is distributed in the hope that it will be useful, but WITHOUT ANY
* WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
* A PARTICULAR PURPOSE. See the GNU General Public License for more details.
*
* You should have received a copy of the GNU General Public License along with
* this program. If not, see <http://www.gnu.org/licenses/>.
*
*******************************************************************************/

package main

import (
	"os"
	"path/filepath"

	cmd_holo "github.com/holocm/holo/cmd/holo"
	cmd_holo_files "github.com/holocm/holo/cmd/holo-files"
)

func main() {
	switch filepath.Base(os.Args[0]) {
	case "holo-files":
		os.Exit(cmd_holo_files.Main())
	default:
		os.Exit(cmd_holo.Main())
	}
}
