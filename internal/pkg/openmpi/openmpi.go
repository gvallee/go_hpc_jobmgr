// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package openmpi

import (
	"github.com/gvallee/go_hpc_jobmgr/internal/pkg/sys"
)

const (
	// VersionTag is the tag used to refer to the MPI version in Open MPI template(s)
	VersionTag = "OMPIVERSION"

	// URLTag is the tag used to refer to the MPI URL in Open MPI template(s)
	URLTag = "OMPIURL"

	// TarballTag is the tag used to refer to the MPI tarball in Open MPI template(s)
	TarballTag = "OMPITARBALL"
)

// GetExtraMpirunArgs returns the set of arguments required for the mpirun command for the target platform
func GetExtraMpirunArgs(sys *sys.Config) []string {
	var extraArgs []string
	/*
		if sys.IBEnabled {
			extraArgs = append(extraArgs, "--mca")
			extraArgs = append(extraArgs, "btl")
			extraArgs = append(extraArgs, "openib,self,vader")
		}
	*/

	return extraArgs
}
