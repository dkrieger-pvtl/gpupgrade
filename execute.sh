#!/bin/bash

rm -fr ~/.gpupgrade/
gpupgrade kill-services

source /usr/local/gpdb5/greenplum_path.sh && source ~/workspace/gpdb/worktree-5X/gpAux/gpdemo/gpdemo-env.sh && echo GPDB-5
pushd ~/workspace/gpdb/worktree-5X/gpAux/gpdemo && rm -fr datadirs && cp -r datadirs_5X_clean datadirs && popd
gpstart -a

gpupgrade initialize --source-gphome=/usr/local/gpdb5 --target-gphome=/usr/local/gpdb6 --source-master-port=15432 --disk-free-ratio 0  --verbose
gpupgrade execute --verbose
gpupgrade revert --verbose
