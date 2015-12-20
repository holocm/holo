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
	"io/ioutil"
	"os"
	"path/filepath"

	"../internal/toml"
)

//Registry lists the provisioned users and groups.
type Registry struct {
	ProvisionedUsers  []string
	ProvisionedGroups []string
}

var reg Registry
var regPath = filepath.Join(os.Getenv("HOLO_STATE_DIR"), "state.toml")

func init() {
	//load registry file (ignore file-not-found since that means that `holo
	//apply` runs for the first time)
	blob, err := ioutil.ReadFile(regPath)
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		os.Exit(1)
	}
	_, err = toml.Decode(string(blob), &reg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		os.Exit(1)
	}
}

func saveRegistry() error {
	regPathNew := regPath + ".new"
	file, err := os.Create(regPathNew)
	if err != nil {
		return err
	}
	err = toml.NewEncoder(file).Encode(&reg)
	if err != nil {
		return err
	}

	//move the new registry file over the old one atomically, to avoid
	//corruption of the registry file in case of unforeseen errors
	return os.Rename(regPathNew, regPath)
}

//RegistryPath returns the path to the registry, for use in information messages.
func RegistryPath() string {
	return regPath
}

//AddProvisionedGroup records that the given group has been provisioned.
func AddProvisionedGroup(name string) error {
	var changed bool
	reg.ProvisionedGroups, changed = appendIfMissing(reg.ProvisionedGroups, name)
	if changed {
		return saveRegistry()
	}
	return nil
}

//AddProvisionedUser records that the given user has been provisioned.
func AddProvisionedUser(name string) error {
	var changed bool
	reg.ProvisionedUsers, changed = appendIfMissing(reg.ProvisionedUsers, name)
	if changed {
		return saveRegistry()
	}
	return nil
}

func appendIfMissing(list []string, value string) (newList []string, changed bool) {
	for _, element := range list {
		if element == value {
			return list, false
		}
	}
	return append(list, value), true
}
