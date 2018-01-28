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
	"os/exec"
	"path/filepath"
	"strings"
)

//Scan reports all entities that exist on stdout.
func Scan() []error {
	//list entries in resource directory
	dir, err := os.Open(resourceDirPath)
	if err != nil {
		return []error{err}
	}
	fis, err := dir.Readdir(-1)
	if err != nil {
		return []error{err}
	}

	//descend into subdirectories
	var errs []error
	entityNameWasSeen := make(map[string]bool)
	for _, fi := range fis {
		if fi.Mode().IsDir() {
			suberrs := scanDirectory(filepath.Join(resourceDirPath, fi.Name()), &entityNameWasSeen)
			errs = append(errs, suberrs...)
		}
	}

	//find orphaned entities (which we once provisioned, but for which there is
	//no source file anymore)
	allEntities, err := ProvisionedEntities()
	if err != nil {
		errs = append(errs, err)
	}
	for _, entityName := range allEntities {
		if !entityNameWasSeen[entityName] {
			fmt.Printf("ENTITY: %s\n", entityName)
			fmt.Println("ACTION: Scrubbing (source file has been deleted)")
		}
	}

	return errs
}

func scanDirectory(path string, entityNameWasSeen *map[string]bool) []error {
	//list entries in directory below resource directory
	dir, err := os.Open(path)
	if err != nil {
		return []error{err}
	}
	fis, err := dir.Readdir(-1)
	if err != nil {
		return []error{err}
	}

	//find key files in this directory
	var errs []error
	for _, fi := range fis {
		if fi.Mode().IsRegular() && strings.HasSuffix(fi.Name(), ".pub") {
			//construct Entity object
			entity, err := NewEntityFromKeyfilePath(filepath.Join(path, fi.Name()))
			if err != nil {
				errs = append(errs, err)
				continue
			}
			(*entityNameWasSeen)[entity.Name] = true

			//get keys contained in this Entity
			keys, err := entity.Keys()
			if err != nil {
				errs = append(errs, err)
				continue
			}

			//calculate fingerprints for all keys
			fingerprints := make([]string, 0, len(keys))
			for _, key := range keys {
				fp, err := getFingerprint(key)
				if err != nil {
					errs = append(errs, err)
				} else {
					fingerprints = append(fingerprints, fp)
				}
			}
			if len(fingerprints) < len(keys) {
				continue //there were errors in the previous loop
			}

			//report entity
			fmt.Printf("ENTITY: %s\n", entity.Name)
			fmt.Printf("SOURCE: %s\n", entity.FilePath)
			fmt.Printf("found in: %s\n", entity.FilePath)
			if len(fingerprints) == 0 {
				fmt.Println("is: empty!")
			} else {
				for _, fingerprint := range fingerprints {
					fmt.Printf("key is: %s\n", fingerprint)
				}
			}
		}
	}

	return errs
}

func getFingerprint(key *Key) (string, error) {
	//`ssh-keygen` cannot read the key from stdin for whatever reason, so use a temporary file
	file, err := ioutil.TempFile("/tmp/", "holo-ssh-key")
	if err != nil {
		return "", err
	}
	path := file.Name()
	fmt.Fprintln(file, key.String())
	err = file.Close()
	if err != nil {
		return "", err
	}

	//run ssh-keygen to compute fingerprint
	cmd := exec.Command("ssh-keygen", "-l", "-f", path)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(buf.Bytes())), nil
}
