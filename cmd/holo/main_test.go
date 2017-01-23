/*******************************************************************************
*
* Copyright 2017 Luke Shumaker <lukeshu@parabola.nu>
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
	"flag"
	"os"
	"strings"
	"testing"
)

var exit int

// TestMain (this name is required) does the setup and teardown when
// running the coverage-testing-tooled binary.  In our case, "setup"
// is fiddling with `flags.CommandLine` to make the "testing" package
// happy (as Holo doesn't use the "flags" package, but "testing"
// requires its use); and "teardown" means exiting with the exit
// status indicated by our main function.
func TestMain(m *testing.M) {
	var flags []string
	if str, ok := os.LookupEnv("HOLO_TEST_FLAGS"); ok {
		flags = append(flags, strings.Split(str, " ")...)
	}

	flag.CommandLine.Parse(flags)
	_ = m.Run()
	os.Exit(exit)
}

// TestSystem runs the "main" function, since binaries produced
// by `go test -c` don't actually run the normal `main()`.
//
// This function does't actually run `main()` either, instead it runs
// a `Main() int` function that returns an exit code rather than
// calling `os.Exit()`.  This is important because if it calls
// `os.Exit()` then the cover file won't have been written yet.
func TestSystem(t *testing.T) {
	exit = Main()
	// Now that we've finished running, close stdout to prevent
	// the "testing" package from printing a message saying that
	// all tests ran successfully.
	_ = os.Stdout.Close()
	// We can't exit directly from here because the cover file has
	// not yet been written.
}
