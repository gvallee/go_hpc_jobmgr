// Copyright (c) 2021, NVIDIA CORPORATION. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gvallee/go_hpc_jobmgr/pkg/jm"
)

func main() {
	statusFlag := flag.String("job-status", "", "Display the status of various jobs; comma-separated list of job IDs")
	runningJobsFlag := flag.String("running-jobs", "", "Display how many jobs are already running on the target (e.g., a Slurm partition)")
	help := flag.Bool("h", false, "Help message")

	flag.Parse()

	cmdName := filepath.Base(os.Args[0])
	if *help {
		fmt.Printf("%s is a command line tool to validate HPC applications and libraries", cmdName)
		fmt.Println("\nUsage:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	jobmgr := jm.Detect()
	if *statusFlag != "" {
		jobIDsStr := strings.Split(*statusFlag, ",")
		if len(jobIDsStr) == 0 {
			fmt.Printf("ERROR: please provide a valid list of job IDs\n")
			os.Exit(1)
		}
		var jobIDs []int
		for _, w := range jobIDsStr {
			jobID, err := strconv.Atoi(w)
			if err != nil {
				fmt.Printf("ERROR: invalid job ID: %s\n", w)
				os.Exit(1)
			}
			jobIDs = append(jobIDs, jobID)
		}

		statuses, err := jobmgr.JobStatus(jobIDs)
		if err != nil {
			fmt.Printf("ERROR: unable to retrieve job(s) status: %s\n", err)
			os.Exit(1)
		}
		for idx := range jobIDs {
			fmt.Printf("%d: %s\n", jobIDs[idx], statuses[idx].Str)
		}
	}

	if *runningJobsFlag != "" {
		u, err := user.Current()
		if err != nil {
			fmt.Printf("ERROR: unable to retrieve the user ID: %s\n", err)
			os.Exit(1)
		}
		num, err := jobmgr.NumJobs(*runningJobsFlag, u.Username)
		if err != nil {
			fmt.Printf("ERROR: unable to retrieve the number of running jobs: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Number of running jobs: %d\n", num)
	}
}
