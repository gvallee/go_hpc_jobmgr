// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package mpich

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gvallee/go_exec/pkg/advexec"
	"github.com/gvallee/go_util/pkg/util"
)

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

func parseMPICHInfoOutputForVersion(output string) (string, error) {
	targetLineIdx := 1
	lines := strings.Split(output, "\n")
	if !strings.Contains(lines[targetLineIdx], "Version:") {
		return "", fmt.Errorf("invalid output format")
	}
	tokens := strings.Split(lines[targetLineIdx], "Version:")
	if len(tokens) != 2 {
		return "", fmt.Errorf("invalid format: %s", lines[targetLineIdx])
	}
	version := strings.TrimPrefix(tokens[1], "Version:")
	version = strings.ReplaceAll(version, " ", "")
	version = strings.ReplaceAll(version, "\t", "")
	version = strings.TrimRight(version, "\n")
	return version, nil
}

// DetectFromDir tries to figure out which version of MPICH is installed in a given directory
func DetectFromDir(dir string, env []string) (string, string, error) {
	targetBin := filepath.Join(dir, "bin", "mpirun")
	if !util.FileExists(targetBin) {
		return "", "", fmt.Errorf("%s does not exist, not an MPICH implementation", targetBin)
	}

	var versionCmd advexec.Advcmd
	versionCmd.BinPath = targetBin
	versionCmd.CmdArgs = append(versionCmd.CmdArgs, "--version")
	versionCmd.ExecDir = filepath.Join(dir, "bin")
	versionCmd.Env = env
	if env == nil {
		newLDPath := filepath.Join(dir, "lib") + ":$LD_LIBRARY_PATH"
		newPath := filepath.Join(dir, "bin") + ":$PATH"
		versionCmd.Env = append(versionCmd.Env, "LD_LIBRARY_PATH="+newLDPath)
		versionCmd.Env = append(versionCmd.Env, "PATH="+newPath)
	}
	res := versionCmd.Run()
	if res.Err != nil {
		return "", "", fmt.Errorf("unable to execute %s --version: %w", targetBin, res.Err)
	}
	version, err := parseMPICHInfoOutputForVersion(res.Stdout)
	if err != nil {
		return "", "", fmt.Errorf("parseOmpiInfoOutputForVersion() failed - %w", err)
	}

	return ID, version, nil
}
