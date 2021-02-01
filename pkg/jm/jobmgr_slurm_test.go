// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// Copyright (c) 2021, NVIDIA CORPORATION. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package jm

import (
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gvallee/go_hpc_jobmgr/internal/pkg/sys"
	"github.com/gvallee/go_hpc_jobmgr/pkg/job"
)

var partition = flag.String("partition", "", "Name of Slurm partition to use to run the test")
var scratchDir = flag.String("scratch", "", "Scratch directory to use to execute the test")

func isDateCmdOutput(output string) bool {
	tokens := strings.Split(output, " ")
	if tokens[0] == "Mon" || tokens[0] == "Tue" || tokens[0] == "Wed" || tokens[0] == "Thu" || tokens[0] == "Fri" || tokens[0] == "Sat" || tokens[0] == "Sun" {
		return true
	}
	return false
}

// TestSlurmSubmit tests detecting, setting and submitting a basic Slurm job,
// assuming the system as Slurm installed (otherwise the test is skipped).
// To run the test on a specific partition, set the environment variable
// 'GO_HPC_JOBMGR_TEST_SLURM_PARTITION' to the target partition
func TestSlurmSubmit(t *testing.T) {
	failed := false

	loaded, jobmgr := SlurmDetect()
	if !loaded {
		t.Skip("slurm cannot be used on this platform")
	}

	var j job.Job
	var err error
	j.App.Name = "date"
	j.App.BinPath, err = exec.LookPath("date")
	if err != nil {
		t.Fatalf("unable to find path to 'date' binnary")
	}

	var sysCfg sys.Config
	installDir, err := ioutil.TempDir(*scratchDir, "install")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %s", err)
	}
	defer os.RemoveAll(installDir)
	sysCfg.ScratchDir, err = ioutil.TempDir(*scratchDir, "")
	if err != nil {
		t.Fatalf("unable to create scratch directory: %s", err)
	}
	defer os.RemoveAll(sysCfg.ScratchDir)
	j.BatchScript = filepath.Join(sysCfg.ScratchDir, "test_run_script.sh")
	j.Partition = *partition

	err = slurmLoad(&jobmgr, &sysCfg)
	if err != nil {
		t.Fatalf("unable to load Slurm: %s", err)
	}

	res := slurmSubmit(&j, &jobmgr, &sysCfg)
	if res.Err != nil {
		t.Fatalf("test failed: %s, stdout:%s, stderr:%s", res.Err, res.Stdout, res.Stderr)
	}

	// Display the content of the batch script
	if !failed {
		f, err := os.Open(j.BatchScript)
		if err != nil {
			failed = true
			t.Logf("failed to open batch script %s: %s", j.BatchScript, err)
		} else {
			b, err := ioutil.ReadAll(f)
			if err != nil {
				t.Logf("failed to read the batch script: %s", err)
			}
			t.Logf("Content of the batch script:\n")
			t.Logf(string(b))
		}
		defer f.Close()
	}

	output := j.GetOutput(&sysCfg)
	if output == "" || !isDateCmdOutput(output) {
		t.Fatalf("invalid output: %s", output)
	}

	/*
		err = job.CleanUp()
		if err != nil {
			t.Logf("failed to call the cleanup function: %s", err)
			failed = true
		}
	*/

	if failed {
		t.Fatalf("test failed")
	}
	t.Logf("Slurm batch script: %s\n", j.BatchScript)

}
