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
	"os"
	"path/filepath"
	"sort"

	"github.com/holocm/holo/cmd/holo-files/internal/common"
)

// Entity represents a configuration file that can be provisioned by holo-files.
type Entity struct {
	relPath   string // the entity path relative to the common.TargetDirectory()
	resources Resources
}

// NewEntity creates a Entity instance for which a path is known.
//
//	entity := NewEntity("etc/locale.conf")
func NewEntity(relPath string) *Entity {
	return &Entity{relPath: relPath}
}

// PathIn returns the path to this entity relative to the given directory.
//
//	var (
//	    targetPath      = entity.pathIn(common.TargetDirectory())      // e.g. "/etc/foo.conf"
//	    basePath        = entity.pathIn(common.BaseDirectory())        // e.g. "/var/lib/holo/files/base/etc/foo.conf"
//	    provisionedPath = entity.pathIn(common.ProvisionedDirectory()) // e.g. "/var/lib/holo/files/provisioned/etc/foo.conf"
//	)
func (entity *Entity) PathIn(directory string) string {
	return filepath.Join(directory, entity.relPath)
}

// AddResource registers a new resource in this Entity instance.
func (entity *Entity) AddResource(resource Resource) {
	entity.resources = append(entity.resources, resource)
}

// Resources returns an ordered list of all resources for this Entity.
func (entity *Entity) Resources() Resources {
	sort.Sort(entity.resources)
	return entity.resources
}

// EntityID returns the entity ID for this entity.
func (entity *Entity) EntityID() string {
	return "file:" + entity.PathIn("/")
}

// PrintReport prints the report required by the "scan" operation for this
// entity.
func (entity *Entity) PrintReport() {
	fmt.Printf("ENTITY: %s\n", entity.EntityID())

	if len(entity.resources) == 0 {
		_, strategy, assessment := entity.scanOrphan()
		fmt.Printf("ACTION: Scrubbing (%s)\n", assessment)
		fmt.Printf("%s: %s\n", strategy, entity.PathIn(common.BaseDirectory()))
	} else {
		fmt.Printf("store at: %s\n", entity.PathIn(common.BaseDirectory()))
		for _, resource := range entity.Resources() {
			fmt.Printf("SOURCE: %s\n", resource.Path())
			fmt.Printf("%s: %s\n", resource.ApplicationStrategy(), resource.Path())
		}
	}
}

// ErrNeedForceToOverwrite is used to signal a command message upwards in the
// call chain.
var ErrNeedForceToOverwrite = errors.New("NeedForceToOverwrite")

// ErrNeedForceToRestore is used to signal a command message upwards in the call
// chain.
var ErrNeedForceToRestore = errors.New("NeedForceToRestore")

// Apply applies the entity.
func (entity *Entity) Apply(withForce bool) (skipReport, needForceToOverwrite, needForceToRestore bool) {
	if len(entity.resources) == 0 {
		errs := entity.applyOrphan()
		skipReport = false
		needForceToOverwrite = false
		needForceToRestore = false

		for _, err := range errs {
			fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		}
	} else {
		var err error
		skipReport, err = entity.applyNonOrphan(withForce)

		//special cases for errors that signal command messages
		needForceToOverwrite = err == ErrNeedForceToOverwrite
		needForceToRestore = err == ErrNeedForceToRestore
		if needForceToRestore || needForceToOverwrite {
			err = nil
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		}
	}
	return
}
