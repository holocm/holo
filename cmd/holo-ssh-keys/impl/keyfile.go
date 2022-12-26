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
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// KeyFile provides methods for reading and writing a file containing SSH public
// keys.
type KeyFile string

// Key represents a single public key, i.e. a non-comment line in a KeyFile.
type Key struct {
	Options string
	KeyType string
	Key     string
	Comment string
}

// list of valid SSH key types, as extracted from man:sshd(8), section
// "authorized_keys file format"
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

// ParseKey creates a Key struct by parsing a line from a KeyFile.
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

// String creates the string representation of this key.
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

// Identifier is like String(), but omits the Comment. This can be used to
// compare keys for functional identity.
func (k *Key) Identifier() string {
	copyOfKey := *k
	copyOfKey.Comment = ""
	return copyOfKey.String()
}

// Process reads the key file, pipes every key in it through keyCallback, runs
// the endCallback in the end to add new lines, then writes the result if it has
// changed.
func (f KeyFile) Process(keyCallback func(key *Key) *Key, endCallback func() (newKeys []*Key)) (changed bool, err error) {
	//the bulk is in doProcess(), this method just generates better errors
	changed, err = f.doProcess(keyCallback, endCallback)
	if err != nil {
		err = fmt.Errorf("failure occurred while processing %s: %s", string(f), err.Error())
	}
	return
}

// Walk is the readonly variant of Process.
func (f KeyFile) Walk(callback func(key *Key)) error {
	_, err := f.Process(func(key *Key) *Key {
		callback(key)
		return key
	}, nil)
	return err
}

func (f KeyFile) doProcess(keyCallback func(key *Key) *Key, endCallback func() (newKeys []*Key)) (bool, error) {
	//read file
	contents, err := os.ReadFile(string(f))
	if err != nil {
		if os.IsNotExist(err) {
			contents = nil
		} else {
			return false, err
		}
	}

	//split into lines, but take care not to create any bogus empty lines
	contentsStr := strings.TrimSuffix(string(contents), "\n")
	lines := strings.Split(contentsStr, "\n")
	if contentsStr == "" {
		lines = nil
	}

	//go through the lines
	resultLines := make([]string, 0, len(lines))
	changed := false
	for _, line := range lines {
		//leave empty lines and comments as-is
		if line == "" || line[0] == '#' {
			resultLines = append(resultLines, line)
			continue
		}

		//all other lines must be valid keys
		key, err := ParseKey(line)
		if err != nil {
			return false, err
		}

		newKey := keyCallback(key)
		if key == newKey {
			//key unchanged
			resultLines = append(resultLines, line)
		} else {
			//replace line with newKey
			if newKey != nil {
				resultLines = append(resultLines, newKey.String())
			}
			changed = true
		}
	}

	//check if more keys shall be appended
	if endCallback != nil {
		newKeys := endCallback()
		if len(newKeys) > 0 {
			for _, key := range newKeys {
				resultLines = append(resultLines, key.String())
			}
			changed = true
		}
	}

	//if nothing changed, we're done
	if !changed {
		return false, nil
	}

	//create directories along the path
	path := string(f)
	err = os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return false, err
	}

	//write new authorized_keys file
	newContents := strings.Join(resultLines, "\n") + "\n"
	//the only files that we will ever write are user's authorized_keys
	//files, so it's a good idea to go with filemode 0600 from the start
	return true, os.WriteFile(path, []byte(newContents), 0600)
}
