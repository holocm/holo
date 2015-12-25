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
	"io/ioutil"
	"os"
)

//InfoLine represents a line in the information section of an Entity.
type InfoLine struct {
	attribute string
	value     string
}

//Entity represents an entity known to some Holo plugin.
type Entity struct {
	plugin       *Plugin
	id           string
	actionVerb   string
	actionReason string
	sourceFiles  []string
	infoLines    []InfoLine
}

//EntityID returns a string that uniquely identifies the entity.
func (e *Entity) EntityID() string { return e.id }

//MatchesSelector checks whether the given string is either the entity ID or a
//source file of this entity.
func (e *Entity) MatchesSelector(value string) bool {
	if e.id == value {
		return true
	}
	for _, file := range e.sourceFiles {
		if file == value {
			return true
		}
	}
	return false
}

//PrintReport prints the scan report describing this Entity.
func (e *Entity) PrintReport(withAction bool) {
	//print initial line with action and entity ID
	//(note that Stdout != os.Stdout)
	var lineFormat string
	if e.actionVerb == "" || !withAction {
		lineFormat = "%12s %s\n"
		fmt.Fprintf(Stdout, "\x1b[1m%s\x1b[0m", e.id)
	} else {
		lineFormat = fmt.Sprintf("%%%ds %%s\n", len(e.actionVerb))
		fmt.Fprintf(Stdout, "%s \x1b[1m%s\x1b[0m", e.actionVerb, e.id)
	}
	if e.actionReason == "" {
		Stdout.Write([]byte{'\n'})
	} else {
		fmt.Fprintf(Stdout, " (%s)\n", e.actionReason)
	}

	//print info lines
	for _, line := range e.infoLines {
		fmt.Fprintf(Stdout, lineFormat, line.attribute, line.value)
	}
	Stdout.EndParagraph()
}

//Apply performs the complete application algorithm for the given Entity.
func (e *Entity) Apply(withForce bool) {
	err := e.doApply(withForce)
	if err != nil {
		Errorf(err.Error())
	}
}

func (e *Entity) doApply(withForce bool) error {
	command := "apply"
	if withForce {
		command = "force-apply"
	}

	//the command channel (file descriptor 3 on the side of the plugin) can
	//only be set up with an *os.File instance, so use a pipe that the plugin
	//writes into and that we read from
	cmdReader, cmdWriterForPlugin, err := os.Pipe()
	if err != nil {
		e.PrintReport(true)
		return err
	}

	//TODO: This implementation is stupid and buffers all the output before
	//deciding what to print and how. Technically we could just patch stdout
	//and stderr through directly, but there is a caveat: We always want the
	//scan report in front of all output.
	var output bytes.Buffer
	cmd := e.plugin.Command([]string{command, e.id}, &output, &output, cmdWriterForPlugin)
	err = cmd.Start() //cannot use Run() since we need to read from the pipe before the plugin exits
	if err != nil {
		e.PrintReport(true)
		return err
	}

	cmdWriterForPlugin.Close() //or next line will block (see Plugin.Command docs)
	cmdBytes, err := ioutil.ReadAll(cmdReader)
	if err != nil {
		e.PrintReport(true)
		return err
	}
	err = cmdReader.Close()
	if err != nil {
		e.PrintReport(true)
		return err
	}
	err = cmd.Wait()

	//only print report if there was output, or if the plugin provisioned the
	//entity (as signaled by the absence of the "not changed\n" command")
	showReport := true
	if output.Len() == 0 && err == nil {
		cmdLines := bytes.Split(cmdBytes, []byte("\n"))
		for _, line := range cmdLines {
			if string(line) == "not changed" {
				showReport = false
			}
		}
	}
	if showReport {
		e.PrintReport(true)
	}

	//forward stdout
	if output.Len() > 0 {
		Stdout.Write(output.Bytes())
		Stdout.EndParagraph()
	}

	return err
}

//RenderDiff creates a unified diff between the current and last
//provisioned version of this entity.
func (e *Entity) RenderDiff() ([]byte, error) {
	var buffer bytes.Buffer
	err := e.plugin.Command([]string{"diff", e.id}, &buffer, os.Stderr, nil).Run()
	return buffer.Bytes(), err
}
