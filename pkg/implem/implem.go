// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package implem

const (
	// OMPI is the identifier for Open MPI
	OMPI = "openmpi"

	// MPICH is the identifier for MPICH
	MPICH = "mpich"
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
