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
	"os"
	"testing"
)

func checkStringEqual(t *testing.T, varName, expected, actual string) {
	if expected != actual {
		t.Errorf("expected %s = %#v, but got %#v", varName, expected, actual)
	}
}

func checkParagraphWriter(t *testing.T, expected string, callback func(w *ParagraphWriter)) {
	var b bytes.Buffer
	tr := ParagraphTracker{PrimaryWriter: &b}
	w := ParagraphWriter{Writer: &b, Tracker: &tr}
	callback(&w)

	checkStringEqual(t, "output", expected, string(b.Bytes()))
}

func TestParagraphWriter(t *testing.T) {
	//check that initial newline is inserted
	checkParagraphWriter(t, "\nabc", func(w *ParagraphWriter) {
		w.Write([]byte("abc"))
	})

	//check that no additional newlines are inserted when not prompted for
	checkParagraphWriter(t, "\naaa\nbbb\nccc\n", func(w *ParagraphWriter) {
		w.Write([]byte("aaa\n"))
		w.Write([]byte("bbb\nccc\n"))
		w.Write([]byte(nil))
	})

	//check EndParagraph with a single paragraph
	checkParagraphWriter(t, "\naaa\n\n", func(w *ParagraphWriter) {
		w.Write([]byte("aaa"))
		w.EndParagraph()
	})
	checkParagraphWriter(t, "\naaa\n\n", func(w *ParagraphWriter) {
		w.Write([]byte("aaa\n"))
		w.EndParagraph()
	})
	checkParagraphWriter(t, "\naaa\n\n", func(w *ParagraphWriter) {
		w.Write([]byte("aaa\n\n"))
		w.EndParagraph()
	})

	//check EndParagraph with multiple paragraphs
	checkParagraphWriter(t, "\naaa\n\nbbb\n\n", func(w *ParagraphWriter) {
		w.Write([]byte("aaa"))
		w.EndParagraph()
		w.Write([]byte("bbb"))
		w.EndParagraph()
	})
	checkParagraphWriter(t, "\naaa\n\nbbb\n\n", func(w *ParagraphWriter) {
		w.Write([]byte("aaa\n"))
		w.EndParagraph()
		w.Write([]byte("bbb"))
		w.EndParagraph()
	})
	checkParagraphWriter(t, "\naaa\n\nbbb\n\n", func(w *ParagraphWriter) {
		w.Write([]byte("aaa\n\n"))
		w.EndParagraph()
		w.Write([]byte("bbb"))
		w.EndParagraph()
	})
	checkParagraphWriter(t, "\naaa\n\n\nbbb\n\n", func(w *ParagraphWriter) {
		w.Write([]byte("aaa\n\n\n"))
		w.EndParagraph()
		w.Write([]byte("bbb"))
		w.EndParagraph()
	})

	//check EndParagraph that is called before any output is printed
	checkParagraphWriter(t, "\naaa\n\n", func(w *ParagraphWriter) {
		w.EndParagraph()
		w.Write([]byte("aaa"))
		w.EndParagraph()
	})
}

func TestPrologueTracker(t *testing.T) {
	x := 0
	tracker := PrologueTracker{Printer: func() { x++ }}

	//check that tracker.Exec() only ever calls the callback once
	for idx := 0; idx < 3; idx++ {
		tracker.Exec()
		if x != 1 {
			Errorf(os.Stderr, "pass %d: expected 1 but got %d", idx, x)
		}
	}
}

func TestPrologueWriter(t *testing.T) {
	var buf bytes.Buffer

	//prologue is "PPP"
	tracker := PrologueTracker{Printer: func() { buf.Write([]byte("PPP")) }}
	writer := PrologueWriter{Writer: &buf, Tracker: &tracker}

	//check that empty write does not produce the prologue
	writer.Write(nil)
	writer.Write([]byte(""))
	checkStringEqual(t, "buffer content", "", string(buf.Bytes()))

	//check that non-empty write prepends the prologue
	writer.Write([]byte("xxx"))
	checkStringEqual(t, "buffer content", "PPPxxx", string(buf.Bytes()))
}
