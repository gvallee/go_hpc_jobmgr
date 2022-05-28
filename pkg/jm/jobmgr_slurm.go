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
	"github.com/gvallee/go_hpc_jobmgr/internal/pkg/openmpi"
	"github.com/gvallee/go_hpc_jobmgr/internal/pkg/slurm"
	"github.com/gvallee/go_hpc_jobmgr/pkg/job"
	"github.com/gvallee/go_hpc_jobmgr/pkg/mpi"
	"github.com/gvallee/go_hpc_jobmgr/pkg/sys"
	"github.com/gvallee/go_util/pkg/util"
)

const (
	slurmJobIDPrefix = "Submitted batch job "
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
		// Try to get more details with a sacct command
		var sacctCmd advexec.Advcmd
		var err error
		sacctCmd.BinPath, err = exec.LookPath("sacct")
		if err != nil {
			return StatusUnknown, err
		}
		sacctCmd.CmdArgs = []string{"-j", strconv.Itoa(jobID), "--format=state"}
		resSacctCmd := sacctCmd.Run()
		if resSacctCmd.Err == nil {
			output := strings.Split(resSacctCmd.Stdout, "\n")
			if len(output) > 2 {
				if strings.Contains(output[2], "COMPLETED") {
					return StatusDone, nil
				}
			}
		}
		return StatusStop, nil
	}

	return StatusUnknown, nil
}

func slurmGetNumJobs(jobmgr *JM, partitionName string, user string) (int, error) {
	var cmd advexec.Advcmd
	var err error
	cmd.BinPath, err = exec.LookPath("squeue")
	if err != nil {
		return -1, err
	}
	cmd.CmdArgs = []string{"-p", partitionName, "-u", user}
	res := cmd.Run()
	if res.Err != nil {
		return -1, res.Err
	}

	lines := strings.Split(res.Stdout, "\n")
	numJobs := 0
	for _, line := range lines {
		if strings.Contains(line, "JOBID") || line == "" {
			continue
		}
		numJobs++
	}

	return numJobs, nil
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
	jm.numJobsJM = slurmGetNumJobs
	jm.postRunJM = slurmPostJob

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
	return nil
}

func getJobOutFilenamePrefix(j *job.Job) string {
	if j.ExecutionTimestamp == "" {
		return ""
	}
	if j.MPICfg != nil && j.MPICfg.Implem.ID != "" {
		return j.Name + "-" + j.ExecutionTimestamp + "-" + j.MPICfg.Implem.ID + j.MPICfg.Implem.Version
	}
	return j.Name + "-" + j.ExecutionTimestamp
}

func getJobOutputFilePath(j *job.Job, sysCfg *sys.Config) string {
	return getJobOutFilenamePrefix(j) + ".out"
}

func getJobErrorFilePath(j *job.Job, sysCfg *sys.Config) string {
	return getJobOutFilenamePrefix(j) + ".err"
}

func generateBatchScriptContent(j *job.Job, sysCfg *sys.Config) (string, error) {
	// TempFile is supposed to set the path to the batch script
	if j.BatchScript == "" {
		return "", fmt.Errorf("batch script path is undefined")
	}

	scriptText := "#!/bin/bash -l\n#\n"
	if j.Partition != "" {
		scriptText += slurm.ScriptCmdPrefix + " -p " + j.Partition + "\n"
	}

	if j.NNodes > 0 {
		scriptText += slurm.ScriptCmdPrefix + " -N " + strconv.Itoa(j.NNodes) + "\n"
	}

	if j.MaxExecTime == "" {
		scriptText += slurm.ScriptCmdPrefix + " -t 0:30:0\n"
	} else {
		scriptText += slurm.ScriptCmdPrefix + " -t " + j.MaxExecTime + "\n"
	}

	/*
		if j.NP > 0 {
			scriptText += slurm.ScriptCmdPrefix + " --ntasks=" + strconv.Itoa(j.NP) + "\n"
		}
	*/

	j.SetTimestamp()
	scriptText += slurm.ScriptCmdPrefix + " --error=" + getJobErrorFilePath(j, sysCfg) + "\n"
	scriptText += slurm.ScriptCmdPrefix + " --output=" + getJobOutputFilePath(j, sysCfg) + "\n"
	scriptText += "\n"

	if len(j.RequiredModules) > 0 {
		scriptText += "\nmodule purge\nmodule load " + strings.Join(j.RequiredModules, " ") + "\n"
	}

	if j.CustomEnv != nil {
		for envvar, val := range j.CustomEnv {
			scriptText += fmt.Sprintf("export %s=%s\n", envvar, val)
		}
	}

	return scriptText, nil
}

