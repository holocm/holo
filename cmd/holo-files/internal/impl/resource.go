/*******************************************************************************
*
* Copyright 2015 Stefan Majewsky <majewsky@gmx.net>
* Copyright 2017 Luke Shumaker <lukeshu@parabola.nu>
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

//Resource represents a single file in $HOLO_RESOURCE_DIR.
type Resource interface {
	// Path returns the path to this resource in the file system.
	Path() string

	// Disambiguator returns the disambiguator, i.e. the Path()
	// element before the EntityPath() that disambiguates multiple
	// resources for the same entity.
	Disambiguator() string

	// EntityPath returns the path to the corresponding entity.
	EntityPath() string

	// ApplicationStrategy returns the human-readable name for the
	// strategy that will be employed to apply this resource.
	ApplicationStrategy() string

	// DiscardsPreviousBuffer indicates whether applying this
	// resource will discard the previous file buffer (and thus
	// the effect of all previous resources).  This is used as a
	// hint by the application algorithm to decide whether
	// application steps can be skipped completely.
	DiscardsPreviousBuffer() bool

	// ApplyTo applies this Resource to a file buffer, as part of
	// the `holo apply` algorithm.
	ApplyTo(entityBuffer common.FileBuffer) (common.FileBuffer, error)
}

type rawResource struct {
	path          string
	disambiguator string
	entityPath    string
}

//NewResource creates a Resource instance when its path in the file system is
//known.
func NewResource(path string) Resource {
	relPath, _ := filepath.Rel(common.ResourceDirectory(), path)
	segments := strings.SplitN(relPath, string(filepath.Separator), 2)
	raw := rawResource{
		path:          path,
		disambiguator: segments[0],
		entityPath:    strings.TrimSuffix(segments[1], ".holoscript"),
	}
	if strings.HasSuffix(raw.path, ".holoscript") {
		return Holoscript{raw}
	}
	return StaticResource{raw}
}

//Path returns the path to this resource in the file system.
func (resource rawResource) Path() string {
	return resource.path
}

//EntityPath returns the path to the corresponding entity.
func (resource rawResource) EntityPath() string {
	return resource.entityPath
}

//Disambiguator returns the disambiguator, i.e. the Path() element before the
//EntityPath() that disambiguates multiple resources for the same entity.
func (resource rawResource) Disambiguator() string {
	return resource.disambiguator
}

// StaticResource is a Resource that is a plain static file that
// replaces the current version of the entity.
type StaticResource struct{ rawResource }

// ApplicationStrategy implements the Resource interface.
func (resource StaticResource) ApplicationStrategy() string { return "apply" }

// DiscardsPreviousBuffer implements the Resource interface.
func (resource StaticResource) DiscardsPreviousBuffer() bool { return true }

// ApplyTo implements the Resource interface.
func (resource StaticResource) ApplyTo(entityBuffer common.FileBuffer) (common.FileBuffer, error) {
	if true { // bogus indentation to make the patch cleaner
		resourceBuffer, err := common.NewFileBuffer(resource.Path())
		if err != nil {
			return common.FileBuffer{}, err
		}
		entityBuffer.Contents = resourceBuffer.Contents
		entityBuffer.Mode = (entityBuffer.Mode &^ os.ModeType) | (resourceBuffer.Mode & os.ModeType)

		//since Linux disregards mode flags on symlinks and always reports 0777 perms,
		//normalize the mode thusly to make FileBuffer.EqualTo() work reliably
		if entityBuffer.Mode&os.ModeSymlink != 0 {
			entityBuffer.Mode = os.ModeSymlink | os.ModePerm
		}
		return entityBuffer, nil
	}
	panic("not reached")
}

// Holoscript is a Resource that is a script that edits the current
// version of the entity.
type Holoscript struct{ rawResource }

// ApplicationStrategy implements the Resource interface.
func (resource Holoscript) ApplicationStrategy() string { return "passthru" }

// DiscardsPreviousBuffer implements the Resource interface.
func (resource Holoscript) DiscardsPreviousBuffer() bool { return false }

// ApplyTo implements the Resource interface.
func (resource Holoscript) ApplyTo(entityBuffer common.FileBuffer) (common.FileBuffer, error) {
	//application of a holoscript requires file contents
	entityBuffer, err := entityBuffer.ResolveSymlink()
	if err != nil {
		return common.FileBuffer{}, err
	}

	//run command, fetch result file into buffer (not into the entity
	//directly, in order not to corrupt the file there if the script run fails)
	var stdout bytes.Buffer
	cmd := exec.Command(resource.Path())
	cmd.Stdin = strings.NewReader(entityBuffer.Contents)
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return common.FileBuffer{}, fmt.Errorf("execution of %s failed: %s", resource.Path(), err.Error())
	}

	//result is the stdout of the script
	entityBuffer.Mode &^= os.ModeType
	entityBuffer.Contents = stdout.String()
	return entityBuffer, nil
}

//Resources holds a slice of Resource instances, and implements some methods
//to satisfy the sort.Interface interface.
type Resources []Resource

func (f Resources) Len() int           { return len(f) }
func (f Resources) Less(i, j int) bool { return f[i].Disambiguator() < f[j].Disambiguator() }
func (f Resources) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
