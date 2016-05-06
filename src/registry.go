/*******************************************************************************
*
* Copyright 2016 Stefan Majewsky <majewsky@gmx.net>
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

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"../localdeps/github.com/BurntSushi/toml"
)

var preImageDir string

func init() {
	preImageDir = filepath.Join(os.Getenv("HOLO_STATE_DIR"), "pre-images")
	err := os.MkdirAll(preImageDir, 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		os.Exit(1)
	}
}

func preImagePathFor(def EntityDefinition) string {
	return filepath.Join(preImageDir, def.EntityID()+".toml")
}

//ProvisionedEntityIDs returns a list of all entities for which pre-images exist.
func ProvisionedEntityIDs() ([]string, error) {
	//open pre-image directory
	dir, err := os.Open(preImageDir)
	if err != nil {
		return nil, err
	}
	fis, err := dir.Readdir(-1)
	if err != nil {
		return nil, err
	}

	//find pre-images
	var ids []string
	for _, fi := range fis {
		if fi.Mode().IsRegular() && strings.HasSuffix(fi.Name(), ".toml") {
			ids = append(ids, strings.TrimSuffix(fi.Name(), ".toml"))
		}
	}
	return ids, nil
}

//LoadPreImageFor retrieves the pre-image for this entity, which was previously
//written by SavePreImage.
func LoadPreImageFor(def EntityDefinition) (EntityDefinition, error) {
	blob, err := ioutil.ReadFile(preImagePathFor(def))
	if err != nil {
		return nil, err
	}

	//prepare an empty instance to decode the file into
	var result EntityDefinition
	switch def.(type) {
	case *GroupDefinition:
		result = &GroupDefinition{}
	case *UserDefinition:
		result = &UserDefinition{}
	}
	_, err = toml.Decode(string(blob), result)
	return result, err
}

//SavePreImage writes a pre-image, i.e. the output of GetProvisionedState()
//before the first apply operation, to /var/lib/holo/users-groups/pre-images.
func SavePreImage(def EntityDefinition) error {
	file, err := os.Create(preImagePathFor(def))
	if err != nil {
		return err
	}
	return toml.NewEncoder(file).Encode(def)
}

//DeletePreImageFor deletes the pre-image for this entity.
func DeletePreImageFor(def EntityDefinition) error {
	return os.Remove(preImagePathFor(def))
}
