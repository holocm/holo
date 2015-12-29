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
	"errors"
	"fmt"
	"regexp"
	"strings"
)

//Key represents a single public key, i.e. a non-comment line in a KeyFile.
type Key struct {
	Options string
	KeyType string
	Key     string
	Comment string
}

//list of valid SSH key types, as extracted from man:sshd(8), section
//"authorized_keys file format"
var isKeyType = map[string]bool{
	"ecdsa-sha2-nistp256": true,
	"ecdsa-sha2-nistp384": true,
	"ecdsa-sha2-nistp521": true,
	"ssh-ed25519":         true,
	"ssh-dss":             true,
	"ssh-rsa":             true,
}
var whiteSpaceRx = regexp.MustCompile(`\s+`)
var whiteSpaceAtEndRx = regexp.MustCompile(`\s+$`)

//ParseKey creates a Key struct by parsing a line from a KeyFile.
func ParseKey(line string) (*Key, error) {
	var key Key

	//trim trailing newline
	line = whiteSpaceAtEndRx.ReplaceAllString(line, "")
	//split off first field
	fields1 := whiteSpaceRx.Split(line, 2)
	if len(fields1) == 1 {
		return nil, fmt.Errorf("key invalid (expected 2+ fields, found less): '%s'", line)
	}

	//first field is either Options or KeyType
	var requiredFieldCount int
	if isKeyType[fields1[0]] {
		key.KeyType = fields1[0]
		requiredFieldCount = 1 //key (comment is optional)
	} else {
		key.Options = fields1[0]
		requiredFieldCount = 2 //key type, key (comment is optional)
	}

	//split remaining fields
	fields2 := whiteSpaceRx.Split(fields1[1], requiredFieldCount+1) //plus 1 for optional comment
	if len(fields2) < requiredFieldCount {
		return nil, fmt.Errorf("key invalid (expected %d+ fields, found %d): '%s'",
			requiredFieldCount+1, //plus 1 for fields1[0]
			len(fields2)+1,       //plus 1 for fields1[0]
			line,
		)
	}

	//place remaining fields in struct (with comment being optional)
	if key.KeyType == "" {
		key.KeyType = fields2[0]
		key.Key = fields2[1]
		if len(fields2) > 2 {
			key.Comment = strings.TrimSpace(fields2[2]) //TrimSpace removes trailing "\n"
		}
	} else {
		key.Key = fields2[0]
		if len(fields2) > 1 {
			key.Comment = strings.TrimSpace(fields2[1])
		}
	}

	//confirm key type
	if !isKeyType[key.KeyType] {
		return nil, fmt.Errorf("key invalid (unknown key type): '%s'", line)
	}

	return &key, nil
}

//String creates the string representation of this key.
func (k *Key) String() string {
	allFields := []string{k.Options, k.KeyType, k.Key, k.Comment}
	fields := make([]string, 0, len(allFields))
	for _, str := range allFields {
		if str != "" {
			fields = append(fields, str)
		}
	}
	return strings.Join(fields, " ")
}
