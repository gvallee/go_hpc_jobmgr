// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// Copyright (c) 2020-2021, NVIDIA CORPORATION. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package jm

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gvallee/go_exec/pkg/advexec"
	"github.com/gvallee/go_hpc_jobmgr/internal/pkg/network"
	"github.com/gvallee/go_hpc_jobmgr/internal/pkg/slurm"
	"github.com/gvallee/go_hpc_jobmgr/internal/pkg/sys"
	"github.com/gvallee/go_hpc_jobmgr/pkg/job"
	"github.com/gvallee/go_hpc_jobmgr/pkg/mpi"
	"github.com/gvallee/go_util/pkg/util"
)

func removeFromSlice(a []string, idx int) []string {
	return append(a[:idx], a[idx+1:]...)
}

func getSlurmJobStatus(jobID int) (JobStatus, error) {
	var cmd advexec.Advcmd
	var err error
	cmd.BinPath, err = exec.LookPath("squeue")
	if err != nil {
		return StatusUnknown, err
	}
	cmd.CmdArgs = []string{"-j", strconv.Itoa(jobID), "--format=%t"}
	res := cmd.Run()
	if res.Err != nil {
		// if it fails it might mean the job is done
		res.Stderr = strings.TrimRight(res.Stderr, "\n")
		if strings.HasSuffix(res.Stderr, "Invalid job id specified") {
			return StatusDone, nil
		}
		return StatusUnknown, res.Err
	}

	lines := strings.Split(res.Stdout, "\n")
	// We do not care about the first lines, just the Slurm header and status that are not
	// relevant to us
	for i := 0; i < len(lines); i++ {
		if lines[i] == "" {
			lines = removeFromSlice(lines, i)
		}
	}
	rawStatus := strings.TrimRight(lines[len(lines)-1], "\n")
	switch rawStatus {
	case "R":
		return StatusRunning, nil
	case "PD":
		return StatusQueued, nil
	case "ST":
		return StatusStop, nil
	}

	fmt.Println("This does not make sense")
	return StatusUnknown, nil
}

func slurmJobStatus(jobmgr *JM, jobIDs []int) ([]JobStatus, error) {
	var s []JobStatus
	if jobmgr == nil {
		return nil, fmt.Errorf("undefined job manager")
	}

	for _, jobID := range jobIDs {
		jobStatus, err := getSlurmJobStatus(jobID)
		if err != nil {
			return nil, err
		}
		s = append(s, jobStatus)
	}

	return s, nil
}

// SlurmDetect is the function used by our job management framework to figure out if Slurm can be used and
// if so return a JM structure with all the "function pointers" to interact with Slurm through our generic
// API.
func SlurmDetect() (bool, JM) {
	var jm JM
	var err error

	jm.BinPath, err = exec.LookPath("sbatch")
	if err != nil {
		log.Println("* Slurm not detected")
		return false, jm
	}

	jm.ID = SlurmID
	jm.submitJM = slurmSubmit
	jm.loadJM = slurmLoad
	jm.jobStatusJM = slurmJobStatus

	return true, jm
}

// slurmGetOutput reads the content of the Slurm output file that is associated to a job
func slurmGetOutput(j *job.Job, sysCfg *sys.Config) string {
	outputFile := getJobOutputFilePath(j, sysCfg)
	output, err := ioutil.ReadFile(outputFile)
	if err != nil {
		return ""
	}

	return string(output)
}

// slurmGetError reads the content of the Slurm error file that is associated to a job
func slurmGetError(j *job.Job, sysCfg *sys.Config) string {
	errorFile := getJobErrorFilePath(j, sysCfg)
	errorTxt, err := ioutil.ReadFile(errorFile)
	if err != nil {
		return ""
	}

	return string(errorTxt)
}

// slurmLoad is the function called when trying to load a JM module
func slurmLoad(jobmgr *JM, sysCfg *sys.Config) error {
	// jobmgr.BinPath has been set during Detect()
	jobmgr.CmdArgs = append(jobmgr.CmdArgs, "-W") // We always wait until the submitted job terminates
	return nil
}

func getJobOutFilenamePrefix(j *job.Job) string {
	if j.MPICfg != nil {
		return j.Name + "-" + j.MPICfg.Implem.ID + "-" + j.MPICfg.Implem.Version
	}
	return j.Name
}

func getJobOutputFilePath(j *job.Job, sysCfg *sys.Config) string {
	errorFilename := getJobOutFilenamePrefix(j) + ".out"
	path := filepath.Join(sysCfg.ScratchDir, errorFilename)
	return path
}

func getJobErrorFilePath(j *job.Job, sysCfg *sys.Config) string {
	outputFilename := getJobOutFilenamePrefix(j) + ".err"
	path := filepath.Join(sysCfg.ScratchDir, outputFilename)
	return path
}

