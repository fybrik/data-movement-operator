ROOT_DIR := ..
include $(ROOT_DIR)/Makefile.env
#include Makefile.env

.PHONY: all
all: manager

include $(ROOT_DIR)/hack/make-rules/tools.mk
DOCKER_NAME ?= dmo-manager
GO_OUTPUT_FILE = manager
KUBE_NAMESPACE ?= fybrik-system
CONTROLLER_NAMESPACE ?= ${KUBE_NAMESPACE}
include $(ROOT_DIR)/hack/make-rules/docker.mk
include $(ROOT_DIR)/hack/make-rules/verify.mk

# Run tests
test: export BLUEPRINT_NAMESPACE?=fybrik-blueprints
test: export CONTROLLER_NAMESPACE?=fybrik-system
test: $(TOOLBIN)/etcd $(TOOLBIN)/kube-apiserver
	go test ./... -v  -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

source-build: generate fmt vet
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go --enable-all-controllers --metrics-bind-addr=0

# Generate manifests e.g. CRD, RBAC etc.
.PHONY: generate
generate: $(TOOLBIN)/controller-gen
	$(TOOLBIN)/controller-gen --version
	$(TOOLBIN)/controller-gen object:headerFile=$(ROOT_DIR)/hack/boilerplate.go.txt,year=$(shell date +%Y) paths="./..."

# Generate code
.PHONY: manifests
manifests: $(TOOLBIN)/controller-gen $(TOOLBIN)/yq
	$(TOOLBIN)/controller-gen --version
	$(TOOLBIN)/controller-gen crd:trivialVersions=true output:crd:artifacts:config=$(ROOT_DIR)/charts/data-movement-operator/files/ paths=./apis/...
	$(TOOLBIN)/controller-gen webhook paths=./apis/... output:stdout | \
		$(TOOLBIN)/yq eval '.metadata.annotations."cert-manager.io/inject-ca-from" |= "{{ .Release.Namespace }}/dmo-serving-cert"' - | \
		$(TOOLBIN)/yq eval '.metadata.annotations."certmanager.k8s.io/inject-ca-from" |= "{{ .Release.Namespace }}/dmo-serving-cert"' - | \
		$(TOOLBIN)/yq eval '(.metadata.name | select(. == "mutating-webhook-configuration")) = "{{ .Release.Namespace }}-dmo-mutating-webhook"' - | \
		$(TOOLBIN)/yq eval '(.metadata.name | select(. == "validating-webhook-configuration")) = "{{ .Release.Namespace }}-dmo-validating-webhook"' - | \
		$(TOOLBIN)/yq eval '(.webhooks.[].clientConfig.service.name) = "dmo-webhook-service"' - | \
		$(TOOLBIN)/yq eval '(.webhooks.[].clientConfig.service.namespace) = "{{ .Release.Namespace }}"' - > $(ROOT_DIR)/charts/data-movement-operator/files/webhook-configs.yaml

# Overwrite docker-build from docker.mk
docker-build: generate source-build
	docker build . -t ${IMG} -f Dockerfile
	rm manager

.PHONY: wait_for_manager
wait_for_manager: $(TOOLBIN)/kubectl
	$(TOOLBIN)/kubectl wait --for=condition=available -n ${CONTROLLER_NAMESPACE} deployment/dmo-manager --timeout=120s

.PHONY: run-integration-tests
run-integration-tests: export DOCKER_HOSTNAME?=localhost:5000
run-integration-tests: export DOCKER_NAMESPACE?=fybrik-system
run-integration-tests: export DOCKER_TAGNAME?=0.0.0
run-integration-tests: export USE_MOCKUP_CONNECTOR?=true
run-integration-tests: wait_for_manager
	NO_SIMULATED_PROGRESS=true USE_EXISTING_CONTROLLER=true USE_EXISTING_CLUSTER=true go test ./... -v -run TestMotionAPIs -count 1

.PHONY: main.deps
main.deps: generate fmt vet manifests

DEBUG := ./debug.out

$(DEBUG): main.go */*/*.go
	go build -o $(DEBUG) main.go
