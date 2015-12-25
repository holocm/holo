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

type reportLine struct {
	key   string
	value string
}

//Report formats information for an action taken on a single target, including
//warning and error messages.
type Report struct {
	Action    string
	Target    string
	State     string
	infoLines []reportLine
	msgText   string
}

//AddLine adds an information line to the given Report.
func (r *Report) AddLine(key, value string) {
	r.infoLines = append(r.infoLines, reportLine{key, value})
}

func (r *Report) addMessage(color, prefix, text string, args ...interface{}) {
	if len(args) > 0 {
		text = fmt.Sprintf(text, args...)
	}
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	r.msgText += fmt.Sprintf("\x1b[%sm\x1b[1m%s\x1b[0m %s", color, prefix, text)
}

//AddError adds an error message to the given Report. If args... are given,
//fmt.Sprintf() is applied.
func (r *Report) AddError(text string, args ...interface{}) { r.addMessage("31", "!!", text, args...) }

var reportsWerePrinted bool

//Print prints the full report on stdout.
func (r *Report) Print() {
	//print to stdout or stderr?
	out := os.Stdout
	if r.msgText != "" {
		out = os.Stderr
	}

	//before the first report, print a newline to get the paragraph formatting right
	if !reportsWerePrinted {
		out.Write([]byte{'\n'})
		reportsWerePrinted = true
	}

	//print initial line with Action, Target and State
	var lineFormat string
	if r.Action == "" {
		lineFormat = "%12s %s\n"
		fmt.Fprintf(out, "\x1b[1m%s\x1b[0m", r.Target)
	} else {
		lineFormat = fmt.Sprintf("%%%ds %%s\n", len(r.Action))
		fmt.Fprintf(out, "%s \x1b[1m%s\x1b[0m", r.Action, r.Target)
	}
	if r.State == "" {
		out.Write([]byte{'\n'})
	} else {
		fmt.Fprintf(out, " (%s)\n", r.State)
	}

	//print infoLines
	for _, line := range r.infoLines {
		if line.key != "" {
			fmt.Fprintf(out, lineFormat, line.key, line.value)
		}
	}
	if len(r.infoLines) > 0 {
		out.Write([]byte{'\n'})
	}

	//print message text, if any
	if r.msgText != "" {
		out.Write([]byte(r.msgText))
		out.Write([]byte{'\n'})
	}
}