func generateBatchScriptContent(j *job.Job, sysCfg *sys.Config) (string, error) {
	// TempFile is supposed to set the path to the batch script
	if j.BatchScript == "" {
		return "", fmt.Errorf("Batch script path is undefined")
	}

	scriptText := "#!/bin/bash\n#\n"
	if j.Partition != "" {
		scriptText += slurm.ScriptCmdPrefix + " --partition=" + j.Partition + "\n"
	}

	if j.NNodes > 0 {
		scriptText += slurm.ScriptCmdPrefix + " --nodes=" + strconv.Itoa(j.NNodes) + "\n"
	}

	if j.NP > 0 {
		scriptText += slurm.ScriptCmdPrefix + " --ntasks=" + strconv.Itoa(j.NP) + "\n"
	}

	scriptText += slurm.ScriptCmdPrefix + " --error=" + getJobErrorFilePath(j, sysCfg) + "\n"
	scriptText += slurm.ScriptCmdPrefix + " --output=" + getJobOutputFilePath(j, sysCfg) + "\n"

	return scriptText, nil
}

func setupMpiJob(j *job.Job, sysCfg *sys.Config) error {
	scriptText, err := generateBatchScriptContent(j, sysCfg)
	if err != nil {
		return err
	}

	netCfg := new(network.Config)
	netCfg.Device = j.Device

	// Add the mpirun command
	mpirunPath := filepath.Join(j.MPICfg.Implem.InstallDir, "bin", "mpirun")
	mpirunArgs, errMpiArgs := mpi.GetMpirunArgs(&j.MPICfg.Implem, &j.App, sysCfg, netCfg, j.MPICfg.UserMpirunArgs)
	if errMpiArgs != nil {
		return fmt.Errorf("unable to get mpirun arguments: %s", err)
	}

	scriptText += "\n" + mpirunPath + " "
	if j.NP > 0 {
		scriptText += fmt.Sprintf("-np %d ", j.NP)
	}
	scriptText += " " + strings.Join(mpirunArgs, " ") + " " + j.App.BinPath + "\n"

	err = ioutil.WriteFile(j.BatchScript, []byte(scriptText), 0644)
	if err != nil {
		return fmt.Errorf("unable to write to file %s: %s", j.BatchScript, err)
	}

	fmt.Printf("batch script ready: %s\n", j.BatchScript)

	return nil
}

func setupNonMpiJob(j *job.Job, sysCfg *sys.Config) error {
	if j.BatchScript == "" {
		return fmt.Errorf("undefined job script path")
	}
	fmt.Printf("Creating %s\n", j.BatchScript)
	scriptText, err := generateBatchScriptContent(j, sysCfg)
	if err != nil {
		return err
	}
	scriptText += "\n" + j.App.BinPath + "\n"

	err = ioutil.WriteFile(j.BatchScript, []byte(scriptText), 0644)
	if err != nil {
		return fmt.Errorf("unable to write to file %s: %s", j.BatchScript, err)
	}

	log.Printf("-> Job script successfully created: %s", j.BatchScript)

	return nil
}

func generateJobScript(j *job.Job, sysCfg *sys.Config) error {
	// Sanity checks
	if j == nil {
		return fmt.Errorf("undefined job")
	}

	if sysCfg.ScratchDir == "" {
		return fmt.Errorf("undefined scratch directory")
	}

	if j.App.BinPath == "" {
		return fmt.Errorf("application binary is undefined")
	}

	// Create the batch script
	if j.BatchScript == "" {
		err := TempFile(j, sysCfg)
		if err != nil {
			return fmt.Errorf("unable to create temporary file: %s", err)
		}
	}

	// Some sanity checks, required to set everything up for MPI
	if j.MPICfg == nil {
		return setupNonMpiJob(j, sysCfg)
	}

	return setupMpiJob(j, sysCfg)
}

// slurmSubmit prepares the batch script necessary to start a given job.
//
// Note that a script does not need any specific environment to be submitted
func slurmSubmit(j *job.Job, jobmgr *JM, sysCfg *sys.Config) advexec.Result {
	var cmd advexec.Advcmd
	var resExec advexec.Result

	// Sanity checks
	if j == nil || !util.FileExists(jobmgr.BinPath) {
		resExec.Err = fmt.Errorf("job is undefined")
		return resExec
	}

	err := generateJobScript(j, sysCfg)
	if err != nil {
		resExec.Err = fmt.Errorf("unable to generate Slurm script: %s", err)
		return resExec
	}
	if j.BatchScript == "" {
		resExec.Err = fmt.Errorf("undefined batch script path")
		return resExec
	}
	cmd.BinPath = jobmgr.BinPath
	cmd.CmdArgs = append(cmd.CmdArgs, jobmgr.CmdArgs...)
	cmd.CmdArgs = append(cmd.CmdArgs, j.BatchScript)

	j.SetOutputFn(slurmGetOutput)
	j.SetErrorFn(slurmGetError)
	return cmd.Run()
}
