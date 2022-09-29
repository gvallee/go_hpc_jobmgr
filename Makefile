# Copyright (c) 2021-2022 NVIDIA CORPORATION. All rights reserved.

.PHNOY: jobmgr mpi_detect

all: jobmgr mpi_detect

jobmgr:
	cd cmd/jobmgr; go build jobmgr.go

mpi_detect:
	cd cmd/mpi_detect; go build mpi_detect.go

clean:
	@rm -f cmd/jobmgr/jobmgr cmd/mpi_detect/mpi_detect
