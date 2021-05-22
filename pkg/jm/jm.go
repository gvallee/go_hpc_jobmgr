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
	"os"
	"path/filepath"

	"github.com/gvallee/go_exec/pkg/advexec"
	"github.com/gvallee/go_hpc_jobmgr/pkg/job"
	"github.com/gvallee/go_hpc_jobmgr/pkg/sys"
	"github.com/gvallee/go_util/pkg/util"
)

const (
	// NativeID is the value set to JM.ID when mpirun shall be used to submit a job
	NativeID = "native"

	// SlurmID is the value set to JM.ID when Slurm shall be used to submit a job
	SlurmID = "slurm"

	// PrunID is the value set to JM.ID when prun shall be used to submit a job
	PrunID = "prun"
)

// Environment represents the job's environment to use
type Environment struct {
	// InstallDir is where software packages needed for the job are installed
	InstallDir string

	//mpiBin string
}

type JobStatus struct {
	Code int
	Str  string
}

const (
	JOB_STATUS_UNKNOWN = iota
	JOB_STATUS_PENDING
	JOB_STATUS_QUEUED
	JOB_STATUS_RUNNING
	JOB_STATUS_STOP
	JOB_STATUS_DONE
)

var StatusUnknown = JobStatus{
	Code: JOB_STATUS_UNKNOWN,
	Str:  "UNKNOWN",
}
var StatusPending = JobStatus{
	Code: JOB_STATUS_PENDING,
	Str:  "PENDING",
}
var StatusQueued = JobStatus{
	Code: JOB_STATUS_QUEUED,
	Str:  "QUEUED",
}
var StatusRunning = JobStatus{
	Code: JOB_STATUS_RUNNING,
	Str:  "RUNNING",
}
var StatusStop = JobStatus{
	Code: JOB_STATUS_STOP,
	Str:  "STOPPED",
}
var StatusDone = JobStatus{
	Code: JOB_STATUS_DONE,
	Str:  "DONE",
}

// Loader checks whether a giv job manager is applicable or not
type Loader interface {
	Load() bool
}

// LoadFn loads a specific job manager once detected
type LoadFn func(jobmgr *JM, sysCfg *sys.Config) error

// SubmitFn is a "function pointer" that lets us job a new job
type SubmitFn func(j *job.Job, jobmgr *JM, sysCfg *sys.Config) advexec.Result

type JobStatusFn func(jobmgr *JM, jobIDs []int) ([]JobStatus, error)

type NumJobsFn func(jobmgr *JM, partition string, user string) (int, error)

// JM is the structure representing a specific JM
type JM struct {
	// ID identifies which job manager has been detected on the system
	ID string

	loadJM LoadFn

	submitJM SubmitFn

	jobStatusJM JobStatusFn

	numJobsJM NumJobsFn

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

// TempFile creates a temporary file that is used to store a batch script
func TempFile(j *job.Job, sysCfg *sys.Config) error {
	filePrefix := "sbatch-" + j.Name
	path := ""
	if sysCfg.Persistent == "" {
		f, err := ioutil.TempFile(sysCfg.ScratchDir, filePrefix+"-")
		if err != nil {
			return fmt.Errorf("failed to create temporary file: %s", err)
		}
		path = f.Name()
		f.Close()
		j.BatchScript = path
	} else {
		fileName := filePrefix + ".sh"
		path = filepath.Join(j.MPICfg.Implem.InstallDir, fileName)
		j.BatchScript = path
		if util.PathExists(path) {
			return fmt.Errorf("file %s already exists", path)
		}
	}

	j.CleanUp = func(...interface{}) error {
		err := os.RemoveAll(path)
		if err != nil {
			return fmt.Errorf("unable to delete %s: %s", path, err)
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

func (jobmgr *JM) JobStatus(jobIDs []int) ([]JobStatus, error) {
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
