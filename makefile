.PHONY: slurm
slurm:
	sudo slurmd & \
	sudo slurmctld &

.PHONY: lint
lint:
	find -name "*.go" | xargs -I{} go fmt {}


.PHONY: save
save:
	git add .
	git commit --amend --no-edit
	git push --force