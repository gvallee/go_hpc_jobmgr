// Copyright (c) 2022, NVIDIA CORPORATION. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gvallee/go_hpc_jobmgr/pkg/mpi"
)

func main() {
	dirFlag := flag.String("dir", "", "Path to the install directory where the MPI is installed")
	help := flag.Bool("h", false, "Help message")

	flag.Parse()

	cmdName := filepath.Base(os.Args[0])
	if *help {
		fmt.Printf("%s is a command line tool to MPI implementation is installed in a given directory", cmdName)
		fmt.Println("\nUsage:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	i, err := mpi.DetectFromDir(*dirFlag)
	if err != nil {
		fmt.Printf("unable to detect the MPI implementation installed in %s: %s\n", *dirFlag, err)
		os.Exit(1)
	}
	fmt.Printf("Detected MPI:\n%s %s\n", i.ID, i.Version)
}
