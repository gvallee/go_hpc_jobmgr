// Copyright (c) 2021, NVIDIA CORPORATION. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package sys

// Config represents the system configuration
type Config struct {
	// ScratchDir is the path to a scratch directory on the system (most HPC systems have one)
	ScratchDir string

	// Persistent is the path to the directory where to installed software packages in the context of a persistent execution
	Persistent string

	// CurPath is the path to the current directory
	CurPath string
}
