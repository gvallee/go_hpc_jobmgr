// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package implem

import (
	"fmt"

	"github.com/gvallee/go_hpc_jobmgr/internal/pkg/mpich"
	"github.com/gvallee/go_hpc_jobmgr/internal/pkg/openmpi"
)

const (
	// OMPI is the identifier for Open MPI
	OMPI = openmpi.ID

	// MPICH is the identifier for MPICH
	MPICH = mpich.ID
)

// Info gathers all data about a specific MPI implementation
type Info struct {
	// ID is the string idenfifying the MPI implementation
	ID string

	// Version is the version of the MPI implementation
	Version string

	// InstallDir is where the MPI implementation is installed
	InstallDir string
}

// IsMPI checks if information passed in is an MPI implementation
func IsMPI(i *Info) bool {
	if i != nil && (i.ID == OMPI || i.ID == MPICH) {
		return true
	}

	return false
}

// Load figures out the details about a specific implementation of MPI so it can be used later on.
// Users have the option to specify:
// - the install directory, then the function figures out the implementation and version
// - the implementation (e.g., openmpi) and the function figures out where it is installed
// - a few other combinations of these to provide a flexible way to handle various implementation of MPI
// If no suitable implementation can be found, the function returns an error
func (i *Info) Load() error {
	if i.InstallDir != "" && (i.ID == "" || i.Version == "") {
		var err error
		i.ID, i.Version, err = openmpi.DetectFromDir(i.InstallDir)
		if err == nil {
			return nil
		}
		i.ID, i.Version, err = mpich.DetectFromDir(i.InstallDir)
		if err == nil {
			return nil
		}
		return fmt.Errorf("unable to detect MPI implementation from %s", i.InstallDir)
	}
	return nil
}
