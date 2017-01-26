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
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/holocm/holo/cmd/holo-files/internal/common"
)

//Resource represents a single file in $HOLO_RESOURCE_DIR. The string
//stored in it is the path to the repo file (also accessible as Path()).
type Resource string

//NewResource creates a Resource instance when its path in the file system is
//known.
func NewResource(path string) Resource {
	return Resource(path)
}

//Path returns the path to this resource in the file system.
func (resource Resource) Path() string {
	return string(resource)
}

//EntityPath returns the path to the corresponding entity.
func (resource Resource) EntityPath() string {
	//the optional ".holoscript" suffix appears only on resources
	path := resource.Path()
	path = strings.TrimSuffix(path, ".holoscript")

	//make path relative
	relPath, _ := filepath.Rel(common.ResourceDirectory(), path)
	//remove the disambiguation path element to get to the relPath for the ConfigFile
	//e.g. path     = '/usr/share/holo/files/23-foo/etc/foo.conf'
	//  -> relPath  = '23-foo/etc/foo.conf'
	//  -> relPath  = 'etc/foo.conf'
	segments := strings.SplitN(relPath, fmt.Sprintf("%c", filepath.Separator), 2)
	relPath = segments[1]

	return relPath
}

//Disambiguator returns the disambiguator, i.e. the Path() element before the
//EntityPath() that disambiguates multiple resources for the same entity.
func (resource Resource) Disambiguator() string {
	//make path relative to ResourceDirectory()
	relPath, _ := filepath.Rel(common.ResourceDirectory(), resource.Path())
	//the disambiguator is the first path element in there
	segments := strings.SplitN(relPath, fmt.Sprintf("%c", filepath.Separator), 2)
	return segments[0]
}

//ApplicationStrategy returns the human-readable name for the strategy that
//will be employed to apply this repo file.
func (resource Resource) ApplicationStrategy() string {
	if strings.HasSuffix(resource.Path(), ".holoscript") {
		return "passthru"
	}
	return "apply"
}

//DiscardsPreviousBuffer indicates whether applying this file will discard the
//previous file buffer (and thus the effect of all previous application steps).
//This is used as a hint by the application algorithm to decide whether
//application steps can be skipped completely.
func (resource Resource) DiscardsPreviousBuffer() bool {
	return resource.ApplicationStrategy() == "apply"
}

//ApplyTo applies this Resource to a file buffer, as part of the `holo apply`
//algorithm. Regular repofiles will replace the file buffer, while a holoscript
//will be executed on the file buffer to obtain the new buffer.
func (resource Resource) ApplyTo(buffer *common.FileBuffer) (*common.FileBuffer, error) {
	if resource.ApplicationStrategy() == "apply" {
		return common.NewFileBuffer(resource.Path(), buffer.BasePath)
	}

	//application of a holoscript requires file contents
	buffer, err := buffer.ResolveSymlink()
	if err != nil {
		return nil, err
	}

	//run command, fetch result file into buffer (not into the entity
	//directly, in order not to corrupt the file there if the script run fails)
	var stdout bytes.Buffer
	cmd := exec.Command(resource.Path())
	cmd.Stdin = bytes.NewBuffer(buffer.Contents)
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("execution of %s failed: %s", resource.Path(), err.Error())
	}

	//result is the stdout of the script
	return common.NewFileBufferFromContents(stdout.Bytes(), buffer.BasePath), nil
}

//Resources holds a slice of Resource instances, and implements some methods
//to satisfy the sort.Interface interface.
type Resources []Resource

func (f Resources) Len() int           { return len(f) }
func (f Resources) Less(i, j int) bool { return f[i].Disambiguator() < f[j].Disambiguator() }
func (f Resources) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
