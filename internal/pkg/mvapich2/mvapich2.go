// Copyright (c) 2022, NVIDIA CORPORATION. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package mvapich2

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gvallee/go_exec/pkg/advexec"
	"github.com/gvallee/go_hpc_jobmgr/internal/pkg/network"
	"github.com/gvallee/go_hpc_jobmgr/pkg/sys"
	"github.com/gvallee/go_util/pkg/util"
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
	if output == "" {
		return "", fmt.Errorf("empty output from version command")
	}
	lines := strings.Split(output, "\n")
	str := strings.TrimPrefix(lines[0], "MVAPICH2 Version:")
	str = strings.TrimLeft(str, " ")
	return str, nil
}

// DetectFromDir tries to figure out which version of OpenMPI is installed in a given directory
func DetectFromDir(dir string, env []string) (string, string, error) {
	targetBin := filepath.Join(dir, "bin", "mpichversion")
	if !util.FileExists(targetBin) {
		return "", "", fmt.Errorf("%s does not exist, not an MVAPICH2 implementation", targetBin)
	}

	var versionCmd advexec.Advcmd
	versionCmd.BinPath = targetBin
	versionCmd.Env = env
	if env == nil {
		newLDPath := filepath.Join(dir, "lib") + ":$LD_LIBRARY_PATH"
		newPath := filepath.Join(dir, "bin") + ":$PATH"
		versionCmd.Env = append(versionCmd.Env, "LD_LIBRARY_PATH="+newLDPath)
		versionCmd.Env = append(versionCmd.Env, "PATH="+newPath)
	}
	res := versionCmd.Run()
	if res.Err != nil {
		return "", "", fmt.Errorf("unable to execute %s: %w", targetBin, res.Err)
	}
	version, err := parseMVAPICH2InfoOutputForVersion(res.Stdout)
	if err != nil {
		return "", "", fmt.Errorf("parseMVAPICH2InfoOutputForVersion() failed: %w", err)
	}

	return ID, version, nil
}
