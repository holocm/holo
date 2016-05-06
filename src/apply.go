/*******************************************************************************
*
* Copyright 2015-2016 Stefan Majewsky <majewsky@gmx.net>
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
	"os"
	"os/exec"
	"strconv"
	"strings"
)

//Apply implements the EntityDefinition interface.
func (g *GroupDefinition) Apply(provisioned EntityDefinition) error {
	//normalize input
	if provisioned != nil && !provisioned.IsProvisioned() {
		provisioned = nil
	}

	//assemble arguments
	var args []string
	if provisioned == nil && g.System {
		args = append(args, "--system")
	}
	if g.GID > 0 {
		args = append(args, "--gid", strconv.Itoa(g.GID))
	}
	args = append(args, g.Name)

	//call groupadd/groupmod
	command := "groupmod"
	if provisioned == nil {
		command = "groupadd"
	}
	return ExecProgramOrMock(command, args...)
}

//Cleanup implements the EntityDefinition interface.
func (g *GroupDefinition) Cleanup() error {
	return ExecProgramOrMock("groupdel", g.Name)
}

//Apply implements the EntityDefinition interface.
func (u *UserDefinition) Apply(provisioned EntityDefinition) error {
	//normalize input
	if provisioned != nil && !provisioned.IsProvisioned() {
		provisioned = nil
	}

	//assemble arguments
	var args []string
	if provisioned == nil && u.System {
		args = append(args, "--system")
	}
	if u.UID > 0 {
		args = append(args, "--uid", strconv.Itoa(u.UID))
	}
	if u.Comment != "" {
		args = append(args, "--comment", u.Comment)
	}
	if u.Home != "" {
		//yay for consistency
		if provisioned == nil {
			args = append(args, "--home-dir", u.Home)
		} else {
			args = append(args, "--home", u.Home)
		}
	}
	if u.Group != "" {
		args = append(args, "--gid", u.Group)
	}
	if len(u.Groups) > 0 {
		args = append(args, "--groups", strings.Join(u.Groups, ","))
	}
	if u.Shell != "" {
		args = append(args, "--shell", u.Shell)
	}
	args = append(args, u.Name)

	//call useradd/usermod
	command := "usermod"
	if provisioned == nil {
		command = "useradd"
	}
	return ExecProgramOrMock(command, args...)
}

//Cleanup implements the EntityDefinition interface.
func (u *UserDefinition) Cleanup() error {
	return ExecProgramOrMock("userdel", u.Name)
}

//ExecProgramOrMock is a wrapper around exec.Command().Run() that, if run in a
//test environment, only prints the command line instead of executing the
//command.
func ExecProgramOrMock(command string, arguments ...string) (err error) {
	mock := os.Getenv("HOLO_ROOT_DIR") != "/"
	if mock {
		fmt.Printf("MOCK: %s %s\n", command, shellEscapeArgs(arguments))
		return nil
	}
	cmd := exec.Command(command, arguments...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func shellEscapeArgs(arguments []string) string {
	//a puny caricature of an actual shell-escape
	var escapedArgs []string
	for _, arg := range arguments {
		if arg == "" || strings.Contains(arg, " ") {
			arg = fmt.Sprintf("'%s'", arg)
		}
		escapedArgs = append(escapedArgs, arg)
	}
	return strings.Join(escapedArgs, " ")
}
