// Copyright (c) 2021, NVIDIA CORPORATION. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package openmpi

import "testing"

func TestParseOMPIInfoOutputForVersion(t *testing.T) {
	output := "Open MPI v3.0.4\n\nhttp://www.open-mpi.org/community/help/\n"
	expectedResult := "3.0.4"

	version, err := parseOmpiInfoOutputForVersion(output)
	if err != nil {
		t.Fatalf("parseOmpiInfoOutputForVersion() failed: %s", err)
	}
	if version != expectedResult {
		t.Fatalf("parseOmpiInfoOutputForVersion() returned %s instead of %s", version, expectedResult)
	}
}