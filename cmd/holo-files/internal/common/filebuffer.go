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

package common

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

//FileBuffer represents the contents of a file. It is used in holo.Apply() as
//an intermediary product of application steps.
type FileBuffer struct {
	//set only for regular files
	Contents []byte
	//set only for symlinks
	SymlinkTarget string
	//used by ResolveSymlink (see doc over there)
	BasePath string
}

//NewFileBuffer creates a FileBuffer object by reading the manageable file at
//the given path. The basePath is stored in the FileBuffer for use in
//holo.FileBuffer.ResolveSymlink().
func NewFileBuffer(path string, basePath string) (*FileBuffer, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}

	//a manageable file is either a symlink...
	if IsFileInfoASymbolicLink(info) {
		target, err := os.Readlink(path)
		if err != nil {
			return nil, err
		}
		return &FileBuffer{
			Contents:      nil,
			SymlinkTarget: target,
			BasePath:      basePath,
		}, nil
	}

	//...or a regular file
	if info.Mode().IsRegular() {
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
		return &FileBuffer{
			Contents:      contents,
			SymlinkTarget: "",
			BasePath:      basePath,
		}, nil
	}

	//other types of files are not acceptable
	return nil, &os.PathError{
		Op:   "holo.NewFileBuffer",
		Path: path,
		Err:  errors.New("not a manageable file"),
	}
}

//NewFileBufferFromContents creates a file buffer containing the given byte
//array. The basePath is stored in the FileBuffer for use in
//holo.FileBuffer.ResolveSymlink().
func NewFileBufferFromContents(fileContents []byte, basePath string) *FileBuffer {
	return &FileBuffer{
		Contents:      fileContents,
		SymlinkTarget: "",
		BasePath:      basePath,
	}
}

func (fb *FileBuffer) Write(path string) error {
	//(check that we're not attempting to overwrite unmanageable files
	info, err := os.Lstat(path)
	if err != nil && !os.IsNotExist(err) {
		//abort because the target location could not be statted
		return err
	}
	if err == nil {
		if !(info.Mode().IsRegular() || IsFileInfoASymbolicLink(info)) {
			return &os.PathError{
				Op:   "holo.FileBuffer.Write",
				Path: path,
				Err:  errors.New("target exists and is not a manageable file"),
			}
		}
	}

	//before writing to the target, remove what was there before
	err = os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	//a manageable file is either a regular file...
	if fb.Contents != nil {
		return ioutil.WriteFile(path, fb.Contents, 600)
	}
	//...or a symlink
	return os.Symlink(fb.SymlinkTarget, path)
}

//ResolveSymlink takes a FileBuffer that contains a symlink, resolves it and
//returns a new FileBuffer containing the contents of the symlink target. This
//operation is used by application strategies that require text input. If
//the given FileBuffer contains file contents, the same buffer is returned
//unaltered.
//
//It uses the FileBuffer's BasePath to resolve relative symlinks. Since
//file buffers are usually written to the target path of a `holo apply`
//operation, the BasePath is most likely the target path.
func (fb *FileBuffer) ResolveSymlink() (*FileBuffer, error) {
	//if the buffer has contents already, we can use that
	if fb.Contents != nil {
		return fb, nil
	}

	//if the symlink target is relative, resolve it
	target := fb.SymlinkTarget
	if !filepath.IsAbs(target) {
		baseDir := filepath.Dir(fb.BasePath)
		target = filepath.Join(baseDir, target)
	}

	//read the contents of the target file (NOTE: It's tempting to just use
	//NewFileBuffer here, but that might give us another FileBuffer with a
	//symlink in it, and this time the symlink target might not resolve
	//correctly against the original BasePath. So we explicitly read the file.)
	contents, err := ioutil.ReadFile(target)
	if err != nil {
		return nil, err
	}
	return NewFileBufferFromContents(contents, fb.BasePath), nil
}

//EqualTo returns whether two file buffers have the same content (or link target).
func (fb *FileBuffer) EqualTo(other *FileBuffer) bool {
	if fb.Contents != nil {
		return bytes.Equal(fb.Contents, other.Contents)
	}
	return fb.SymlinkTarget == other.SymlinkTarget
}
