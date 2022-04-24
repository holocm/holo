/*******************************************************************************
*
* Copyright 2017-2018 Luke Shumaker <lukeshu@parabola.nu>
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
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/holocm/holo/cmd/holo-files/internal/common"
)

// Patchfile is a Resource that is a `patch(1)` file that edits the
// current version of the entity.
type Patchfile struct{ rawResource }

// ApplicationStrategy implements the Resource interface.
func (resource Patchfile) ApplicationStrategy() string { return "patch" }

// DiscardsPreviousBuffer implements the Resource interface.
func (resource Patchfile) DiscardsPreviousBuffer() bool { return false }

// ApplyTo implements the Resource interface.
func (resource Patchfile) ApplyTo(entityBuffer common.FileBuffer) (common.FileBuffer, error) {
	// `patch` requires that the file it's operating on be a real
	// file (not a pipe).  So, we'll write entityBuffer to a
	// temporary file, run `patch`, then read it back.

	// We really only normally need 1 temporary file, but:
	//  1. since common.FileBuffer.Write removes the file and then
	//     re-creates it, that's a bit racy
	//  2. The only way to limit patch to operating on a single
	//     file is to name that file on the command line, but
	//     doing that prevents it from unlinking the file, which
	//     prevents type changes.
	//
	// Using a temporary directory lets us easily work around both
	// of these issues.  Unfortunately, this allows the patch to
	// create new files other than the one for the entity we are
	// applying.  However, it can't escape the temporary
	// directory, so we'll just "allow" that, and document that we
	// ignore those files.
	targetDir, err := ioutil.TempDir(os.Getenv("HOLO_CACHE_DIR"), "patch-target.")
	if err != nil {
		return common.FileBuffer{}, err
	}
	defer os.RemoveAll(targetDir)
	targetPath := filepath.Join(targetDir, filepath.Base(entityBuffer.Path))

	// Write entityBuffer to the temporary file
	err = entityBuffer.Write(targetPath)
	if err != nil {
		return common.FileBuffer{}, err
	}

	// Run `patch` on the temporary file
	patchfile, err := filepath.Abs(resource.Path())
	if err != nil {
		return common.FileBuffer{}, err
	}
	cmd := exec.Command("patch",
		"-N",
		"-i", patchfile,
	)
	cmd.Dir = targetDir
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return common.FileBuffer{}, fmt.Errorf("execution failed: %s: %s", strings.Join(cmd.Args, " "), err.Error())
	}

	// Read the result back
	//
	// Allow `patch` to override everything but the filepath:
	//  - file type (changable with git-style "deleted file
	//    mode"/"new file mode" lines, which are implemented by at
	//    least GNU patch, if not in strict POSIX mode)
	//  - file permissions (changable with git-style "new mode"
	//    lines, which are implemented by at least GNU patch)
	//  - UID/GID (I don't know of a patch syntax that does this,
	//    but maybe it will exist in the future)
	//  - contents (obviously)
	targetBuffer, err := common.NewFileBuffer(targetPath)
	if err != nil {
		return common.FileBuffer{}, err
	}
	targetBuffer.Path = entityBuffer.Path
	return targetBuffer, nil
}
