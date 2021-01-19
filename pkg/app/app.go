// Copyright (c) 2021, NVIDIA CORPORATION. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package app

// Info gathers information about a given application
type Info struct {
	// Name is the name of the application
	Name string

	// BinName is the name of the binary to start executing the application
	BinName string

	// BinPath is the path to the binary to start executing the application
	BinPath string

	// BinArgs is the list of argument that the application's binary needs
	BinArgs []string
}
