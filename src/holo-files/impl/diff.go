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
	"regexp"
	"strings"
	"syscall"

	"../common"
)

type fileType int

const (
	fileUnknown fileType = 0
	fileMissing fileType = 1
	fileRegular fileType = 2
	fileSymlink fileType = 3
)

//RenderDiff creates a unified diff of a target file and its last provisioned
//version, similar to `diff /var/lib/holo/files/provisioned/$FILE $FILE`, but it also
//handles symlinks and missing files gracefully. The output is always a patch
//that can be applied to last provisioned version into the current version.
func (target *TargetFile) RenderDiff() ([]byte, error) {
	fromPath := target.PathIn(common.ProvisionedDirectory())
	toPath := target.PathIn(common.TargetDirectory())

	fromPathToUse, err := checkFile(fromPath)
	if err != nil {
		return nil, err
	}
	toPathToUse, err := checkFile(toPath)
	if err != nil {
		return nil, err
	}

	//run git-diff to obtain the diff
	var buffer bytes.Buffer
	cmd := exec.Command("git", "diff", "--no-index", "--", fromPathToUse, toPathToUse)
	cmd.Stdout = &buffer
	cmd.Stderr = os.Stderr

	//error "exit code 1" is normal for different files, only exit code > 2 means trouble
	err = cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				if status.ExitStatus() == 1 {
					err = nil
				}
			}
		}
	}
	//did a relevant error occur?
	if err != nil {
		return nil, err
	}

	//remove "index <SHA1>..<SHA1> <mode>" lines
	result := buffer.Bytes()
	rx := regexp.MustCompile(`(?m:^index .*$)\n`)
	result = rx.ReplaceAll(result, nil)

	//remove "/var/lib/holo/files/provisioned" from path displays to make it appear like we
	//just diff the target path
	if fromPathToUse == fromPath {
		fromPathQuoted := strings.TrimPrefix(regexp.QuoteMeta(fromPath), "/")
		toPathQuoted := strings.TrimPrefix(regexp.QuoteMeta(toPath), "/")
		toPathTrimmed := strings.TrimPrefix(toPath, "/")

		rx = regexp.MustCompile(`(?m:^)diff --git a/` + fromPathQuoted)
		result = rx.ReplaceAll(result, []byte("diff --git a/"+toPathTrimmed))

		rx = regexp.MustCompile(`(?m:^)diff --git a/` + toPathQuoted + ` b/` + fromPathQuoted)
		result = rx.ReplaceAll(result, []byte("diff --git a/"+toPathTrimmed+" b/"+toPathTrimmed))

		rx = regexp.MustCompile(`(?m:^)--- a/` + fromPathQuoted)
		result = rx.ReplaceAll(result, []byte("--- a/"+toPathTrimmed))
	}

	return result, nil
}

func checkFile(path string) (pathToUse string, returnError error) {
	//check that files are either non-existent (in which case git-diff needs to
	//be given /dev/null instead or manageable (e.g. we can't diff directories
	//or device files)
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "/dev/null", nil
		}
		return path, err
	}

	if !common.IsManageableFileInfo(info) {
		return path, fmt.Errorf("%s is not a manageable file", path)
	}
	return path, nil
}
