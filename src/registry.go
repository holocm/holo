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

//ImageDir is a path to a directory containing serialized entity definitions.
type ImageDir string

//PreImageDir is usually /var/lib/holo/users-groups/base.
var PreImageDir ImageDir

//ProvisionedImageDir is usually /var/lib/holo/users-groups/provisioned.
var ProvisionedImageDir ImageDir

func init() {
	stateDir := os.Getenv("HOLO_STATE_DIR")
	PreImageDir = ImageDir(filepath.Join(stateDir, "base"))
	ProvisionedImageDir = ImageDir(filepath.Join(stateDir, "provisioned"))

	for _, dir := range []string{string(PreImageDir), string(ProvisionedImageDir)} {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
			os.Exit(1)
		}
	}
}

func (dir ImageDir) imagePathFor(def EntityDefinition) string {
	return filepath.Join(string(dir), def.EntityID()+".toml")
}

//ProvisionedEntityIDs returns a list of all entities for which pre-images exist.
func ProvisionedEntityIDs() ([]string, error) {
	//open pre-image directory
	dir, err := os.Open(string(PreImageDir))
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

//LoadImageFor retrieves a stored image for this entity, which was previously
//written by SaveImage.
func (dir ImageDir) LoadImageFor(def EntityDefinition) (EntityDefinition, error) {
	blob, err := ioutil.ReadFile(dir.imagePathFor(def))
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

//SaveImage writes an image for this entity to the specified image directory.
func (dir ImageDir) SaveImage(def EntityDefinition) error {
	file, err := os.Create(dir.imagePathFor(def))
	if err != nil {
		return err
	}
	return toml.NewEncoder(file).Encode(def)
}

//DeleteImageFor deletes the pre-image for this entity.
func DeleteImageFor(def EntityDefinition, dir ImageDir) error {
	err := os.Remove(dir.imagePathFor(def))
	//ignore does-not-exist error
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
