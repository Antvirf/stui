.PHONY: setup lint build update-readme
setup:
	pip install pre-commit
	pre-commit install
	go mod download

lint:
	find -name "*.go" | xargs -I{} go fmt {}
	go mod tidy

build: lint
	go build

update-readme: build
	@echo "Updating README.md with current help output..."
	@(echo "    stui help"; ./stui -help 2>&1) | \
		sed -e '1d' -e 's/^/    /' | \
		awk 'BEGIN {print "    ```"} {print} END {print "    ```"}' > .help.tmp
	@sed -i '/<!-- REPLACE_START -->/,/<!-- REPLACE_END -->/{//!d}' README.md
	@sed -i '/<!-- REPLACE_START -->/r .help.tmp' README.md
	@rm -f .help.tmp
	@echo "README.md updated successfully"


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



