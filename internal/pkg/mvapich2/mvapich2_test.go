// Copyright (c) 2022, NVIDIA CORPORATION. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package mvapich2

import "testing"

func TestParseMVAPICH2InfoOutputForVersion(t *testing.T) {
	output := `MVAPICH2 Version:       2.3.7
	MVAPICH2 Release date:  Wed March 02 22:00:00 EST 2022
	MVAPICH2 Device:        ch3:mrail
	MVAPICH2 configure:     --prefix=/global/scratch/users/benjaminm/Geoffroy-OpenHPCA-benchmark-work/mv2-2.3.7/install CC=gcc CXX=g++ --disable-fortran --enable-fast=all --enable-g=none --with-device=ch3:mrail CFLAGS=-lpthread LDFLAGS=-lpthread
	MVAPICH2 CC:    gcc -lpthread   -DNDEBUG -DNVALGRIND -O2
	MVAPICH2 CXX:   g++   -DNDEBUG -DNVALGRIND -O2
	MVAPICH2 F77:   gfortran
	MVAPICH2 FC:    gfortran`
	expectedResult := "2.3.7"

	version, err := parseMVAPICH2InfoOutputForVersion(output)
	if err != nil {
		t.Fatalf("parseMVAPICH2InfoOutputForVersion() failed: %s", err)
	}
	if version != expectedResult {
		t.Fatalf("parseMVAPICH2InfoOutputForVersion() returned %s instead of %s", version, expectedResult)
	}
}
