# Copyright (c) 2021 NVIDIA CORPORATION. All rights reserved.

.PHNOY: jobmgr

all: jobmgr

jobmgr:
	cd cmd/jobmgr; go build jobmgr.go

clean:
	@rm -f cmd/jobmgr/jobmgr
