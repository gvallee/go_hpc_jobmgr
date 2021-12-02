// Copyright (c) 2021, NVIDIA CORPORATION. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package mpich

import "testing"

func TestParseMPICHInfoOutputForVersion(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedOutput string
	}{
		{
			name: "v3.4.2",
			input: `HYDRA build details:
    Version:                                 3.4.2
    Release Date:                            Wed May 26 15:51:40 CDT 2021
    CC:                              gcc    
    Configure options:                       '--disable-option-checking' '--prefix=/home/gvallee/install/mpich-3.4.2' 'FCFLAGS=-fallow-argument-mismatch -O2' 'FFLAGS=-fallow-argument-mismatch -O2' '--with-ucx=/home/gvallee/install/ucx-1.9.0' '--cache-file=/dev/null' '--srcdir=.' 'CC=gcc' 'CFLAGS= -O2' 'LDFLAGS= -L/home/gvallee/install/ucx-1.9.0/lib' 'LIBS=' 'CPPFLAGS= -I/home/gvallee/install/ucx-1.9.0/include -DNETMOD_INLINE=__netmod_inline_ucx__ -I/home/gvallee/src/mpich-3.4.2/src/mpl/include -I/home/gvallee/src/mpich-3.4.2/src/mpl/include -I/home/gvallee/src/mpich-3.4.2/modules/yaksa/src/frontend/include -I/home/gvallee/src/mpich-3.4.2/modules/yaksa/src/frontend/include -I/home/gvallee/src/mpich-3.4.2/modules/json-c -I/home/gvallee/src/mpich-3.4.2/modules/json-c -D_REENTRANT -I/home/gvallee/src/mpich-3.4.2/src/mpi/romio/include' 'MPLLIBNAME=mpl'
    Process Manager:                         pmi
    Launchers available:                     ssh rsh fork slurm ll lsf sge manual persist
    Topology libraries available:            hwloc
    Resource management kernels available:   user slurm ll lsf sge pbs cobalt
    Demux engines available:                 poll select`,
			expectedOutput: "3.4.2",
		},
		{
			name: "4.0b1",
			input: `HYDRA build details:
    Version:                                 4.0b1
    Release Date:                            Mon Nov 15 10:22:52 CST 2021
    CC:                              gcc      
    Configure options:                       '--disable-option-checking' '--prefix=/home/gvallee/install/mpich-4.0b1' 'FCFLAGS=-fallow-argument-mismatch -O2' 'FFLAGS=-fallow-argument-mismatch -O2' '--cache-file=/dev/null' '--srcdir=.' 'CC=gcc' 'CFLAGS= -O2' 'LDFLAGS=' 'LIBS=' 'CPPFLAGS=-DNETMOD_INLINE=__netmod_inline_ofi__ -D__HIP_PLATFORM_AMD__ -I/home/gvallee/src/mpich-4.0b1/src/mpl/include -I/home/gvallee/src/mpich-4.0b1/modules/json-c -I/home/gvallee/src/mpich-4.0b1/modules/hwloc/include -D_REENTRANT -I/home/gvallee/src/mpich-4.0b1/src/mpi/romio/include -I/home/gvallee/src/mpich-4.0b1/modules/yaksa/src/frontend/include -I/home/gvallee/src/mpich-4.0b1/modules/libfabric/include'
    Process Manager:                         pmi
    Launchers available:                     ssh rsh fork slurm ll lsf sge manual persist
    Topology libraries available:            hwloc
    Resource management kernels available:   user slurm ll lsf sge pbs cobalt
    Demux engines available:                 poll select`,
			expectedOutput: "4.0b1",
		},
	}

	for _, tt := range tests {
		version, err := parseMPICHInfoOutputForVersion(tt.input)
		if err != nil {
			t.Fatalf("parseOmpiInfoOutputForVersion() failed: %s", err)
		}
		if version != tt.expectedOutput {
			t.Fatalf("parseOmpiInfoOutputForVersion() returned %s instead of %s", version, tt.expectedOutput)
		}
	}
}
