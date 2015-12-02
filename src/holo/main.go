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

	"./plugins"
)

// #include <locale.h>
import "C"

func init() {
	//Holo requires a neutral locale, esp. for deterministic sorting of file paths
	lcAll := C.int(0)
	C.setlocale(lcAll, C.CString("C"))
}

//Note: This line is parsed by the Makefile to get the version string. If you
//change the format, adjust the Makefile too.
var version = "v0.10.0-pre"

const (
	optionApplyForce = iota
	optionScanShort
)

func main() {
	//a command word must be given as first argument
	if len(os.Args) < 2 {
		commandHelp()
		return
	}

	//check that it is a known command word
	var command func([]*plugins.Entity, map[int]bool)
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

	//load configuration
	config := plugins.ReadConfiguration()
	if config == nil {
		//some fatal error occurred - it was already reported, so just exit
		os.Exit(255)
	}

	//ask all plugins to scan for entities
	var entities []*plugins.Entity
	for _, plugin := range config.Plugins {
		pluginEntities := plugin.Scan()
		if pluginEntities == nil {
			//some fatal error occurred - it was already reported, so just exit
			os.Exit(255)
		}
		entities = append(entities, pluginEntities...)
	}

	//build a lookup hash for all known entities (for argument parsing)
	isEntityID := make(map[string]bool, len(entities))
	for _, entity := range entities {
		isEntityID[entity.EntityID()] = true
	}

	//parse command line
	options := make(map[int]bool)
	isEntityIDSelected := make(map[string]bool, len(entities))
	hasUnrecognizedArgs := false

	args := os.Args[2:]
	for _, arg := range args {
		//either it's a known option for this subcommand...
		if value, ok := knownOpts[arg]; ok {
			options[value] = true
			continue
		}
		//...or it must be an entity ID
		if isEntityID[arg] {
			isEntityIDSelected[arg] = true
		} else {
			fmt.Fprintf(os.Stderr, "Unrecognized argument: %s\n", arg)
			hasUnrecognizedArgs = true
		}
	}
	if hasUnrecognizedArgs {
		os.Exit(255)
	}

	//if entities have been selected, limit the entities slice to these
	if len(isEntityIDSelected) > 0 {
		selectedEntities := make([]*plugins.Entity, 0, len(entities))
		for _, entity := range entities {
			if isEntityIDSelected[entity.EntityID()] {
				selectedEntities = append(selectedEntities, entity)
			}
		}
		entities = selectedEntities
	}

	//execute command
	command(entities, options)

	//cleanup
	plugins.CleanupRuntimeCache()
}

func commandHelp() {
	program := os.Args[0]
	fmt.Printf("Usage: %s <operation> [...]\nOperations:\n", program)
	fmt.Printf("    %s apply [-f|--force] [entity ...]\n", program)
	fmt.Printf("    %s diff [entity ...]\n", program)
	fmt.Printf("    %s scan [-s|--short] [entity ...]\n", program)
	fmt.Printf("\nSee `man 8 holo` for details.\n")
}

func commandApply(entities []*plugins.Entity, options map[int]bool) {
	withForce := options[optionApplyForce]
	for _, entity := range entities {
		entity.Apply(withForce)
	}
}

func commandScan(entities []*plugins.Entity, options map[int]bool) {
	isShort := options[optionScanShort]
	for _, entity := range entities {
		if isShort {
			fmt.Println(entity.EntityID())
		} else {
			entity.Report().Print()
		}
	}
}

func commandDiff(entities []*plugins.Entity, options map[int]bool) {
	for _, entity := range entities {
		output, err := entity.RenderDiff()
		if err != nil {
			report := plugins.Report{Action: "diff", Target: entity.EntityID()}
			report.AddError(err.Error())
			report.Print()
		}
		os.Stdout.Write(output)
	}
}
