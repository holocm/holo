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
	"regexp"
	"sort"
	"strings"
)

//Scan discovers entities available for the given entity. Errors are reported
//immediately and will result in nil being returned. "No entities found" will
//be reported as a non-nil empty slice.
//there are no entities.
func (p *Plugin) Scan() []*Entity {
	//invoke scan operation
	stdout, hadError := p.runScanOperation()
	if hadError {
		return nil
	}

	//parse scan output
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	lineRx := regexp.MustCompile(`^\s*([^:]+): (.+)\s*$`)
	actionRx := regexp.MustCompile(`^([^()]+) \((.+)\)$`)
	hadError = false
	var currentEntity *Entity
	var result []*Entity
	for idx, line := range lines {
		//skip empty lines
		if line == "" {
			continue
		}

		//keep format strings from getting too long
		errorIntro := fmt.Sprintf("error in scan report of %s, line %d", p.ID(), idx+1)

		//general line format is "key: value"
		match := lineRx.FindStringSubmatch(line)
		if match == nil {
			Errorf(Stderr, "%s: parse error (line was \"%s\")", errorIntro, line)
			hadError = true
			continue
		}
		key, value := match[1], match[2]

		switch {
		case key == "ENTITY":
			//starting new entity
			if currentEntity != nil {
				result = append(result, currentEntity)
			}
			currentEntity = &Entity{plugin: p, id: value, actionVerb: "Working on"}
		case currentEntity == nil:
			//if not, we need to be inside an entity
			//(i.e. line with idx = 0 must start an entity)
			Errorf(Stderr, "%s: expected entity ID, found attribute \"%s\"", errorIntro, line)
			hadError = true
		case key == "SOURCE":
			currentEntity.sourceFiles = append(currentEntity.sourceFiles, value)
		case key == "ACTION":
			//parse action verb/reason
			match = actionRx.FindStringSubmatch(value)
			if match == nil {
				currentEntity.actionVerb = value
				currentEntity.actionReason = ""
			} else {
				currentEntity.actionVerb = match[1]
				currentEntity.actionReason = match[2]
			}
		default:
			//store unrecognized keys as info lines
			currentEntity.infoLines = append(currentEntity.infoLines,
				InfoLine{key, value},
			)
		}
	}

	//store last entity
	if currentEntity != nil {
		result = append(result, currentEntity)
	}

	//report errors
	if hadError {
		return nil
	}

	//on success, ensure non-nil return value
	if result == nil {
		result = []*Entity{}
	}

	sort.Sort(entitiesByID(result))
	return result
}

func (p *Plugin) runScanOperation() (stdout string, hadError bool) {
	var stdoutBuffer bytes.Buffer
	err := p.Command([]string{"scan"}, &stdoutBuffer, Stderr, nil).Run()

	if err != nil {
		Errorf(Stderr, "scan with plugin %s failed: %s", p.ID(), err.Error())
	}

	return string(stdoutBuffer.Bytes()), err != nil
}

type entitiesByID []*Entity

func (e entitiesByID) Len() int           { return len(e) }
func (e entitiesByID) Less(i, j int) bool { return e[i].EntityID() < e[j].EntityID() }
func (e entitiesByID) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
