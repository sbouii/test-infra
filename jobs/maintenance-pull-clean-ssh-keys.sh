#!/bin/bash
# Copyright 2016 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

projects=(
  k8s-jkns-pr-bldr-e2e-gce-fdrtn
  k8s-jkns-pr-cnry-e2e-gce-fdrtn
  k8s-jkns-pr-gce
  k8s-jkns-pr-gce-bazel
  k8s-jkns-pr-gce-etcd3
  k8s-jkns-pr-gci-gce
  k8s-jkns-pr-gci-gke
  k8s-jkns-pr-gci-kubemark
  k8s-jkns-pr-gke
  k8s-jkns-pr-kubeadm
  k8s-jkns-pr-kubemark
  )

for proj in "${projects[@]}"; do
  echo ./jenkins/clean_metadata.py --project="${proj}"
done
