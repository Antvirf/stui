#!/bin/bash
for NUMBER in {1..30}; do
    cat testing/mock-cluster-slurmconf.conf  | grep PartitionName | cut -d= -f2 | cut -d' ' -f1 | \
        xargs -I{} sbatch --partition={} --job-name=job-{}-$NUMBER --out=/dev/null testing/sleep.sh "$NUMBER"
done
