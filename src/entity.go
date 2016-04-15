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

const (
	//MergeWhereCompatible can be used as second argument for EntityDefinition.Merge.
	MergeWhereCompatible = false
	//MergeEmptyOnly can be used as second argument for EntityDefinition.Merge.
	MergeEmptyOnly = true
)

//EntityDefinition contains data from a definition file that describes an entity
//(a user account or group). Definitions can also be obtained by scanning the
//user/group databases.
type EntityDefinition interface {
	//TypeName returns the part of the entity ID before the ":", i.e. either
	//"group" or "user".
	TypeName() string
	//EntityID returns exactly that, e.g. "user:john".
	EntityID() string
	//Attributes returns a human-readable stringification of this definition.
	Attributes() string
	//GetProvisionedState reads the current state of this entity from the
	//system database (/etc/passwd or /etc/group). The return value has the same
	//concrete type as the callee. If no entity with the same ID exists in
	//there, nil is returned.
	GetProvisionedState() (EntityDefinition, error)
	//WithSerializableState brings the definition into a safely serializable
	//state, executes the callback, and then restores the original state.
	WithSerializableState(callback func(EntityDefinition))
	//Merge constructs a new EntityDefinition of the same concrete type whose
	//attributes are merged from the callee and the argument. The argument's
	//concrete type must be identical to that of the callee. If both sources
	//have different values set for the same attribute, the callee's value
	//takes precedence, and an error is returned in the second argument.
	//If merge conflicts are not a problem, the error argument may be ignored.
	//
	//If `emptyOnly` is true, only empty arguments may be merged.
	Merge(other EntityDefinition, emptyOnly bool) (EntityDefinition, []error)
	//Apply provisions this entity. The argument indicates the currently
	//provisioned state. If the entity is not provisioned yet, it will be nil.
	//If not nil, the argument's concrete type must match the callee.
	Apply(provisioned EntityDefinition) error
}

//Entity provides a common interface for configuration entities, such as
//configuration files, user accounts and user groups.
type Entity interface {
	//Definition returns the underlying Entity definition.
	Definition() EntityDefinition
	//Orphaned is just temporary. TODO: remove
	IsOrphaned() bool
	//PrintReport prints the scan report for this entity on stdout.
	PrintReport()
	//Apply performs the complete application algorithm for the given Entity.
	Apply(withForce bool) (entityWasChanged bool)
}
