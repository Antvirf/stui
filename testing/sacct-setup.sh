#!/bin/bash


sacm() {
    sacctmgr $@ --immediate
}

sacm create cluster stui-test-cluster 
sacm create account science 
sacm create user name=slurm cluster=stui-test-cluster account=physics 

# Create and configure QOS
sacm add qos immediate
sacm add qos high
sacm add qos low
sacm add qos sparecapacity


sacm modify qos immediate set priority=1000
sacm modify qos high set priority=100
sacm modify qos low set priority=50
sacm add qos sparecapacity=5


# sacct Account for each department
cat testing/mock-cluster-slurmconf.conf  | grep PartitionName | cut -d= -f2 | cut -d' ' -f1 | \
        xargs -I{} sacm create account name={} parent=science 

# account for John Doe in each department
cat testing/mock-cluster-slurmconf.conf  | grep PartitionName | cut -d= -f2 | cut -d' ' -f1 | \
        xargs -I{} sacm create user name=johndoe cluster=stui-test-cluster account={} 

# Give John Doe access to every QOS
sacm modify user johndoe set qos+=immediate
sacm modify user johndoe set qos+=high
sacm modify user johndoe set qos+=low
sacm modify user johndoe set qos+=sparecapacity