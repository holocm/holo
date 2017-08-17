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
	"regexp"
	"strings"
	"testing"
)

//This is a real public key, but don't worry, it has never been used in
//production.
var testKey = "AAAAB3NzaC1yc2EAAAADAQABAAABAQC9lA02DybCuFKOMhcvCTgUphvpGht1waGT93RvqXYBTGKUcJYz09abjaArAv/dQGnX8gjYogwzXvre5tRZiLaGvpMBRQvozSU9NVQSZs4Qv6wXGEqS7eFc7A+sCQFBhFy7H86woJhWa47L7c7TzX0OD9mjksJrH8AZON4Vv3gUDJjQqfAx8HAF8l96VHuaM+DVnYYcZjRUTyt1kLH40Wi/v/R8LF74Nq9Ah72I8KGEHOB+4xoz5VX3flur1md2MYOdBFOOwFERJMqjp3ZQ2KErdq/UcPE92O89yIMGbaACL9pObh3K3oR5SHitw2nz4oveunZh0yOfLsubIazfHoIn"

func checkParseKey(t *testing.T, line string, expectedKey *Key, expectedError string) {
	//line may contain "%s" which stands for the testKey
	if strings.Contains(line, "%s") {
		line = fmt.Sprintf(line, testKey)
	}

	actualKey, actualErr := ParseKey(line + "\n")

	//print for divergences in a readable format
	var divergences []string
	if actualKey == nil {
		if expectedKey != nil {
			divergences = append(divergences, "expected a key, but got nil")
		}
	} else {
		if expectedKey == nil {
			divergences = append(divergences, fmt.Sprintf("expected no key, but got %#v", actualKey))
		} else {
			//both keys exist, so check the individual fields
			if actualKey.Options != expectedKey.Options {
				divergences = append(divergences, fmt.Sprintf(
					"key.Options = %#v, expected %#v", actualKey.Options, expectedKey.Options,
				))
			}
			if actualKey.KeyType != expectedKey.KeyType {
				divergences = append(divergences, fmt.Sprintf(
					"key.KeyType = %#v, expected %#v", actualKey.KeyType, expectedKey.KeyType,
				))
			}
			if actualKey.Key != expectedKey.Key {
				divergences = append(divergences, fmt.Sprintf(
					"key.Key = %#v, expected %#v", actualKey.Key, expectedKey.Key,
				))
			}
			if actualKey.Comment != expectedKey.Comment {
				divergences = append(divergences, fmt.Sprintf(
					"key.Comment = %#v, expected %#v", actualKey.Comment, expectedKey.Comment,
				))
			}
		}
	}
	if actualErr == nil {
		if expectedError != "" {
			divergences = append(divergences, fmt.Sprintf("expected error '%s', but got nil", expectedError))
		}
	} else {
		if expectedError == "" {
			divergences = append(divergences, fmt.Sprintf("got error '%s', but expected nil", actualErr.Error()))
		} else {
			if !regexp.MustCompile(expectedError).MatchString(actualErr.Error()) {
				divergences = append(divergences, fmt.Sprintf(
					"err = %#v, expected to match %#v", actualErr.Error(), expectedError,
				))
			}
		}
	}

	if len(divergences) > 0 {
		t.Errorf("ParseKey(%#v) failed", strings.Replace(line, testKey, "{key}", -1))
		for _, str := range divergences {
			t.Log("- " + strings.Replace(str, testKey, "{key}", -1))
		}
	}

	//also check Key.String() method; it should recover the same line
	if actualKey != nil {
		str := actualKey.String()
		if str != line {
			msg := fmt.Sprintf("ParseKey(%#v).String() failed, returned %#v", line, str)
			t.Error(strings.Replace(msg, testKey, "{key}", -1))
		}
	}
}

func TestParseKey(t *testing.T) {
	//check the correct handling of presence/absence of optional fields
	checkParseKey(t, "ssh-rsa %s",
		&Key{KeyType: "ssh-rsa", Key: testKey},
		"",
	)
	checkParseKey(t, "no-X11-forwarding ssh-rsa %s",
		&Key{Options: "no-X11-forwarding", KeyType: "ssh-rsa", Key: testKey},
		"",
	)
	checkParseKey(t, "ssh-rsa %s john@foobar",
		&Key{KeyType: "ssh-rsa", Key: testKey, Comment: "john@foobar"},
		"",
	)
	checkParseKey(t, "no-X11-forwarding ssh-rsa %s john@foobar",
		&Key{Options: "no-X11-forwarding", KeyType: "ssh-rsa", Key: testKey, Comment: "john@foobar"},
		"",
	)

	//check error cases
	checkParseKey(t, "", nil,
		`key invalid \(expected 2\+ fields, found less\): '.*'`,
	)
	checkParseKey(t, "ssh-rsa", nil,
		`key invalid \(expected 2\+ fields, found less\): '.*'`,
	)
	checkParseKey(t, "ssh-dsa %s", nil, //there is no key type "ssh-dsa", it's "ssh-dss"
		`key invalid \(expected 3\+ fields, found 2\): '.*'`,
	)
	checkParseKey(t, "no-X11-forwarding ssh-dsa %s", nil, //same error as above in different context
		`key invalid \(unknown key type\): '.*'`,
	)
}
