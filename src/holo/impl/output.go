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

import (
	"fmt"
	"io"
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

//ParagraphWriter is an io.Writer that forwards to another io.Writer, but
//ensures that input is written in paragraphs, with newlines in between.
type ParagraphWriter struct {
	Writer               io.Writer
	hadOutput            bool
	trailingNewlineCount int
}

//Stdout is an alias of os.Stdout with the correct type, so custom functions
//can be called.
var Stdout = &ParagraphWriter{Writer: os.Stdout}

//Write implements the io.Writer interface.
func (w *ParagraphWriter) Write(p []byte) (n int, e error) {
	//print the initial newline before any other output
	if !w.hadOutput {
		w.Writer.Write([]byte{'\n'})
		w.hadOutput = true
	}

	//count trailing newlines on the output that was seen
	cnt := 0
	for cnt < len(p) && p[len(p)-1-cnt] == '\n' {
		cnt++
	}
	if cnt == len(p) {
		w.trailingNewlineCount += cnt
	} else {
		w.trailingNewlineCount = cnt
	}

	return w.Writer.Write(p)
}

//EndParagraph inserts newlines to start the next paragraph of output.
func (w *ParagraphWriter) EndParagraph() {
	if !w.hadOutput {
		return
	}
	for w.trailingNewlineCount < 2 {
		w.Write([]byte{'\n'})
	}
}
