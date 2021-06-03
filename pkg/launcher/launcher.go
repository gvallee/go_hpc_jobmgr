// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// Copyright (c) 2020-2021, NVIDIA CORPORATION. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package launcher

import (
	"fmt"
	"log"
	"os"

	"github.com/gvallee/go_exec/pkg/advexec"
	"github.com/gvallee/go_exec/pkg/results"
	"github.com/gvallee/go_hpc_jobmgr/pkg/jm"
	"github.com/gvallee/go_hpc_jobmgr/pkg/job"
	"github.com/gvallee/go_hpc_jobmgr/pkg/mpi"
	"github.com/gvallee/go_hpc_jobmgr/pkg/sys"
)

// Info gathers all the details to start a job
type Info struct {
	// Cmd represents the command to launch a job
	Cmd advexec.Advcmd
}

// Load gathers all the details to start running experiments or create containers for apps
//
// todo: should be in a different package (but where?)
func Load() (sys.Config, jm.JM, error) {
	var cfg sys.Config
	var jobmgr jm.JM

	/* Figure out the directory of this binary */
	var err error
	cfg.CurPath, err = os.Getwd()
	if err != nil {
		return cfg, jobmgr, fmt.Errorf("cannot detect current directory")
	}

	// Load the job manager component first
	jobmgr = jm.Detect()

	return cfg, jobmgr, nil
}

/*
func checkOutput(output string, expected string) bool {
	return strings.Contains(output, expected)
}

func checkJobOutput(output string, expectedOutput string, jobInfo *job.Job) bool {
	if jobInfo.NP > 0 {
		expected := strings.ReplaceAll(expectedOutput, "#NP", strconv.Itoa(jobInfo.NP))
		for i := 0; i < jobInfo.NP; i++ {
			curExpectedOutput := strings.ReplaceAll(expected, "#RANK", strconv.Itoa(i))
			if checkOutput(output, curExpectedOutput) {
				return true
			}
		}
		return false
	}
	return checkOutput(output, expectedOutput)
}

func expectedOutput(stdout string, stderr string, appInfo *app.Info, jobInfo *job.Job) bool {
	if appInfo.ExpectedRankOutput == "" {
		log.Println("App does not define any expected output, skipping check...")
		return true
	}

	// The output can be on stderr or stdout, we just cannot know in advanced.
	// For instance, some MPI applications sends output to stderr by default
	matched := checkJobOutput(stdout, appInfo.ExpectedRankOutput, jobInfo)
	if !matched {
		matched = checkJobOutput(stderr, appInfo.ExpectedRankOutput, jobInfo)
	}

	return matched
}
*/

// Run executes a job with a specific version of MPI on the host.
// This is a blocking function, it returns when the job has completed
func Run(j *job.Job, hostMPI *mpi.Config, jobmgr *jm.JM, sysCfg *sys.Config, args []string) (results.Result, advexec.Result) {
	var execRes advexec.Result
	var expRes results.Result
	expRes.Pass = true
	errorMsg := ""

	if hostMPI != nil {
		j.MPICfg = new(mpi.Config)
		j.MPICfg.Implem = hostMPI.Implem
		j.MPICfg.UserMpirunArgs = hostMPI.UserMpirunArgs
	}

	if len(args) == 0 {
		// No arguments are specified so we make sure we have some basic default values that make sense
		if j.NNodes == 0 {
			j.NNodes = 2
		}
		if j.NP == 0 {
			j.NP = 2
		}
	} else {
		j.Args = append(j.Args, args...)
	}

	// We submit the job
	execRes = jobmgr.Submit(j, sysCfg)
	if execRes.Err != nil {
		// The command simply failed and the Go runtime caught it
		expRes.Pass = false
		errorMsg = fmt.Sprintf("[ERROR] Command failed - stdout: %s - stderr: %s - err: %s\n", execRes.Stdout, execRes.Stderr, execRes.Err)
		log.Printf("%s", errorMsg)
	}

	if !expRes.Pass {
		expRes.Note = errorMsg
	}

	return expRes, execRes
}
