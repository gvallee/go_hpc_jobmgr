// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// Copyright (c) 2020-2021, NVIDIA CORPORATION. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package launcher

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/gvallee/go_exec/pkg/advexec"
	"github.com/gvallee/go_exec/pkg/results"
	"github.com/gvallee/go_hpc_jobmgr/internal/pkg/sys"
	"github.com/gvallee/go_hpc_jobmgr/pkg/jm"
	"github.com/gvallee/go_hpc_jobmgr/pkg/job"
	"github.com/gvallee/go_hpc_jobmgr/pkg/mpi"
)

// Info gathers all the details to start a job
type Info struct {
	// Cmd represents the command to launch a job
	Cmd advexec.Advcmd
}

// PrepareLaunchCmd interacts with a job manager backend to figure out how to launch a job
func prepareLaunchCmd(j *job.Job, jobmgr *jm.JM, sysCfg *sys.Config) (advexec.Advcmd, error) {
	var cmd advexec.Advcmd

	// Sanity checks
	if j == nil || jobmgr == nil || sysCfg == nil {
		return cmd, fmt.Errorf("invalid parameter(s)")
	}

	cmd.Ctx, cmd.CancelFn = context.WithTimeout(context.Background(), advexec.CmdTimeout*time.Minute)
	cmd.Cmd = exec.CommandContext(cmd.Ctx, jobmgr.BinPath, jobmgr.CmdArgs...)
	cmd.Cmd.Stdout = &j.OutBuffer
	cmd.Cmd.Stderr = &j.ErrBuffer
	//cmd.Cmd.Env = jm.Env

	res := jobmgr.Submit(j, sysCfg)
	if res.Err != nil {
		return cmd, fmt.Errorf("failed to submit the job: %s, stdout:%s, stderr:%s", res.Err, res.Stdout, res.Stderr)
	}

	return cmd, nil
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

// Run executes a container with a specific version of MPI on the host
func Run(j *job.Job, hostMPI *mpi.Config, jobmgr *jm.JM, sysCfg *sys.Config, args []string) (results.Result, advexec.Result) {
	var execRes advexec.Result
	var expRes results.Result
	expRes.Pass = true
	errorMsg := ""

	if hostMPI != nil {
		j.HostCfg = &hostMPI.Implem
	}

	if len(args) == 0 {
		j.NNodes = 2
		j.NP = 2
	} else {
		j.Args = append(j.Args, args...)
	}

	// We submit the job
	var submitCmd advexec.Advcmd
	submitCmd, execRes.Err = prepareLaunchCmd(j, jobmgr, sysCfg)
	if execRes.Err != nil {
		execRes.Err = fmt.Errorf("failed to prepare the launch command: %s", execRes.Err)
		expRes.Pass = false
		return expRes, execRes
	}

	var stdout, stderr bytes.Buffer
	submitCmd.Cmd.Stdout = &stdout
	submitCmd.Cmd.Stderr = &stderr
	defer submitCmd.CancelFn()

	execRes = jobmgr.Submit(j, sysCfg)

	/*
		err := submitCmd.Cmd.Run()
		// Get the command out/err
		execRes.Stderr = stderr.String()
		execRes.Stdout = stdout.String()
		execRes.Err = err
		// And add the job out/err (for when we actually use a real job manager such as Slurm)
		execRes.Stdout += newjob.GetOutput(&newjob, sysCfg)
		execRes.Stderr += newjob.GetError(&newjob, sysCfg)
	*/

	if execRes.Err != nil {
		// The command simply failed and the Go runtime caught it
		expRes.Pass = false
		errorMsg = fmt.Sprintf("[ERROR] Command failed - stdout: %s - stderr: %s - err: %s\n", execRes.Stdout, execRes.Stderr, execRes.Err)
		log.Printf(errorMsg)
	}
	if submitCmd.Ctx.Err() == context.DeadlineExceeded {
		// The command timed out
		expRes.Pass = false
		errorMsg = fmt.Sprintf("[ERROR] Command timed out - stdout: %s - stderr: %s\n", execRes.Stdout, execRes.Stderr)
		log.Printf(errorMsg)
	}

	if !expRes.Pass {
		expRes.Note = errorMsg
	}

	/*
		if expRes.Pass {
			// Regex to catch errors where mpirun returns 0 but is known to have failed because displaying the help message
			var re = regexp.MustCompile(`^(\n?)Usage:`)
			if re.Match(stdout.Bytes()) {
				// mpirun actually failed, exited with 0 as return code but displayed the usage message (so nothing really ran)
				expRes.Pass = false
				log.Printf("[ERROR] mpirun failed and returned help messafe - stdout: %s - stderr: %s\n", stdout.String(), stderr.String())
			}
			if !expectedOutput(execRes.Stdout, execRes.Stderr, appInfo, &newjob) {
				// The output is NOT the expected output
				expRes.Pass = false
				log.Printf("[ERROR] Run succeeded but output is not matching expectation - stdout: %s - stderr: %s\n", stdout.String(), stderr.String())
			}
		}

		// For any error, we save details to give a chance to the user to analyze what happened
		if !expRes.Pass {
			if hostMPI != nil && containerMPI != nil {
				err = SaveErrorDetails(&hostMPI.Implem, &containerMPI.Implem, sysCfg, &execRes)
				if err != nil {
					// We only log the error because the most important error is the error
					// that happened while executing the command
					log.Printf("impossible to cleanly handle error: %s", err)
				}
			} else {
				log.Println("Not an MPI job, not saving error details")
			}
		}
	*/

	return expRes, execRes
}
