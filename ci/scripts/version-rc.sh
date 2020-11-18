#!/bin/bash
#
# Copyright (c) 2017-2020 VMware, Inc. or its affiliates
# SPDX-License-Identifier: Apache-2.0

set -ex

# This makes each build of a given git SHA unique.  For instance, on a branch a while after the 0.4.0 tag,
# git describe gives "0.4.0-32-g763a08e5" and rpmVersion is then "0.4.0+dev.32.g973669ba".
IFS='- ' read -r -a parts <<< "$(git -C ./gpupgrade_src describe)"
rpmVersion="${parts[0]}+dev"
if [ -n "${parts[1]}" ]; then
  rpmVersion="${rpmVersion}.${parts[1]}.${parts[2]}"
fi

# rename the files with a unique per git SHA version.
# So greenplum-upgrade-0.4.0-1.el7.x86_64.rpm becomes
#    greenplum-upgrade-0.4.0+dev.32.g763a08e5.el7.x86_64.rpm

RPM=$(basename "$(ls rpm_gpupgrade_oss/greenplum-upgrade*.rpm)")
VERSIONED=${RPM/"${parts[0]}"/"${rpmVersion}"}
cp rpm_gpupgrade_oss/greenplum-upgrade*.rpm build_artifacts_rc_oss/"${VERSIONED}"

RPM=$(basename "$(ls rpm_gpupgrade_enterprise/greenplum-upgrade*.rpm)")
VERSIONED=${RPM/"${parts[0]}"/"${rpmVersion}"}
cp rpm_gpupgrade_enterprise/greenplum-upgrade*.rpm build_artifacts_rc_enterprise/"${VERSIONED}"
