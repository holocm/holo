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

package main

//EntityDefinition contains data from a definition file that describes an entity
//(a user account or group). Definitions can also be obtained by scanning the
//user/group databases.
type EntityDefinition interface {
	//TypeName returns the part of the entity ID before the ":", i.e. either
	//"group" or "user".
	TypeName() string
	//GetProvisionedState reads the current state of this entity from the
	//system database (/etc/passwd or /etc/group). The return value has the same
	//concrete type as the callee. If no entity with the same ID exists in
	//there, nil is returned.
	GetProvisionedState() (EntityDefinition, error)
}

//Entity provides a common interface for configuration entities, such as
//configuration files, user accounts and user groups.
type Entity interface {
	//EntityID returns a string that uniquely identifies the entity, usually in
	//the form "type:name". This is how the entity can be addressed as a target
	//in the argument list for "holo apply", e.g. "holo apply /etc/sudoers
	//group:foo" will apply the target file "/etc/sudoers" and the group "foo".
	//Therefore, entity IDs should not contain whitespaces or characters that
	//have a special meaning on the shell.
	EntityID() string
	//PrintReport prints the scan report for this entity on stdout.
	PrintReport()
	//Apply performs the complete application algorithm for the given Entity.
	Apply(withForce bool) (entityWasChanged bool)
	//PrepareDiff creates temporary files that the frontend can use to generate
	//a diff.
	PrepareDiff() (expectedState string, actualState string, e error)
}
