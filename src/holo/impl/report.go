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

package impl

//TODO: A lot of this file is redundant and can probably be inlined or deleted.

import (
	"fmt"
	"os"
	"strings"
)

//Errorf formats and prints an error message on stderr.
func Errorf(text string, args ...interface{}) {
	if len(args) > 0 {
		text = fmt.Sprintf(text, args...)
	}
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	fmt.Fprintf(os.Stderr, "\x1b[31m\x1b[1m>>\x1b[0m %s", text)
}
