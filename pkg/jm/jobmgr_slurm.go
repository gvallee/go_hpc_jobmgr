// Copyright (c) 2019, Sylabs Inc. All rights reserved.
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
	"github.com/gvallee/go_hpc_jobmgr/internal/pkg/job"
	"github.com/gvallee/go_hpc_jobmgr/internal/pkg/slurm"
	"github.com/gvallee/go_hpc_jobmgr/internal/pkg/sys"
	"github.com/gvallee/go_hpc_jobmgr/pkg/mpi"
)

// LoadSlurm is the function used by our job management framework to figure out if Slurm can be used and
// if so return a JM structure with all the "function pointers" to interact with Slurm through our generic
// API.
func SlurmDetect() (bool, JM) {
	var jm JM

	_, err := exec.LookPath("sbatch")
	if err != nil {
		log.Println("* Slurm not detected")
		return false, jm
	}

	jm.ID = SlurmID
	jm.Submit = SlurmSubmit
	jm.Load = SlurmLoad

	return true, jm
}

// SlurmGetOutput reads the content of the Slurm output file that is associated to a job
func SlurmGetOutput(j *job.Job, sysCfg *sys.Config) string {
	outputFile := getJobOutputFilePath(j, sysCfg)
	output, err := ioutil.ReadFile(outputFile)
	if err != nil {
		return ""
	}

	return string(output)
}

// SlurmGetError reads the content of the Slurm error file that is associated to a job
func SlurmGetError(j *job.Job, sysCfg *sys.Config) string {
	errorFile := getJobErrorFilePath(j, sysCfg)
	errorTxt, err := ioutil.ReadFile(errorFile)
	if err != nil {
		return ""
	}

	return string(errorTxt)
}

// SlurmLoad is the function called when trying to load a JM module
func SlurmLoad(jm *JM, sysCfg *sys.Config) error {
	return nil
}

func getJobOutFilenamePrefix(j *job.Job) string {
	return "host-" + j.HostCfg.ID + "-" + j.HostCfg.Version
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

func generateJobScript(j *job.Job, sysCfg *sys.Config) error {
	// Sanity checks
	if j == nil {
		return fmt.Errorf("undefined job")
	}

	// Some sanity checks
	if j.HostCfg == nil {
		return fmt.Errorf("undefined host configuration")
	}

	if sysCfg.ScratchDir == "" {
		return fmt.Errorf("undefined scratch directory")
	}

	if j.App.BinPath == "" {
		return fmt.Errorf("application binary is undefined")
	}

	// Create the batch script
	err := TempFile(j, sysCfg)
	if err != nil {
		return fmt.Errorf("unable to create temporary file: %s", err)
	}

	// TempFile is supposed to set the path to the batch script
	if j.BatchScript == "" {
		return fmt.Errorf("Batch script path is undefined")
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

	// Add the mpirun command
	mpirunPath := filepath.Join(j.MPICfg.Implem.InstallDir, "bin", "mpirun")
	mpirunArgs, err := mpi.GetMpirunArgs(j.HostCfg, &j.App, sysCfg)
	if err != nil {
		return fmt.Errorf("unable to get mpirun arguments: %s", err)
	}
	scriptText += "\n" + mpirunPath + " " + strings.Join(mpirunArgs, " ") + "\n"

	err = ioutil.WriteFile(j.BatchScript, []byte(scriptText), 0644)
	if err != nil {
		return fmt.Errorf("unable to write to file %s: %s", j.BatchScript, err)
	}

	return nil
}

// SlurmSubmit prepares the batch script necessary to start a given job.
//
// Note that a script does not need any specific environment to be submitted
func SlurmSubmit(j *job.Job, sysCfg *sys.Config) (advexec.Advcmd, error) {
	var cmd advexec.Advcmd
	cmd.BinPath = "sbatch"
	cmd.CmdArgs = append(cmd.CmdArgs, "-W") // We always wait until the submitted job terminates

	// Sanity checks
	if j == nil {
		return cmd, fmt.Errorf("job is undefined")
	}

	err := generateJobScript(j, sysCfg)
	if err != nil {
		return cmd, fmt.Errorf("unable to generate Slurm script: %s", err)
	}
	cmd.CmdArgs = append(cmd.CmdArgs, j.BatchScript)

	j.GetOutput = SlurmGetOutput
	j.GetError = SlurmGetError

	return cmd, nil
}
