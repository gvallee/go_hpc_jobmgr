// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package mpich

import "fmt"

const (
	// VersionTag is the tag used to refer to the MPI version in MPICH template(s)
	VersionTag = "MPICHVERSION"
	// URLTag is the tag used to refer to the MPI URL in MPICH template(s)
	URLTag = "MPICHURL"
	// TarballTag is the tag used to refer to the MPI tarball in MPICH template(s)
	TarballTag = "MPICHTARBALL"
	// ID is the internal ID for MPICH
	ID = "mpich"
)

// GetExtraMpirunArgs returns the extra mpirun arguments required by MPICH for a specific configuration
func GetExtraMpirunArgs() []string {
	var extraArgs []string
	return extraArgs
}

// GetConfigureExtraArgs returns the extra arguments required to configure MPICH
func GetConfigureExtraArgs() []string {
	var extraArgs []string
	return extraArgs
}

// DetectFromDir tries to figure out which version of MPICH is installed in a given directory
func DetectFromDir(dir string, env []string) (string, string, error) {
	return "", "", fmt.Errorf("not implemented")
}
