// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package openmpi

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/gvallee/go_exec/pkg/advexec"
	"github.com/gvallee/go_hpc_jobmgr/internal/pkg/sys"
	"github.com/gvallee/go_util/pkg/util"
)

const (
	// VersionTag is the tag used to refer to the MPI version in Open MPI template(s)
	VersionTag = "OMPIVERSION"

	// URLTag is the tag used to refer to the MPI URL in Open MPI template(s)
	URLTag = "OMPIURL"

	// TarballTag is the tag used to refer to the MPI tarball in Open MPI template(s)
	TarballTag = "OMPITARBALL"

	// ID is the internal ID for Open MPI
	ID = "openmpi"
)

// GetExtraMpirunArgs returns the set of arguments required for the mpirun command for the target platform
func GetExtraMpirunArgs(sys *sys.Config) []string {
	var extraArgs []string
	// By default we always prefer UCX rather than openib
	extraArgs = append(extraArgs, "--mca")
	extraArgs = append(extraArgs, "btl")
	extraArgs = append(extraArgs, "^openib")
	extraArgs = append(extraArgs, "--mca")
	extraArgs = append(extraArgs, "pml")
	extraArgs = append(extraArgs, "ucx")
	return extraArgs
}

func parseOmpiInfoOutputForVersion(output string) (string, error) {
	lines := strings.Split(output, "\n")
	if !strings.HasPrefix(lines[0], "Open MPI") {
		return "", fmt.Errorf("invalid output format")
	}
	version := strings.TrimLeft(lines[0], "Open MPI v")
	version = strings.TrimRight(version, "\n")
	return version, nil
}

// DetectFromDir tries to figure out which version of OpenMPI is installed in a given directory
func DetectFromDir(dir string) (string, string, error) {
	targetBin := filepath.Join(dir, "bin", "ompi-info")
	if !util.FileExists(targetBin) {
		return "", "", fmt.Errorf("%s does not exist, not an OpenMPI implementation", targetBin)
	}

	var versionCmd advexec.Advcmd
	versionCmd.BinPath = targetBin
	versionCmd.CmdArgs = append(versionCmd.CmdArgs, "--version")
	res := versionCmd.Run()
	if res.Err != nil {
		log.Printf("unable to run ompi_info: %s; stdout: %s; stderr: %s", res.Err, res.Stdout, res.Stderr)
		return "", "", res.Err
	}
	version, err := parseOmpiInfoOutputForVersion(res.Stdout)
	if err != nil {
		return "", "", err
	}

	return ID, version, nil
}
