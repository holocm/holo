/*******************************************************************************
*
* Copyright 2020 Peter Werner <peter.wr@protonmail.com>
* Copyright 2021 Stefan Majewsky <majewsky@gmx.net>
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
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RunGenerators executes all generators in the generator directory
// and changes the resource path of plugins for which files were
// generated to.
func RunGenerators(config *Configuration) error {
	targetDir := filepath.Join(CachePath(), "generated-resources")
	err := os.Mkdir(targetDir, 0777)
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = runGenerators(targetDir)
	if err != nil {
		return err
	}

	for _, plugin := range config.Plugins {
		err := updatePluginPaths(plugin, targetDir)
		if err != nil {
			return fmt.Errorf("while preparing virtual resource directory for plugin %q: %w", plugin.id, err)
		}
	}
	return nil
}

func runGenerators(targetDir string) error {
	generatorsDir := filepath.Join(RootDirectory(), "/usr/share/holo/generators")
	return filepath.Walk(generatorsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		//NOTE: We don't need to check for executability here. Having a
		//non-executable file in the generators directory just produces an obvious
		//error during exec.Command() down below.
		if info.Mode().IsRegular() {
			return runGenerator(path, targetDir)
		}
		return nil
	})
}

func updatePluginPaths(plugin *Plugin, dir string) error {
	pluginDir := plugin.ResourceDirectory()
	newPluginDir := filepath.Join(dir, plugin.id)
	if info, err := os.Stat(newPluginDir); err == nil && info.IsDir() {
		// Files were generated for this plugin.
		// Fill the plugins directory with existsing static files.
		if err := symlinkFiles(pluginDir, newPluginDir); err != nil {
			return err
		}
		// Change the plugin resource dir to point to the generated dir.
		resource, _ := filepath.Rel(RootDirectory(), dir)
		plugin.SetResourceRoot(resource)
	}
	return nil
}

func symlinkFiles(oldDir string, newDir string) error {
	return filepath.Walk(oldDir,
		func(oldFile string, info os.FileInfo, err error) error {
			if err != nil || oldFile == oldDir {
				return err
			}
			newFile := filepath.Join(newDir, filepathMustRel(oldDir, oldFile))
			err = os.Symlink(filepathMustRel(filepath.Dir(newFile), oldFile), newFile)
			if os.IsExist(err) {
				// newFile already exists. Examine it.
				newFileInfo, err := os.Lstat(newFile)
				if err == nil && info.IsDir() && !newFileInfo.IsDir() {
					// newFile exists but is not a directory.
					// If oldFile is a directory trying to symlink its contents
					// will result in errors. Skip it.
					return filepath.SkipDir
				}
				return nil
			}
			if err != nil {
				return err
			}
			if info.IsDir() {
				// Symlink to dir was created don't check its contents.
				return filepath.SkipDir
			}
			return nil
		})
}

//Like filepath.Rel(), but assumes that no error occurs. This assumption is
//safe if both inputs are absolute paths.
func filepathMustRel(base, target string) string {
	rel, _ := filepath.Rel(base, target)
	return rel
}

func runGenerator(fileToRun, targetDir string) error {
	//prepare a cache directory with a unique name for the generator
	generatorID := sha256.Sum256([]byte(fileToRun))
	cacheDir := filepath.Join(CachePath(), string(generatorID[:]))
	err := os.Mkdir(cacheDir, 0777)
	if err != nil {
		return err
	}

	cmd := exec.Command(fileToRun)
	cmd.Env = append(os.Environ(),
		"HOLO_CACHE_DIR="+cacheDir,
		"HOLO_RESOURCE_ROOT="+filepath.Join(RootDirectory(), "/usr/share/holo"),
		"OUT="+targetDir,
	)

	out, err := cmd.CombinedOutput()
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			Warnf(Stderr, "output from %s: %s", fileToRun, line)
		}
	}
	if err != nil {
		return fmt.Errorf("could not run %s: %w", fileToRun, err)
	}
	return nil
}
