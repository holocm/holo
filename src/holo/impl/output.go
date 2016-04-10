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
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

//Errorf formats and prints an error message on stderr.
func Errorf(writer io.Writer, text string, args ...interface{}) {
	if len(args) > 0 {
		text = fmt.Sprintf(text, args...)
	}
	fmt.Fprintf(writer, "\x1b[1;31m!! %s\x1b[0m\n", strings.TrimSuffix(text, "\n"))
}

//Warnf formats and prints an warning message on stderr.
func Warnf(writer io.Writer, text string, args ...interface{}) {
	if len(args) > 0 {
		text = fmt.Sprintf(text, args...)
	}
	fmt.Fprintf(writer, "\x1b[1;33m>> %s\x1b[0m\n", strings.TrimSuffix(text, "\n"))
}

//ParagraphTracker is used in conjunction with ParagraphWriter. See explanation
//over there.
type ParagraphTracker struct {
	PrimaryWriter        io.Writer
	hadOutput            bool
	trailingNewlineCount int
}

//ParagraphWriter is an io.Writer that forwards to another io.Writer, but
//ensures that input is written in paragraphs, with newlines in between.
//
//Since, in this usecase, both stdout and stderr need to be PrologueWriter
//instances, the logic that prints the additional newlines must be shared by
//both. Thus the newlines are tracked with a ParagraphTracker instance.
type ParagraphWriter struct {
	Writer  io.Writer
	Tracker *ParagraphTracker
}

var stdTracker = &ParagraphTracker{PrimaryWriter: os.Stdout}

//Stdout wraps os.Stdout into a ParagraphWriter.
var Stdout = &ParagraphWriter{Writer: os.Stdout, Tracker: stdTracker}

//Stderr wraps os.Stderr into a ParagraphWriter.
var Stderr = &ParagraphWriter{Writer: os.Stderr, Tracker: stdTracker}

func (t *ParagraphTracker) observeOutput(p []byte) {
	//print the initial newline before any other output
	if !t.hadOutput {
		t.PrimaryWriter.Write([]byte{'\n'})
		t.hadOutput = true
	}

	//count trailing newlines on the output that was seen
	cnt := 0
	for cnt < len(p) && p[len(p)-1-cnt] == '\n' {
		cnt++
	}
	if cnt == len(p) {
		t.trailingNewlineCount += cnt
	} else {
		t.trailingNewlineCount = cnt
	}
}

//Write implements the io.Writer interface.
func (w *ParagraphWriter) Write(p []byte) (n int, e error) {
	w.Tracker.observeOutput(p)
	return w.Writer.Write(p)
}

//EndParagraph inserts newlines to start the next paragraph of output.
func (w *ParagraphWriter) EndParagraph() {
	if !w.Tracker.hadOutput {
		return
	}
	for w.Tracker.trailingNewlineCount < 2 {
		w.Write([]byte{'\n'})
	}
}

//PrologueTracker is used in conjunction with PrologueWriter. See explanation
//over there.
type PrologueTracker struct {
	Printer func()
}

//Exec prints the prologue if it has not been printed before.
func (t *PrologueTracker) Exec() {
	//print prologue exactly once
	if t.Printer != nil {
		t.Printer()
		t.Printer = nil
	}
}

//PrologueWriter is an io.Writer that ensures that a prologue is printed before
//any writes to the underlying io.Writer occur. This is used by entity.Apply()
//to print the scan report before any other output, but only if there is output.
//
//Since, in this usecase, both stdout and stderr need to be PrologueWriter
//instances, the function that prints the prologue must be shared by both, and
//it needs to be made sure that the prologue is only printed once. Thus the
//prologue is tracked with a PrologueTracker instance.
type PrologueWriter struct {
	Writer  io.Writer
	Tracker *PrologueTracker
}

//Write implements the io.Writer interface.
func (w *PrologueWriter) Write(p []byte) (n int, e error) {
	//skip empty writes
	if len(p) == 0 {
		return 0, nil
	}

	//ensure that prologue is printed
	w.Tracker.Exec()
	return w.Writer.Write(p)
}

//LineColorizingRule is a rule for the LineColorizingWriter (see there).
type LineColorizingRule struct {
	Prefix []byte
	Color  []byte
}

//ColorizeLine adds color to the given line according to the first of the given
//`rules` that matches.
func ColorizeLine(line []byte, rules []LineColorizingRule) []byte {
	for _, rule := range rules {
		if bytes.HasPrefix(line, rule.Prefix) {
			return bytes.Join([][]byte{rule.Color, line, []byte("\x1b[0m")}, nil)
		}
	}
	return line
}

//ColorizeLines is like ColorizeLine, but acts on multiple lines.
func ColorizeLines(lines []byte, rules []LineColorizingRule) []byte {
	sep := []byte{'\n'}
	in := bytes.Split(lines, sep)
	out := make([][]byte, 0, len(in))
	for _, line := range in {
		out = append(out, ColorizeLine(line, rules))
	}
	return bytes.Join(out, sep)
}

//LineColorizingWriter is an io.Writer that adds ANSI colors to lines of text
//written into it. It then passes the colorized lines to another writer.
//Coloring is based on prefixes. For example, to turn all lines with a "!!"
//prefix red, use
//
//    colorizer = &LineColorizingWriter {
//        Writer: otherWriter,
//        Rules: []LineColorizingRule {
//            LineColorizingRule { []byte("!!"), []byte("\x1B[1;31m") },
//        },
//    }
//
type LineColorizingWriter struct {
	Writer io.Writer
	Rules  []LineColorizingRule
	buffer []byte
}

//Write implements the io.Writer interface.
func (w *LineColorizingWriter) Write(p []byte) (n int, err error) {
	//append `p` to buffer and report everything as written
	w.buffer = append(w.buffer, p...)
	n = len(p)

	for {
		//check if we have a full line in the buffer
		idx := bytes.IndexByte(w.buffer, '\n')
		if idx == -1 {
			return n, nil
		}

		//extract line from buffer
		line := append(ColorizeLine(w.buffer[0:idx], w.Rules), '\n')
		w.buffer = w.buffer[idx+1:]

		//check if a colorizing rule matches
		_, err := w.Writer.Write(line)
		if err != nil {
			return n, err
		}
	}
}
