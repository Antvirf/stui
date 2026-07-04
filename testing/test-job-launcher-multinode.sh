#!/bin/bash
# Submit multi-node jobs to verify node list expansion in stui.
# When these jobs are running/completed, the NodeList column will show
# compressed notation (e.g. "linux[0100-0104]") which stui now expands
# for searchability. You can verify by searching for an individual node
# name like "linux0102" in the stui interface.

for NODES in 2 4 8; do
    for QOS in high low; do
        cat testing/mock-cluster-slurmconf.conf | grep PartitionName | cut -d= -f2 | cut -d' ' -f1 | \
          xargs -I{} sbatch \
            --partition={} \
            --uid 1337 \
            --job-name="multinode-{}-${NODES}n" \
            --comment="multi-node job for testing node list expansion" \
            --nodes=$NODES \
            --ntasks-per-node=1 \
            --qos=$QOS \
            --out=/dev/null \
            testing/sleep.sh "60"
    done
done
