.PHONY: slurm
slurm:
	sudo slurmd & \
	sudo slurmctld &

save:
	git add .
	git commit --amend --no-edit
	git push --force