// Copyright (c) 2022, NVIDIA CORPORATION. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package mvapich2

import "testing"

func TestParseMVAPICH2InfoOutputForVersion(t *testing.T) {
	output := "\n"
	expectedResult := "0.0.0"

	version, err := parseMVAPICH2InfoOutputForVersion(output)
	if err != nil {
		t.Fatalf("parseMVAPICH2InfoOutputForVersion() failed: %s", err)
	}
	if version != expectedResult {
		t.Fatalf("parseMVAPICH2InfoOutputForVersion() returned %s instead of %s", version, expectedResult)
	}
}
