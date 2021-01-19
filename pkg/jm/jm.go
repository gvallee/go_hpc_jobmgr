// Copyright (c) 2019, Sylabs Inc. All rights reserved.
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
	"github.com/gvallee/go_hpc_jobmgr/internal/pkg/job"
	"github.com/gvallee/go_hpc_jobmgr/internal/pkg/sys"
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

type Environment struct {
	InstallDir string

	mpiBin string
}

// Loader checks whether a giv job manager is applicable or not
type Loader interface {
	Load() bool
}

// LoadFn loads a specific job manager once detected
type LoadFn func(*JM, *sys.Config) error

// SubmitFn is a "function pointer" that lets us job a new job
type SubmitFn func(*job.Job, *sys.Config) (advexec.Advcmd, error)

// JM is the structure representing a specific JM
type JM struct {
	// ID identifies which job manager has been detected on the system
	ID string

	Load LoadFn

	// Submit is the function to submit a job through the current job manager
	Submit SubmitFn
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

// Load is the function to use to load the JM component
func Load(jm *JM) error {
	return nil
}

// TempFile creates a temporary file that is used to store a batch script
func TempFile(j *job.Job, sysCfg *sys.Config) error {
	filePrefix := "sbash-" + j.Name
	path := ""
	if sysCfg.Persistent == "" {
		f, err := ioutil.TempFile("", filePrefix+"-")
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
