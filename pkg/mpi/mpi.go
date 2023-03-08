// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.
// Copyright (c) 2001-2023, The Ohio State University. All rights reserved.

package mpi

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"

	"github.com/gvallee/go_exec/pkg/manifest"
	"github.com/gvallee/go_hpc_jobmgr/internal/pkg/mpich"
	"github.com/gvallee/go_hpc_jobmgr/internal/pkg/mvapich2"
	"github.com/gvallee/go_hpc_jobmgr/internal/pkg/network"
	"github.com/gvallee/go_hpc_jobmgr/internal/pkg/openmpi"
	"github.com/gvallee/go_hpc_jobmgr/pkg/app"
	"github.com/gvallee/go_hpc_jobmgr/pkg/implem"
	"github.com/gvallee/go_hpc_jobmgr/pkg/sys"
)

// Config represents a configuration of MPI for a target platform
// todo: revisit this, i do not think we actually need it, i think it would make everything
// easier if we were dealing with the different elements separately
type Config struct {
	// Implem gathers information about the MPI implementation to use
	Implem implem.Info

	// UserMpirunArgs is a list of extra arguments defined by the user to pass to the mpirun commands
	UserMpirunArgs []string
}

// GetPathToMpirun returns the path to mpirun based a configuration of MPI
func GetPathToMpirun(mpiCfg *implem.Info) (string, error) {
	// Sanity checks
	if mpiCfg == nil {
		return "", fmt.Errorf("invalid parameter(s)")
	}

	path := filepath.Join(mpiCfg.InstallDir, "bin", "mpirun")

	// the path to mpiexec is something like <path_to_mpi_install/bin/mpiexec> and we need <path_to_mpi_install>
	basedir := filepath.Dir(path)
	basedir = filepath.Join(basedir, "..")
	err := CheckIntegrity(basedir)
	if err != nil {
		return path, err
	}

	return path, nil
}

// GetMpirunArgs returns the arguments required by a mpirun
func GetMpirunArgs(myHostMPICfg *implem.Info, app *app.Info, sysCfg *sys.Config, netCfg *network.Config, mpirunArgs []string) ([]string, error) {
	var extraArgs []string

	// We really do not want to do this but MPICH is being picky about args so for now, it will do the job.
	switch myHostMPICfg.ID {
	case implem.OMPI:
		extraArgs = append(extraArgs, openmpi.GetExtraMpirunArgs(sysCfg, netCfg, mpirunArgs)...)
	case implem.MVAPICH2:
		extraArgs = append(extraArgs, mvapich2.GetExtraMpirunArgs(sysCfg, netCfg, mpirunArgs)...)
	}

	return extraArgs, nil
}

// CheckIntegrity checks if a given installation of MPI has been compromised
func CheckIntegrity(basedir string) error {
	log.Println("* Checking intergrity of MPI...")

	mpiManifest := filepath.Join(basedir, "mpi.MANIFEST")
	return manifest.Check(mpiManifest)
}

// Detect figures out the details about the default MPI implementation
// that is available
func Detect() (*implem.Info, error) {
	mpirunPath, err := exec.LookPath("mpirun")
	if err != nil {
		return nil, err
	}

	mpiInfo := new(implem.Info)
	mpiBinDir := filepath.Dir(mpirunPath)
	// We assume that MPI was not installed in a system directory where binaries
	// and libraries are in totally different directories
	if filepath.Base(mpiBinDir) != "bin" {
		return nil, fmt.Errorf("%s is not a valid MPI installation", mpiBinDir)
	}
	mpiInfo.InstallDir = filepath.Dir(mpiBinDir)

	return mpiInfo, nil
}

func DetectFromDir(dir string) (implem.Info, error) {
	var m implem.Info
	id, version, err := openmpi.DetectFromDir(dir, nil)
	if err == nil {
		m.ID = id
		m.Version = version
		m.InstallDir = dir
		return m, nil
	}
	// Always check for MVAPICH before MPICH since they share some code, otherwise MVAPICH is not correctly detected
	id, version, err = mvapich2.DetectFromDir(dir, nil)
	if err == nil {
		m.ID = id
		m.Version = version
		m.InstallDir = dir
		return m, nil
	}
	id, version, err = mpich.DetectFromDir(dir, nil)
	if err == nil {
		m.ID = id
		m.Version = version
		m.InstallDir = dir
		return m, nil
	}

	return m, fmt.Errorf("unable to detect any supported MPI implementation from %s", dir)
}
