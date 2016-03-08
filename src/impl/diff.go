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
	"strconv"
	"strings"

	"../../localdeps/github.com/BurntSushi/toml"
)

//RenderDiff implements the Entity interface.
func (group Group) RenderDiff() ([]byte, error) {
	//does this group exist already?
	groupExists, actualGid, err := group.checkExists()
	if err != nil {
		return nil, err
	}

	//if the group is orphaned, and there exists no actual group, then there is no diff
	if !groupExists && group.Orphaned {
		return nil, nil
	}

	//to simplify the diff process, replace a non-existing group by an empty group
	headers := generateDiffHeader("group", group.EntityID(), groupExists)

	//generate body
	var lines []string
	if groupExists {
		if group.Orphaned {
			lines = []string{"+[[group]]"}
		} else {
			lines = []string{" [[group]]"}
		}
	} else {
		lines = []string{"-[[group]]"}
	}

	lines, err = addDiffForField(lines, groupExists, group.Orphaned, "name", group.Name, group.Name, "")
	if err != nil {
		return nil, err
	}
	lines, err = addDiffForField(lines, groupExists, group.Orphaned, "gid", group.GID, actualGid, 0)
	if err != nil {
		return nil, err
	}

	//is there any diff?
	if !hasDiff(lines) {
		return nil, nil
	}

	//count lines for "@@ -from +to" hunk header
	hunkHeader := generateHunkHeader(lines)
	allLines := append(append(headers, hunkHeader), lines...)
	return []byte(strings.Join(allLines, "\n") + "\n"), nil
}

//RenderDiff implements the Entity interface.
func (user User) RenderDiff() ([]byte, error) {
	//does this user exist already?
	userExists, actualUser, err := user.checkExists()
	if err != nil {
		return nil, err
	}

	//if the user is orphaned, and there exists no actual user, then there is no diff
	if !userExists && user.Orphaned {
		return nil, nil
	}

	//to simplify the diff process, replace a non-existing user by an empty user
	if !userExists {
		actualUser = &User{}
	}
	headers := generateDiffHeader("user", user.EntityID(), userExists)

	//generate body
	var lines []string
	if userExists {
		if user.Orphaned {
			lines = []string{"+[[user]]"}
		} else {
			lines = []string{" [[user]]"}
		}
	} else {
		lines = []string{"-[[user]]"}
	}

	lines, err = addDiffForField(lines, userExists, user.Orphaned, "name", user.Name, user.Name, "")
	if err != nil {
		return nil, err
	}
	lines, err = addDiffForField(lines, userExists, user.Orphaned, "comment", user.Comment, actualUser.Comment, "")
	if err != nil {
		return nil, err
	}
	lines, err = addDiffForField(lines, userExists, user.Orphaned, "uid", user.UID, actualUser.UID, 0)
	if err != nil {
		return nil, err
	}
	lines, err = addDiffForField(lines, userExists, user.Orphaned, "home", user.HomeDirectory, actualUser.HomeDirectory, "")
	if err != nil {
		return nil, err
	}
	lines, err = addDiffForField(lines, userExists, user.Orphaned, "group", user.Group, actualUser.Group, "")
	if err != nil {
		return nil, err
	}
	lines, err = addDiffForField(lines, userExists, user.Orphaned, "groups", user.Groups, actualUser.Groups, []string{})
	if err != nil {
		return nil, err
	}
	lines, err = addDiffForField(lines, userExists, user.Orphaned, "shell", user.Shell, actualUser.Shell, "")
	if err != nil {
		return nil, err
	}

	//is there any diff?
	if !hasDiff(lines) {
		return nil, nil
	}

	//count lines for "@@ -from +to" hunk header
	hunkHeader := generateHunkHeader(lines)
	allLines := append(append(headers, hunkHeader), lines...)
	return []byte(strings.Join(allLines, "\n") + "\n"), nil
}

func generateDiffHeader(entityType, entityID string, entityExists bool) []string {
	//generate diff header (much of this is made up since there is no external
	//reference for a diff format for users/groups)
	headers := []string{
		fmt.Sprintf("diff --holo %s", entityID),
	}
	if !entityExists {
		headers = append(headers, "deleted "+entityType)
	}
	headers = append(headers, fmt.Sprintf("--- %s", entityID))
	if entityExists {
		return append(headers, fmt.Sprintf("+++ %s", entityID))
	}
	return append(headers, "+++ /dev/null")
}

func generateHunkHeader(lines []string) string {
	//count lines for "@@ -from +to" hunk header
	fromLines, toLines := 0, 0
	for _, line := range lines {
		switch line[0] {
		case ' ':
			fromLines++
			toLines++
		case '-':
			fromLines++
		case '+':
			toLines++
		}
	}
	return fmt.Sprintf("@@ -%s +%s", lineCountToString(fromLines), lineCountToString(toLines))
}

func lineCountToString(count int) string {
	//format line count for use in hunk header of unified diff
	switch count {
	case 0:
		return "0,0"
	case 1:
		return "1"
	default:
		return "1," + strconv.Itoa(count)
	}
}

//Produce a content diff for the given field, by encoding the expectedValue and
//actualValue as TOML. No output is produced if the expectedValue matches the
//ignoredValue, which means that the value is not set in the entity definition.
func addDiffForField(lines []string, entityExists, isEntityOrphaned bool, field string, expectedValue, actualValue, ignoredValue interface{}) ([]string, error) {
	//early exit if entity exists, but is orphaned, i.e. we print a diff with "+" lines only
	if isEntityOrphaned {
		actualData, err := encodeField(field, actualValue)
		if err != nil {
			return lines, err
		}
		return append(lines, "+"+actualData), nil
	}

	//encode values into TOML
	expectedData, err := encodeField(field, expectedValue)
	if err != nil {
		return lines, err
	}
	ignoredData, err := encodeField(field, ignoredValue)
	if err != nil {
		return lines, err
	}
	if expectedData == ignoredData {
		//this field is not included in the entity definition, so don't print it in the diff
		return lines, nil
	}

	//early exit if there is no previous entity, i.e. we print a diff with "-" lines only
	if !entityExists {
		return append(lines, "-"+expectedData), nil
	}

	//encode actual value into TOML
	actualData, err := encodeField(field, actualValue)
	if err != nil {
		return lines, err
	}

	//if both are the same, print as context, else print as diff
	if expectedData == actualData {
		return append(lines, " "+expectedData), nil
	}
	return append(lines, "-"+expectedData, "+"+actualData), nil
}

func encodeField(field string, value interface{}) (string, error) {
	var buf bytes.Buffer
	err := toml.NewEncoder(&buf).Encode(map[string]interface{}{field: value})
	return strings.TrimSpace(buf.String()), err
}

//Check if the diff contains any differences (i.e. lines that are not context lines).
func hasDiff(lines []string) bool {
	for _, line := range lines {
		if line[0] != ' ' {
			return true
		}
	}
	return false
}
