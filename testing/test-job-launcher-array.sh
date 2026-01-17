#!/bin/bash
  for QOS in high low immediate sparecapacity; do
      CPUS=$(( (RANDOM % 16) + 10 ))
      MEMORY=$(( (RANDOM % 64) + 1 ))

      cat testing/mock-cluster-slurmconf.conf  | grep PartitionName | cut -d= -f2 | cut -d' ' -f1 | \
          xargs -I{} sbatch --partition={} --uid 1337 --job-name="array job {} $NUMBER" --comment="array this is a multi word comment of job-{}-$NUMBER" --mem="$MEMORY"G  --cpus-per-task=$CPUS --qos=$QOS --out=/dev/null --array 0-5 testing/sleep.sh "$NUMBER"
  done
