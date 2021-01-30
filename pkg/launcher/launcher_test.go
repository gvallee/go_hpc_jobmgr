// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// Copyright (c) 2020, NVIDIA CORPORATION. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package launcher

import (
	"os/exec"
	"testing"

	"github.com/gvallee/go_hpc_jobmgr/pkg/app"
	"github.com/gvallee/go_hpc_jobmgr/pkg/jm"
)

func SlurmLaunchTest(t *testing.T) {
	var appInfo app.Info
	var err error
	appInfo.BinPath, err = exec.LookPath("date")
	appInfo.Name = "date"

	sysCfg, jobmgr, err := Load()
	if err != nil {
		t.Fatalf("unable to load the launcher: %s", err)
	}

	if jobmgr.ID != jm.SlurmID {
		t.Fatalf("Slurm not available, skipping")
	}

	res, _ := Run(&appInfo, nil, nil, &jobmgr, &sysCfg, nil)
	if err != nil {
		t.Fatalf("unable to execute Slurm test: %s", err)
	}

	if !res.Pass {
		t.Fatalf("execution failed: %s", res.Note)
	}
}
