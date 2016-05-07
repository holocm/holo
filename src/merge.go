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
	"sort"
	"strings"
)

//MergeError is used by Merge().
type MergeError struct {
	Field    string
	EntityID string
	Value1   interface{}
	Value2   interface{}
}

//MergeError implements the error interface.
func (e MergeError) Error() string {
	return fmt.Sprintf("conflicting %s for %s (%v vs. %v)",
		e.Field, e.EntityID, e.Value1, e.Value2,
	)
}

//Merge implements the EntityDefinition interface.
func (g *GroupDefinition) Merge(other EntityDefinition, method MergeMethod) (EntityDefinition, []error) {
	//start by cloning `other`
	if other.EntityID() != g.EntityID() {
		panic("tried to merge entities with different IDs")
	}
	result := *(other.(*GroupDefinition))

	//merge attributes
	var e []error
	if g.GID != 0 {
		if result.GID != 0 && result.GID != g.GID {
			e = append(e, &MergeError{"GID", g.EntityID(), result.GID, g.GID})
		}
		result.GID = g.GID
	}

	//with MergeNumericIDOnly, only the GID may be merged
	if method == MergeNumericIDOnly {
		//we need all the other attributes from `u`, so it's easier to just take another copy
		theResult := *g
		theResult.GID = result.GID
		return &theResult, e
	}

	//the system flag can be set by any side without causing a merge conflict
	result.System = result.System || g.System

	return &result, e
}

//Merge implements the EntityDefinition interface.
func (u *UserDefinition) Merge(other EntityDefinition, method MergeMethod) (EntityDefinition, []error) {
	//start by cloning `other`
	if other.EntityID() != u.EntityID() {
		panic("tried to merge entities with different IDs")
	}
	result := *(other.(*UserDefinition))

	//merge attributes
	var e []error
	if u.UID != 0 {
		if result.UID != 0 && result.UID != u.UID {
			e = append(e, &MergeError{"UID", u.EntityID(), result.UID, u.UID})
		}
		result.UID = u.UID
	}

	//with MergeNumericIDOnly, only the UID may be merged
	if method == MergeNumericIDOnly {
		//we need all the other attributes from `u`, so it's easier to just take another copy
		theResult := *u
		theResult.UID = result.UID
		return &theResult, e
	}

	if u.Home != "" {
		if result.Home != "" && result.Home != u.Home {
			e = append(e, &MergeError{"home directory", u.EntityID(), result.Home, u.Home})
		}
		result.Home = u.Home
	}
	if u.Group != "" {
		if result.Group != "" && result.Group != u.Group {
			e = append(e, &MergeError{"login group", u.EntityID(), result.Group, u.Group})
		}
		result.Group = u.Group
	}
	if u.Shell != "" {
		if result.Shell != "" && result.Shell != u.Shell {
			e = append(e, &MergeError{"login shell", u.EntityID(), result.Shell, u.Shell})
		}
		result.Shell = u.Shell
	}

	//comment is assumed to be informational only, so no merge conflict arises
	if u.Comment != "" {
		result.Comment = u.Comment
	}

	//the system flag can be set by any side without causing a merge conflict
	result.System = result.System || u.System

	//auxiliary groups can always be added, but only under the
	//MergeWhereCompatible method
	if method == MergeEmptyOnly {
		if len(result.Groups) > 0 && len(u.Groups) > 0 {
			sort.Strings(result.Groups)
			sort.Strings(u.Groups)
			resultGroups := strings.Join(result.Groups, ",")
			calleeGroups := strings.Join(u.Groups, ",")
			if resultGroups != calleeGroups {
				e = append(e, &MergeError{"auxiliary groups", u.EntityID(), resultGroups, calleeGroups})
			}
		}
		if len(u.Groups) > 0 {
			result.Groups = u.Groups
		}
	} else {
		for _, group := range u.Groups {
			result.Groups, _ = appendIfMissing(result.Groups, group)
		}
	}
	//make sure that the groups list is always sorted (esp. for reproducible test output)
	sort.Strings(result.Groups)

	return &result, e
}

func appendIfMissing(list []string, value string) (newList []string, changed bool) {
	for _, element := range list {
		if element == value {
			return list, false
		}
	}
	return append(list, value), true
}
