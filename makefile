## DEVELOPMENT HELPERS

.PHONY: run run-with-all-columns setup lint build update-readme install
setup:
	pip install pre-commit
	pre-commit install
	go mod download
	sudo useradd johndoe -u 1337 -g 1337 -m -s /bin/bash

lint:
	find -name "*.go" | xargs -I{} go fmt {}
	go mod tidy
	go vet ./...

build: lint
	goreleaser build --snapshot --clean --id stui-linux --single-target
	cp dist/stui-linux_linux_amd64_v1/stui ./stui

run:
	go run main.go

run-with-all-columns:
	go run main.go -show-all-columns

install:
	go install

update-readme: build
	@echo "Updating README.md with current help output..."
	@(echo "    stui help"; ./stui -help 2>&1 | expand -i -t 2) | \
		sed -e '1d' -e 's/^/    /' | \
		awk 'BEGIN {print "    ```txt"} {print} END {print "    ```"}' > .help.tmp
	@sed -i '/<!-- REPLACE_START -->/,/<!-- REPLACE_END -->/{//!d}' README.md
	@sed -i '/<!-- REPLACE_START -->/r .help.tmp' README.md
	@rm -f .help.tmp

	@echo "Updating README.md with current shortcuts output..."
	@(echo "    stui shortcuts"; ./stui -show-keyboard-shortcuts 2>&1) | \
		sed -e '1d' -e 's/^/    /' | \
		awk 'BEGIN {print "    ```txt"} {print} END {print "    ```"}' > .help.tmp
	@sed -i '/<!-- REPLACE_SHORTCUTS_START -->/,/<!-- REPLACE_SHORTCUTS_END -->/{//!d}' README.md
	@sed -i '/<!-- REPLACE_SHORTCUTS_START -->/r .help.tmp' README.md
	@rm -f .help.tmp

	@echo "Updating README.md with lines of code badge..."
	@LOC=$$(rg 'package' -l | grep ".go" | xargs wc -l | grep total | cut -d' ' -f2) && \
	sed -i "s/lines%20of%20code-[0-9]*/lines%20of%20code-$${LOC}/" README.md

	@echo "Updating README.md with binary size badge..."
	@SIZE=$$(du -h ./dist/stui-linux_linux_amd64_v1/stui | cut -f1 | sed 's/\./%2E/') && \
	sed -i "s/binary%20size-[^-]*/binary%20size-$${SIZE}/" README.md

	@echo "README.md updated successfully"

## DEVELOPMENT SLURM CLUSTER

.PHONY: build-cluster config-cluster run-cluster setup-sacct launch-jobs stop-cluster mock
build-cluster:
	mkdir -p ./build && \
		cd ./build && \
		stat slurm.tar.bz2 > /dev/null || wget https://download.schedmd.com/slurm/slurm-24.11.3.tar.bz2 -O slurm.tar.bz2

	cd ./build && \
		stat slurm-* > /dev/null || tar -xaf slurm*tar.bz2

	cd ./build/slurm-* && \
		make distclean && \
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
	sudo cp ./testing/mock-cluster-slurmdbd.conf /etc/slurm/slurmdbd.conf
	sudo chown slurm /etc/slurm/slurmdbd.conf
	sudo chmod 600 /etc/slurm/slurmdbd.conf

run-cluster:
	cd testing/ && docker compose --file mariadb-compose.yaml up -d
	sudo useradd slurm || true
	sudo slurmdbd && echo "Launched Slurmdbd"
	sudo slurmctld && echo "Launched Slurmctld"
	sudo slurmd -N localhost && echo "Launched Slurmd"
	@sleep 5
	@sdiag && echo "\n\nCluster up and running!"

setup-sacct:
	sudo bash testing/sacct-setup.sh

launch-jobs:
	sudo bash testing/test-job-launcher.sh

stop-cluster:
	sudo kill -9 $$(ps aux | grep '[s]lurm' | awk '{print $$2}')

mock: config-cluster run-cluster setup-sacct launch-jobs


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

release:
	sed -i "s/\(STUI_VERSION[[:space:]]*=[[:space:]]*\)\".*\"/\1\"$$GIT_TAG\"/" ./internal/config/config.go
	git add internal/config/config.go
	git commit -m "release: $$GIT_TAG"
	git tag $$GIT_TAG
	git push
	git push origin $$GIT_TAG

fail-if-any-files-changed:
	git diff --exit-code
	if [ "$$?" -ne 0 ]; then \
		echo "There are uncommitted changes in the repository. Please commit or stash them before running this target."; \
		exit 1; \
	fi

execute-demo: build install launch-jobs
	uv run testing/demo.py

convert-demo-to-gif:
	rm -f assets/demo.gif
	ffmpeg -i $$(ls demo* | tail -n1) \
    	-vf "fps=7,scale=800:-1:flags=lanczos,split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse" \
    	-loop 0 assets/demo.gif
