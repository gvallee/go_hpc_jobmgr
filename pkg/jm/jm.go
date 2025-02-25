// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// Copyright (c) 2020-2025, NVIDIA CORPORATION. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package jm

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gvallee/go_exec/pkg/advexec"
	"github.com/gvallee/go_hpc_jobmgr/pkg/job"
	"github.com/gvallee/go_hpc_jobmgr/pkg/sys"
	"github.com/gvallee/go_hpcjob/pkg/hpcjob"
	"github.com/gvallee/go_util/pkg/util"
)

const (
	// NativeID is the value set to JM.ID when mpirun shall be used to submit a job
	NativeID = "native"

	// SlurmID is the value set to JM.ID when Slurm shall be used to submit a job
	SlurmID = "slurm"

	// IntelSlurmID is the value set to JM.ID when Intel-Slurm shall be used to submit a job
	IntelSlurmID = "intel-slurm"

	// PrunID is the value set to JM.ID when prun shall be used to submit a job
	PrunID = "prun"
)

// Environment represents the job's environment to use
type Environment struct {
	// InstallDir is where software packages needed for the job are installed
	InstallDir string

	//mpiBin string
}

// Loader checks whether a giv job manager is applicable or not
type Loader interface {
	Load() bool
}

// LoadFn loads a specific job manager once detected
type LoadFn func(jobmgr *JM, sysCfg *sys.Config) error

// SubmitFn is a "function pointer" that lets us submit a new job
type SubmitFn func(j *job.Job, jobmgr *JM, sysCfg *sys.Config) advexec.Result

// JobStatusFn is a "function pointer" that lets us query the status of a job
type JobStatusFn func(jobmgr *JM, jobIDs []int) ([]hpcjob.Status, error)

// NumJobsFn is a "function pointer" that lets us know how many jobs the job manager is currently handling
type NumJobsFn func(jobmgr *JM, partition string, user string) (int, error)

// PostJobFn is a "function pointer" that lets us update results once the job completes. By default jobs are blocking, in which case this does not need to be used.
type PostJobFn func(cmdRes *advexec.Result, j *job.Job, sysCfg *sys.Config) advexec.Result

// JM is the structure representing a specific JM
type JM struct {
	// ID identifies which job manager has been detected on the system
	ID string

	loadJM LoadFn

	submitJM SubmitFn

	jobStatusJM JobStatusFn

	numJobsJM NumJobsFn

	postRunJM PostJobFn

	BinPath string

	CmdArgs []string
}

// Detect figures out which job manager must be used on the system and return a
// structure that gather all the data necessary to interact with it
func Detect() JM {
	// Default job manager
	loaded, comp := NativeDetect()
	if !loaded {
		log.Fatalln("unable to find a default job manager")
	}

	// Now we check if we can find better
	loaded, slurmComp := SlurmDetect()
	if loaded {
		return slurmComp
	}

	loaded, prunComp := PrunDetect()
	if loaded {
		return prunComp
	}

	return comp
}

func getBatchScriptPath(j *job.Job, sysCfg *sys.Config, batchScriptFilenamePrefix string) (string, error) {
	if j.RunDir != "" {
		return filepath.Join(j.RunDir, batchScriptFilenamePrefix+".sh"), nil
	}

	if sysCfg.Persistent == "" {
		f, err := os.CreateTemp(sysCfg.ScratchDir, batchScriptFilenamePrefix+"-")
		if err != nil {
			return "", fmt.Errorf("failed to create temporary file: %s", err)
		}
		path := f.Name()
		f.Close()
		return path, nil
	}

	if j.MPICfg == nil {
		fileName := batchScriptFilenamePrefix + ".sh"
		path := filepath.Join(j.MPICfg.Implem.InstallDir, fileName)
		if util.PathExists(path) {
			return "", fmt.Errorf("file %s already exists", path)
		}
		return path, nil
	}
	return "", fmt.Errorf("unable to determine the path to use for the batch script")
}

// TempFile creates a temporary file that is used to store a batch script
func TempFile(j *job.Job, sysCfg *sys.Config) error {
	j.SetTimestamp()
	filePrefix := "sbatch-" + j.ExecutionTimestamp + "-"
	filePrefix += j.Name
	var err error
	j.BatchScript, err = getBatchScriptPath(j, sysCfg, filePrefix)
	if err != nil {
		return err
	}

	j.CleanUp = func(...interface{}) error {
		err := os.RemoveAll(j.BatchScript)
		if err != nil {
			return fmt.Errorf("unable to delete %s: %s", j.BatchScript, err)
		}
		return nil
	}

	return nil
}

// Load sets data specific to the job managers that was previously detected
func (jobmgr *JM) Load(sysCfg *sys.Config) error {
	return jobmgr.loadJM(jobmgr, sysCfg)
}

// Submit executes a job with a job manager that was previously detected and loaded
func (jobmgr *JM) Submit(j *job.Job, sysCfg *sys.Config) advexec.Result {
	return jobmgr.submitJM(j, jobmgr, sysCfg)
}

func (jobmgr *JM) JobStatus(jobIDs []int) ([]hpcjob.Status, error) {
	if jobmgr.jobStatusJM == nil {
		return nil, fmt.Errorf("not implemented")
	}
	return jobmgr.jobStatusJM(jobmgr, jobIDs)
}

func (jobmgr *JM) NumJobs(partition string, user string) (int, error) {
	if jobmgr.numJobsJM == nil {
		return -1, fmt.Errorf("not implemented")
	}
	return jobmgr.numJobsJM(jobmgr, partition, user)
}

func (jobmgr *JM) PostRun(cmdRes *advexec.Result, j *job.Job, sysCfg *sys.Config) advexec.Result {
	var res advexec.Result
	if jobmgr.postRunJM == nil {
		res.Err = fmt.Errorf("not implemented")
		return res
	}
	return jobmgr.postRunJM(cmdRes, j, sysCfg)
}
