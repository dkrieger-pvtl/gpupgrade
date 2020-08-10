#!/bin/bash

source /usr/local/gpdb6/greenplum_path.sh
export PGPORT=$1
export MASTER_DATA_DIRECTORY=$2
gpstop -a


source /usr/local/gpdb5/greenplum_path.sh && source ~/workspace/gpdb/worktree-5X/gpAux/gpdemo/gpdemo-env.sh && echo GPDB-5
gpstart -a
gpstate

