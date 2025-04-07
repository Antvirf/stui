## DEVELOPMENT HELPERS

.PHONY: setup lint build update-readme install
setup:
	pip install pre-commit
	pre-commit install
	go mod download
	sudo useradd johndoe -u 1337 -g 1337 -m -s /bin/bash

lint:
	find -name "*.go" | xargs -I{} go fmt {}
	go mod tidy

build: lint
	go build

install:
	go install

update-readme: build
	@echo "Updating README.md with current help output..."
	@(echo "    stui help"; ./stui -help 2>&1) | \
		sed -e '1d' -e 's/^/    /' | \
		awk 'BEGIN {print "    ```"} {print} END {print "    ```"}' > .help.tmp
	@sed -i '/<!-- REPLACE_START -->/,/<!-- REPLACE_END -->/{//!d}' README.md
	@sed -i '/<!-- REPLACE_START -->/r .help.tmp' README.md
	@rm -f .help.tmp

	@echo "Updating README.md with current shortcuts output..."
	@(echo "    stui shortcuts"; ./stui -show-keyboard-shortcuts 2>&1) | \
		sed -e '1d' -e 's/^/    /' | \
		awk 'BEGIN {print "    ```"} {print} END {print "    ```"}' > .help.tmp
	@sed -i '/<!-- REPLACE_SHORTCUTS_START -->/,/<!-- REPLACE_SHORTCUTS_END -->/{//!d}' README.md
	@sed -i '/<!-- REPLACE_SHORTCUTS_START -->/r .help.tmp' README.md
	@rm -f .help.tmp

	@echo "README.md updated successfully"


## DEVELOPMENT SLURM CLUSTER

.PHONY: build-cluster config-cluster run-cluster launch-jobs stop-cluster mock
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
	sudo mkdir -p /etc/slurm
	sudo cp ./testing/mock-cluster-slurmconf.conf /etc/slurm/slurm.conf

run-cluster:
	sudo useradd slurm || true
	sudo slurmctld && echo "Launched Slurmctld"
	sudo slurmd -N localhost && echo "Launched Slurmd"
	@sleep 2
	@sdiag && echo "\n\nCluster up and running!"

launch-jobs:
	sudo bash testing/test-job-launcher.sh

stop-cluster:
	sudo kill $$(ps aux | grep '[s]lurm' | awk '{print $$2}')

mock: config-cluster run-cluster launch-jobs


## TESTING UTILITIES

generate-test-data:
	sdiag > ./internal/model/testdata/sdiag.txt
	scontrol show nodes --all --oneliner > ./internal/model/testdata/nodes.txt
	scontrol show jobs --all --oneliner > ./internal/model/testdata/jobs.txt
	scontrol show partitions --all --oneliner > ./internal/model/testdata/partitions.txt

test: launch-jobs
	go test -v ./internal/model


offline-test:
	go test -v ./internal/model -run TestParse*


## RELEASE UTILITIES
.PHONY: release-check release-dryrun release update-version-in-go fail-if-any-files-changed execute-demo convert-demo-to-gif
release-check:
	~/go/bin/goreleaser check

release-dryrun:
	~/go/bin/goreleaser release --snapshot --clean

release:
	~/go/bin/goreleaser --release --clean

update-version-in-go:
	export GIT_TAG=$$(git describe --tags --abbrev=0) && \
	echo $$GIT_TAG && \
	sed -i "s/\(STUI_VERSION[[:space:]]*=[[:space:]]*\)\".*\"/\1\"$$GIT_TAG\"/" ./internal/config/config.go

fail-if-any-files-changed:
	git diff --exit-code
	if [ "$$?" -ne 0 ]; then \
		echo "There are uncommitted changes in the repository. Please commit or stash them before running this target."; \
		exit 1; \
	fi

execute-demo: build install launch-jobs
	uv run testing/demo.py

convert-demo-to-gif:
	rm assets/demo.gif
	ffmpeg -i $$(ls demo* | tail -n1) \
    	-vf "fps=7,scale=800:-1:flags=lanczos,split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse" \
    	-loop 0 assets/demo.gif