func setupMpiJob(j *job.Job, sysCfg *sys.Config) error {
	scriptText, err := generateBatchScriptContent(j, sysCfg)
	if err != nil {
		return err
	}

	netCfg := new(network.Config)
	netCfg.Device = j.Device

	if j.CustomEnv != nil {
		for envvar, val := range j.CustomEnv {
			scriptText += fmt.Sprintf("export %s=%s\n", envvar, val)
		}
	}

	// Add the mpirun command
	if j.MPICfg != nil && len(j.RequiredModules) == 0 {
		scriptText += "\nMPI_DIR=" + j.MPICfg.Implem.InstallDir + "\n"
		scriptText += "export PATH=$MPI_DIR/bin:$PATH\n"
		scriptText += "export LD_LIBRARY_PATH=$MPI_DIR/lib:$LD_LIBRARY_PATH\n\n"
	}
	mpirunArgs, errMpiArgs := mpi.GetMpirunArgs(&j.MPICfg.Implem, &j.App, sysCfg, netCfg, j.MPICfg.UserMpirunArgs)
	if errMpiArgs != nil {
		return fmt.Errorf("unable to get mpirun arguments: %s", err)
	}

	scriptText += "\nwhich mpirun\n"

	scriptText += "\nmpirun "
	if j.NP > 0 {
		scriptText += fmt.Sprintf("-np %d ", j.NP)
	}
	// todo: this should really be in the openmpi package
	if j.MPICfg.Implem.ID == openmpi.ID {
		ppr := j.NP / j.NNodes
		scriptText += fmt.Sprintf("--map-by ppr:%d:node -rank-by core -bind-to core", ppr)
	}
	scriptText += " " + strings.Join(mpirunArgs, " ") + " " + j.App.BinPath + strings.Join(j.App.BinArgs," ") + "\n"

	err = ioutil.WriteFile(j.BatchScript, []byte(scriptText), 0644)
	if err != nil {
		return fmt.Errorf("unable to write to file %s: %s", j.BatchScript, err)
	}

	log.Printf("batch script ready: %s\n", j.BatchScript)

	return nil
}

func setupNonMpiJob(j *job.Job, sysCfg *sys.Config) error {
	if j.BatchScript == "" {
		return fmt.Errorf("undefined job script path")
	}
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

	// If we know nothing about the app and there is no batch script to use, we do
	// not know how to launch the application
	if j.App.BinPath == "" && j.BatchScript == "" {
		return fmt.Errorf("application binary and batch script are undefined")
	}

	// Create the batch script if the user did not specify a batch script to use.
	// If the user specified the batch script to use, we use it as it is
	if j.BatchScript == "" {
		err := TempFile(j, sysCfg)
		if err != nil {
			return fmt.Errorf("unable to create temporary file: %s", err)
		}

		// Some sanity checks, required to set everything up for MPI
		if j.MPICfg == nil || j.MPICfg.Implem.ID == "" {
			return setupNonMpiJob(j, sysCfg)
		}
		return setupMpiJob(j, sysCfg)
	}

	fmt.Printf("-> Using the user defined batch script %s\n", j.BatchScript)
	return nil
}

func slurmPostJob(cmdRes *advexec.Result, j *job.Job, sysCfg *sys.Config) advexec.Result {
	var expRes advexec.Result
	expRes.Err = cmdRes.Err

	stdoutFile := getJobOutputFilePath(j, sysCfg)
	if j.RunDir != "" {
		stdoutFile = filepath.Join(j.RunDir, stdoutFile)
	}
	outputFileContent, err := ioutil.ReadFile(stdoutFile)
	if err != nil {
		expRes.Err = fmt.Errorf("unable to read %s: %s", stdoutFile, err)
		return expRes
	}
	expRes.Stdout = string(outputFileContent)

	stderrFile := getJobOutputFilePath(j, sysCfg)
	if j.RunDir != "" {
		stderrFile = filepath.Join(j.RunDir, stderrFile)
	}
	errFileContent, err := ioutil.ReadFile(stderrFile)
	if err != nil {
		expRes.Err = fmt.Errorf("unable to read %s: %s", stderrFile, err)
		return expRes
	}
	expRes.Stderr = string(errFileContent)
	return expRes
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
	cmd.ExecDir = j.RunDir
	// We want the default to be blocking sbatch but users can request non-blocking
	if !j.NonBlocking {
		jobmgr.CmdArgs = append(jobmgr.CmdArgs, "-W")
	}

	if len(jobmgr.CmdArgs) > 0 {
		cmd.CmdArgs = append(cmd.CmdArgs, jobmgr.CmdArgs...)
	}
	//cmd.CmdArgs = append(cmd.CmdArgs, j.BatchScript)
	cmd.CmdArgs = []string{j.BatchScript}

	j.SetOutputFn(slurmGetOutput)
	j.SetErrorFn(slurmGetError)

	if !util.PathExists(sysCfg.ScratchDir) {
		resExec.Err = fmt.Errorf("scratch directory does not exist")
		return resExec
	}

	cmdRes := cmd.Run()
	if strings.HasPrefix(cmdRes.Stdout, slurmJobIDPrefix) {
		jobIDStr := strings.TrimPrefix(cmdRes.Stdout, slurmJobIDPrefix)
		jobIDStr = strings.TrimRight(jobIDStr, "\n")
		j.ID, err = strconv.Atoi(jobIDStr)
		if err != nil {
			resExec.Err = fmt.Errorf("unable to get job ID: %s", err)
			return resExec
		}
	}

	if !j.NonBlocking {
		return slurmPostJob(&cmdRes, j, sysCfg)
	}

	return cmdRes
}
