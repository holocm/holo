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
)

// RunGenerators executes all generators in the generator directory
// and changes the resource path of plugins for which files were
// generated to.
func RunGenerators(config *Configuration) error {
	inputDir := filepath.Join(RootDirectory(), "/usr/share/holo/generators")
	_, err := os.Stat(inputDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	targetDir := filepath.Join(CachePath(), "generated-resources")
	err = os.Mkdir(targetDir, 0777)
	if err != nil && !os.IsExist(err) {
		return err
	}

	runGenerators(inputDir, targetDir)
	for _, plugin := range config.Plugins {
		err := updatePluginPaths(plugin, targetDir)
		if err != nil {
			Errorf(Stderr,
				"Failed to perpare generated dir for plugin '%s': %s",
				plugin.id, err.Error(),
			)
		}
	}
	return nil
}

func runGenerators(inputDir string, targetDir string) {
	filepath.Walk(inputDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				Warnf(Stderr, "%s: %s", path, err.Error())
				return nil
			}
			if isExecutableFile(info) {
				out, err := runGenerator(path, targetDir)
				// Keep silent unless an error occurred or generator has
				// printed output.
				if err != nil || len(out) > 0 {
					shortPath, _ := filepath.Rel(inputDir, path)
					fmt.Fprintf(os.Stdout, "Ran generator %s\n", shortPath)
					fmt.Fprintf(os.Stdout, "     found at %s\n", path)
					Stdout.Write(out)
					if err != nil {
						Errorf(Stderr, err.Error())
					}
				}
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

func runGenerator(fileToRun string, targetDir string) ([]byte, error) {
	//prepare a cache directory with a unique name for the generator
	generatorID := sha256.Sum256([]byte(fileToRun))
	cacheDir := filepath.Join(CachePath(), string(generatorID[:]))
	err := os.Mkdir(cacheDir, 0777)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(fileToRun)
	cmd.Env = append(os.Environ(),
		"HOLO_CACHE_DIR="+cacheDir,
		"HOLO_RESOURCE_ROOT="+filepath.Join(RootDirectory(), "/usr/share/holo"),
		"OUT="+targetDir,
	)
	return cmd.CombinedOutput()
}

func isExecutableFile(stat os.FileInfo) bool {
	mode := stat.Mode()
	if !mode.IsRegular() {
		return false
	}
	if (mode & 0111) == 0 {
		return false
	}
	return true
}
