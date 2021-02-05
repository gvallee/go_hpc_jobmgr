// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// Copyright (c) 2021, NVIDIA CORPORATION. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package job

import (
	"bytes"

	"github.com/gvallee/go_hpc_jobmgr/internal/pkg/sys"
	"github.com/gvallee/go_hpc_jobmgr/pkg/app"
	"github.com/gvallee/go_hpc_jobmgr/pkg/mpi"
)

// CleanUpFn is a "function pointer" to call to clean up the system after the completion of a job
type CleanUpFn func(...interface{}) error

// GetOutputFn is a "function pointer" to call to gather the output of an application after completion of a job
type GetOutputFn func(*Job, *sys.Config) string

// GetErrorFn is a "function pointer" to call to gather stderr from an application after completion of a job
type GetErrorFn func(*Job, *sys.Config) string

// Job represents a job
type Job struct {
	// Name is the name of the job
	Name string

	// NP is the number of ranks
	NP int

	// NNodes is the number of nodes
	NNodes int

	// CleanUp is the function to call once the job is completed to clean the system
	CleanUp CleanUpFn

	// BatchScript is the path to the script required to start a job (optional)
	BatchScript string

	// App is the path to the application's binary, i.e., the binary to start
	App app.Info

	// OutBuffer is a buffer with the output of the job
	OutBuffer bytes.Buffer

	// ErrBuffer is a buffer with the stderr of the job
	ErrBuffer bytes.Buffer

	// internalGetOutput is the function to call to gather the output of the application based on the use of a given job manager
	internalGetOutput GetOutputFn

	// internalGetError is the function to call to gather stderr of the application based on the use of a given job manager
	internalGetError GetErrorFn

	// Args is a set of arguments to be used for launching the job
	Args []string

	// MPICfg is the MPI configuration to use for the execution of the job
	MPICfg *mpi.Config

	// Partition is the name of the partition to use with the jobmgr (optional)
	Partition string
}

// GetOutput is the function to call to gather the output (stdout) of the application after execution of the job
func (j *Job) GetOutput(sysCfg *sys.Config) string {
	return j.internalGetOutput(j, sysCfg)
}

// GetError is the function to call to gather stderr of the application after execution of the job
func (j *Job) GetError(sysCfg *sys.Config) string {
	return j.internalGetError(j, sysCfg)
}

// SetOutputFn sets the internal function specific to the job manager to get the output of a job
func (j *Job) SetOutputFn(fn GetOutputFn) {
	j.internalGetOutput = fn
}

// SetErrorFn sets the internal function specific to the job manager to get stderr of a job
func (j *Job) SetErrorFn(fn GetErrorFn) {
	j.internalGetError = fn
}
