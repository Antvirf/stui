.PHONY: lint
lint:
	find -name "*.go" | xargs -I{} go fmt {}


.PHONY: save
save:
	git add .
	git commit --amend --no-edit
	git push --force


.PHONY: build-cluster config-cluster run-cluster launch-jobs stop-cluster
build-cluster:
	mkdir -p ./build && \
		cd ./build && \
		stat slurm.tar.bz2 > /dev/null || wget https://download.schedmd.com/slurm/slurm-24.11.3.tar.bz2 -O slurm.tar.bz2

	cd ./build && \
		stat slurm-* > /dev/null || tar -xaf slurm*tar.bz2

	cd ./build/slurm-* && \
		./configure \
		--enable-debug \
		--sysconfdir=/etc/slurm \
		--enable-multiple-slurmd \
		--enable-front-end && \
		sudo make install && \
		sudo ldconfig -n /usr/lib64

config-cluster:
	sudo cp ./testing/mock-cluster-slurmconf.conf /etc/slurm/slurm.conf

run-cluster:
	sudo slurmctld && echo "Launched Slurmctld"
	sudo slurmd -N localhost && echo "Launched Slurmd"
	@sleep 2
	@sdiag && echo "\n\nCluster up and running!"

launch-jobs:
	bash testing/test-job-launcher.sh

stop-cluster:
	sudo kill $$(ps aux | grep '[s]lurm' | awk '{print $$2}')

.PHONY: mock
mock: config-cluster run-cluster launch-jobs



