#!/bin/bash
for NUMBER in {1..30}; do
    CPUS=$(( (RANDOM % 16) + 1 ))
    MEMORY=$(( (RANDOM % 64) + 1 ))

    cat testing/mock-cluster-slurmconf.conf  | grep PartitionName | cut -d= -f2 | cut -d' ' -f1 | \
        xargs -I{} sbatch --partition={} --uid 1337 --job-name=job-{}-$NUMBER --mem="$MEMORY"G  --cpus-per-task=$CPUS --out=/dev/null testing/sleep.sh "$NUMBER"
done
