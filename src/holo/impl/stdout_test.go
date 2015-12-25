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
	"testing"
)

func checkParagraphWriter(t *testing.T, expected string, callback func(w *ParagraphWriter)) {
	var b bytes.Buffer
	w := ParagraphWriter{Writer: &b}
	callback(&w)

	result := string(b.Bytes())
	if result != expected {
		t.Errorf("expected output %#v but got %#v", expected, result)
	}
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
