// Copyright (c) 2022, NVIDIA CORPORATION. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package mvapich2

import (
	"github.com/gvallee/go_hpc_jobmgr/internal/pkg/network"
	"github.com/gvallee/go_hpc_jobmgr/pkg/sys"
)

const (
	// VersionTag is the tag used to refer to the MPI version in MVAPICH2 template(s)
	VersionTag = "MVAPICH2VERSION"

	// URLTag is the tag used to refer to the MPI URL in MVAPICH2 template(s)
	URLTag = "MVAPICH2URL"

	// TarballTag is the tag used to refer to the MPI tarball in MVAPICH2 template(s)
	TarballTag = "MVAPICH2TARBALL"

	// ID is the internal ID for MVAPICH2
	ID = "mvapich2"
)

// GetExtraMpirunArgs returns the set of arguments required for the mpirun command for the target platform
func GetExtraMpirunArgs(sys *sys.Config, netCfg *network.Config, extraArgs []string) []string {
	return nil
}

func parseMVAPICH2InfoOutputForVersion(output string) (string, error) {
	return "", nil
}

// DetectFromDir tries to figure out which version of OpenMPI is installed in a given directory
func DetectFromDir(dir string, env []string) (string, string, error) {
	return "", "", nil
}
