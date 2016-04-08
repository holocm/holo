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

import (
	"fmt"
	"os"

	"./impl"
)

// #include <locale.h>
import "C"

func init() {
	//Holo requires a neutral locale, esp. for deterministic sorting of file paths
	lcAll := C.int(0)
	C.setlocale(lcAll, C.CString("C"))
}

//this is populated at compile-time, see Makefile
var version string

const (
	optionApplyForce = iota
	optionScanShort
)

//Selector represents a command-line argument that selects entities. The Used
//field tracks whether entities match this selector (to report unrecognized
//selectors).
type Selector struct {
	String string
	Used   bool
}

func main() {
	//a command word must be given as first argument
	if len(os.Args) < 2 {
		commandHelp()
		return
	}

	//check that it is a known command word
	var command func([]*impl.Entity, map[int]bool)
	knownOpts := make(map[string]int)
	switch os.Args[1] {
	case "apply":
		command = commandApply
		knownOpts = map[string]int{"-f": optionApplyForce, "--force": optionApplyForce}
	case "diff":
		command = commandDiff
	case "scan":
		command = commandScan
		knownOpts = map[string]int{"-s": optionScanShort, "--short": optionScanShort}
	case "version", "--version":
		fmt.Println(version)
		return
	default:
		commandHelp()
		return
	}

	//ensure that we're the only Holo instance
	impl.AcquireLockfile()

	//load configuration
	config := impl.ReadConfiguration()
	if config == nil {
		//some fatal error occurred - it was already reported, so just exit
		impl.CleanupRuntimeCache()
		impl.ReleaseLockfile()
		os.Exit(255)
	}

	//parse command line
	options := make(map[int]bool)
	selectors := make([]*Selector, 0, len(os.Args)-2)

	args := os.Args[2:]
	for _, arg := range args {
		//either it's a known option for this subcommand...
		if value, ok := knownOpts[arg]; ok {
			options[value] = true
			continue
		}
		//...or it must be a selector
		selectors = append(selectors, &Selector{String: arg, Used: false})
	}

	//ask all plugins to scan for entities
	var entities []*impl.Entity
	for _, plugin := range config.Plugins {
		pluginEntities := plugin.Scan()
		if pluginEntities == nil {
			//some fatal error occurred - it was already reported, so just exit
			impl.CleanupRuntimeCache()
			impl.ReleaseLockfile()
			os.Exit(255)
		}
		entities = append(entities, pluginEntities...)
	}

	//if there are selectors, check which entities have been selected by them
	if len(selectors) > 0 {
		selectedEntities := make([]*impl.Entity, 0, len(entities))
		for _, entity := range entities {
			isEntitySelected := false
			for _, selector := range selectors {
				if entity.MatchesSelector(selector.String) {
					isEntitySelected = true
					selector.Used = true
					//NOTE: don't break from the selectors loop; we want to
					//look at every selector because this loop also verifies
					//that selectors are valid
				}
			}
			if isEntitySelected {
				selectedEntities = append(selectedEntities, entity)
			}
		}
		entities = selectedEntities
	}

	//were there unrecognized selectors?
	hasUnrecognizedArgs := false
	for _, selector := range selectors {
		if !selector.Used {
			fmt.Fprintf(os.Stderr, "Unrecognized argument: %s\n", selector.String)
			hasUnrecognizedArgs = true
		}
	}
	if hasUnrecognizedArgs {
		impl.CleanupRuntimeCache()
		impl.ReleaseLockfile()
		os.Exit(255)
	}

	//build a lookup hash for all known entities (for argument parsing)
	isEntityID := make(map[string]bool, len(entities))
	for _, entity := range entities {
		isEntityID[entity.EntityID()] = true
	}

	//execute command
	command(entities, options)

	//cleanup
	impl.CleanupRuntimeCache()
	impl.ReleaseLockfile()
}

func commandHelp() {
	program := os.Args[0]
	fmt.Printf("Usage: %s <operation> [...]\nOperations:\n", program)
	fmt.Printf("    %s apply [-f|--force] [entity ...]\n", program)
	fmt.Printf("    %s diff [entity ...]\n", program)
	fmt.Printf("    %s scan [-s|--short] [entity ...]\n", program)
	fmt.Printf("\nSee `man 8 holo` for details.\n")
}

func commandApply(entities []*impl.Entity, options map[int]bool) {
	withForce := options[optionApplyForce]
	for _, entity := range entities {
		entity.Apply(withForce)
	}
}

func commandScan(entities []*impl.Entity, options map[int]bool) {
	isShort := options[optionScanShort]
	for _, entity := range entities {
		if isShort {
			fmt.Println(entity.EntityID())
		} else {
			entity.PrintReport(false)
		}
	}
}

func commandDiff(entities []*impl.Entity, options map[int]bool) {
	for _, entity := range entities {
		output, err := entity.RenderDiff()
		if err != nil {
			impl.Errorf("cannot diff %s: %s", entity.EntityID(), err.Error())
		}
		os.Stdout.Write(output)
	}
}
