/*******************************************************************************
*
* Copyright 2020 Peter Werner <peter.wr@protonmail.com>
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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// RunGenerators executes all generators in the generator directory
// and changes the resource path of plugins for which files were
// generated to.
func RunGenerators(config *Configuration) error {
	inputDir := getGenertorsDir()
	if _, err := os.Stat(inputDir); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	targetDir, err := getGeneratorCacheDir()
	if err != nil {
		return fmt.Errorf(
			"couldn't access cache-dir ('%s') for generators: %s",
			targetDir, err,
		)
	}
	fmt.Fprintln(Stdout, "Processing generators...")
	filepath.Walk(inputDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				Warnf(Stderr, "%s: %s", path, err.Error())
				return nil
			}
			if isExecutableFile(info) {
				fmt.Fprintln(Stdout, path)
				if err = runGenerator(path, targetDir); err != nil {
					Errorf(
						Stderr,
						"Running generator failed: %s", err.Error(),
					)
				}
			}
			return nil
		})
	for _, plugin := range config.Plugins {
		if err := updatePluginPaths(plugin, targetDir); err != nil {
			Errorf(Stderr,
				"Failed to perpare generated dir for plugin '%s': %s",
				plugin.id, err.Error(),
			)
		}
	}
	return nil
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
			relPath, _ := filepath.Rel(oldDir, oldFile)
			newFile := filepath.Join(newDir, relPath)
			err = os.Symlink(oldFile, newFile)
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

func runGenerator(fileToRun string, targetDir string) error {
	cmd := exec.Command(fileToRun)
	env := os.Environ()
	env = append(
		env,
		fmt.Sprintf("OUT=%s", targetDir),
	)
	cmd.Env = env
	return cmd.Run()
}

func getGenertorsDir() string {
	return filepath.Join(RootDirectory(), "/usr/share/holo/generators")
}

func getGeneratorCacheDir() (string, error) {
	path, err := prepareDir(RootDirectory(), "/var/tmp/holo/generated")
	if err == nil {
		return path, nil
	}
	path, err = prepareDir(
		os.Getenv("HOLO_CACHE_DIR"), "holo/generated",
	)
	if err == nil {
		return path, nil
	}
	return "", err
}

func prepareDir(pathParts ...string) (string, error) {
	path := filepath.Join(pathParts...)
	if err := os.MkdirAll(path, 0755); err != nil {
		if os.IsExist(err) {
			return path, nil
		}
		return "", err
	}
	return path, nil
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
