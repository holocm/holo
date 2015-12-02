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

package plugins

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var rootDirectory string

func init() {
	rootDirectory = os.Getenv("HOLO_ROOT_DIR")
	if rootDirectory == "" {
		rootDirectory = "/"
	}
}

//RootDirectory returns the environment variable $HOLO_ROOT_DIR, or else the
//default value "/".
func RootDirectory() string {
	return rootDirectory
}

//Configuration contains the parsed contents of /etc/holorc.
type Configuration struct {
	Plugins []*Plugin
}

//ReadConfiguration reads the configuration file /etc/holorc.
func ReadConfiguration() *Configuration {
	path := filepath.Join(RootDirectory(), "etc/holorc")

	contents, err := ioutil.ReadFile(path)
	if err != nil {
		r := Report{Action: "read", Target: path}
		r.AddError(err.Error())
		r.Print()
		return nil
	}

	var result Configuration
	lines := strings.SplitN(strings.TrimSpace(string(contents)), "\n", -1)
	for _, line := range lines {
		//ignore comments and empty lines
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		//collect plugin IDs
		if strings.HasPrefix(line, "plugin ") {
			pluginID := strings.TrimSpace(strings.TrimPrefix(line, "plugin"))
			if strings.Contains(pluginID, "=") {
				fields := strings.SplitN(pluginID, "=", 2)
				result.Plugins = append(result.Plugins,
					NewPluginWithExecutablePath(fields[0], fields[1]),
				)
			} else {
				result.Plugins = append(result.Plugins, NewPlugin(pluginID))
			}
			continue
		} else {
			//unknown line
			r := Report{Action: "read", Target: path}
			r.AddError("unknown command: %s", line)
			r.Print()
			return nil
		}
	}

	//check existence of resource directories
	errorReport := Report{Action: "Errors occurred during", Target: "plugin discovery"}
	hasError := false
	for _, plugin := range result.Plugins {
		dir := plugin.ResourceDirectory()
		fi, err := os.Stat(dir)
		switch {
		case err != nil:
			errorReport.AddError("Cannot open %s: %s", dir, err.Error())
			hasError = true
		case !fi.IsDir():
			errorReport.AddError("Cannot open %s: not a directory!", dir)
			hasError = true
		}
	}
	if hasError {
		errorReport.Print()
		return nil
	}

	//ensure existence of cache and state directories
	for _, plugin := range result.Plugins {
		dirs := []string{plugin.CacheDirectory(), plugin.StateDirectory()}
		for _, dir := range dirs {
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				errorReport.AddError(err.Error())
				hasError = true
			}
		}
	}
	if hasError {
		errorReport.Print()
		return nil
	}

	return &result
}
