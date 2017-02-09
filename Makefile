PROJECT_NAME=fabric8-forker
PACKAGE_NAME := github.com/fabric8io/fabric8-forker
CUR_DIR=$(shell pwd)
TMP_PATH=$(CUR_DIR)/tmp
INSTALL_PREFIX=$(CUR_DIR)/bin
VENDOR_DIR=vendor
WORKSPACE ?= /tmp

# If running in Jenkins we don't allow for interactively running the container
ifneq ($(BUILD_TAG),)
	DOCKER_RUN_INTERACTIVE_SWITCH :=
else
	DOCKER_RUN_INTERACTIVE_SWITCH := -i
endif

DOCKER_IMAGE_CORE := $(PROJECT_NAME)
DOCKER_IMAGE_DEPLOY := $(PROJECT_NAME)-deploy
DOCKER_BUILD_DIR := $(WORKSPACE)/$(PROJECT_NAME)-build

# The BUILD_TAG environment variable will be set by jenkins
# to reflect jenkins-${JOB_NAME}-${BUILD_NUMBER}
BUILD_TAG ?= $(PROJECT_NAME)-local-build
DOCKER_CONTAINER_NAME := $(BUILD_TAG)

# Where is the GOPATH inside the build container?
GOPATH_IN_CONTAINER=/tmp/go
PACKAGE_PATH=$(GOPATH_IN_CONTAINER)/src/$(PACKAGE_NAME)

.PHONY: docker-build-build
docker-build-build:
	mkdir -p $(DOCKER_BUILD_DIR)
	docker build -t $(DOCKER_IMAGE_CORE) -f $(CUR_DIR)/Dockerfile.builder $(CUR_DIR)
	docker run \
		--detach=true \
		-t \
		$(DOCKER_RUN_INTERACTIVE_SWITCH) \
		--name="$(DOCKER_CONTAINER_NAME)" \
		-v $(CUR_DIR):$(PACKAGE_PATH):Z \
		-u $(shell id -u $(USER)):$(shell id -g $(USER)) \
		-e GOPATH=$(GOPATH_IN_CONTAINER) \
		-w $(PACKAGE_PATH) \
		$(DOCKER_IMAGE_CORE)
		@echo "Docker container \"$(DOCKER_CONTAINER_NAME)\" created. Continue with \"make docker-deps\"."

.PHONY: docker-build-run
docker-build-run:
	docker build -t $(DOCKER_IMAGE_DEPLOY) -f $(CUR_DIR)/Dockerfile.deploy $(CUR_DIR)
	docker tag $(DOCKER_IMAGE_DEPLOY) fabric8io/fabric8-forker:latest

.PHONY: docker-run-deploy
docker-run-deploy:
	docker tag fabric8io/fabric8-forker registry.ci.centos.org:5000/fabric8io/fabric8-forker:latest
	docker push registry.ci.centos.org:5000/fabric8io/fabric8-forker:latest 

# This is a wildcard target to let you call any make target from the normal makefile
# but it will run inside the docker container. This target will only get executed if
# there's no specialized form available. For example if you call "make docker-start"
# not this target gets executed but the "docker-start" target. 
docker-%:
	$(eval makecommand:=$(subst docker-,,$@))
ifeq ($(strip $(shell docker ps -qa --filter "name=$(DOCKER_CONTAINER_NAME)" 2>/dev/null)),)
	$(error No container name "$(DOCKER_CONTAINER_NAME)" exists to run the command "make $(makecommand)")
endif
	docker exec -t $(DOCKER_RUN_INTERACTIVE_SWITCH) "$(DOCKER_CONTAINER_NAME)" bash -ec 'make $(makecommand)'


.PHONY: test
test:
	go test $$(glide novendor)

.PHONY: build
build:
	mkdir -p bin
	go build -o bin/fabric8-forker

.PHONY: install
install:
	glide install