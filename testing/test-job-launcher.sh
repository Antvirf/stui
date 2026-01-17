#!/bin/bash
for NUMBER in {1..15}; do
    for QOS in high low immediate sparecapacity; do
        CPUS=1
        MEMORY=$(( (RANDOM % 2) + 1 ))

        cat testing/mock-cluster-slurmconf.conf  | grep PartitionName | cut -d= -f2 | cut -d' ' -f1 | \
          xargs -I{} sbatch --partition={} --uid 1337 --job-name="job {} $NUMBER" --comment="this is a multi word comment with pipes (|)  of job-{}-$NUMBER" --mem="$MEMORY"G  --cpus-per-task=$CPUS --qos=$QOS --out=/dev/null testing/sleep.sh "$NUMBER"
    done
done
