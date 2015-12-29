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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var stateFilePath = filepath.Join(os.Getenv("HOLO_STATE_DIR"), "provisioned-entities")

//ProvisionedEntities returns all entity names for which keys have been provisioned.
func ProvisionedEntities() ([]string, error) {
	contents, err := ioutil.ReadFile(stateFilePath)
	if err != nil {
		return nil, err
	}
	str := strings.TrimSpace(string(contents))
	return strings.Split(str, "\n"), nil
}

//SetEntityProvisioned adds or removes an entity name from the list of
//ProvisionedEntities().
func SetEntityProvisioned(entityName string, provisioned bool) error {
	entities, err := ProvisionedEntities()
	if err != nil {
		return err
	}

	if provisioned {
		//add entity to list
		for _, entity := range entities {
			if entity == entityName {
				return nil //already in list - nothing to do
			}
		}
		entities = append(entities, entityName)
	} else {
		//remove entity from list
		newEntities := make([]string, 0, len(entities))
		for _, entity := range entities {
			if entity != entityName {
				newEntities = append(newEntities, entity)
			}
		}
		//does this change the list?
		if len(newEntities) == len(entities) {
			return nil
		}
		entities = newEntities
	}

	//write changed list
	str := strings.Join(entities, "\n") + "\n"
	return ioutil.WriteFile(stateFilePath, []byte(str), 0644)
}
