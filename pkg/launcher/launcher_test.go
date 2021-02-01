// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// Copyright (c) 2020, NVIDIA CORPORATION. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package launcher

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/gvallee/go_hpc_jobmgr/internal/pkg/job"
	"github.com/gvallee/go_hpc_jobmgr/pkg/jm"
)

var partition = flag.String("partition", "", "Name of Slurm partition to use to run the test")
var scratchDir = flag.String("scratch", "", "Scratch directory to use to execute the test")

func TestSlurmLaunch(t *testing.T) {
	var j job.Job
	var err error
	j.App.BinPath, err = exec.LookPath("date")
	if err != nil {
		t.Fatalf("unable to find path to 'date' binnary")
	}
	j.App.Name = "date"
	j.Partition = *partition

	sysCfg, jobmgr, err := Load()
	if err != nil {
		t.Fatalf("unable to load the launcher: %s", err)
	}
	sysCfg.ScratchDir, err = ioutil.TempDir(*scratchDir, "")
	if err != nil {
		t.Fatalf("unable to create temporary directory: %s", err)
	}
	defer os.RemoveAll(sysCfg.ScratchDir)

	if jobmgr.ID != jm.SlurmID {
		t.Fatalf("Slurm not available, skipping")
	}

	res, execRes := Run(&j, nil, &jobmgr, &sysCfg, nil)
	if !res.Pass {
		t.Fatalf("execution failed: %s, %s, %s %s", execRes.Err, execRes.Stdout, execRes.Stderr, res.Note)
	}

	fmt.Printf("Note: %s\n", res.Note)
}
