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

//#include <sys/types.h>
//#include <pwd.h>
import "C"
import (
	"errors"
	"os"
	"path/filepath"
)

// User represents a user account on the system. The methods on this struct
// inspect and modify this user's authorized_keys file.
type User struct {
	Name string
	Home string
	UID  int
	GID  int
}

var rootDir = os.Getenv("HOLO_ROOT_DIR") //determines whether we run in unit-test mode

// NewUser returns the User with the given name.
func NewUser(name string) (*User, error) {
	if rootDir == "/" {
		return newUserActual(name)
	}
	return newUserMock(name)
}

func newUserActual(name string) (*User, error) {
	//find UID, GID, home directory with getpwnam()
	structPasswd, err := C.getpwnam(C.CString(name))
	if err != nil {
		return nil, err
	}
	return &User{
		Name: name,
		Home: C.GoString(structPasswd.pw_dir),
		UID:  int(structPasswd.pw_uid),
		GID:  int(structPasswd.pw_gid),
	}, nil
}

func newUserMock(name string) (*User, error) {
	//in testing mode, don't check the system user database;
	//assume all users have /home/$username as home directory
	homeDirPath := filepath.Join(rootDir, "home", name)
	fi, err := os.Stat(homeDirPath)
	if err != nil {
		return nil, err
	}
	if !fi.Mode().IsDir() {
		return nil, errors.New("not a directory: " + homeDirPath)
	}
	return &User{
		Name: name,
		Home: homeDirPath,
		UID:  -1,
		GID:  -1,
	}, nil
}

// KeyFile returns a KeyFile struct for the authorized_keys file for this user.
func (u *User) KeyFile() KeyFile {
	return KeyFile(filepath.Join(u.Home, ".ssh/authorized_keys"))
}

// CheckPermissions checks the permissions on the user's .ssh directory and
// authorized_keys file.
func (u *User) CheckPermissions() error {
	pathHome := u.Home
	pathDssh := filepath.Join(pathHome, ".ssh")
	pathKeys := filepath.Join(pathDssh, "authorized_keys")

	//no chown when running in test mode
	if rootDir == "/" {
		err := os.Chown(pathHome, u.UID, u.GID)
		if err != nil {
			return err
		}
		err = os.Chown(pathDssh, u.UID, u.GID)
		if err != nil {
			return err
		}
		err = os.Chown(pathKeys, u.UID, u.GID)
		if err != nil {
			return err
		}
	}
	err := os.Chmod(pathDssh, 0700)
	if err != nil {
		return err
	}
	return os.Chmod(pathKeys, 0600)
}
