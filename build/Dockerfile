# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

ARG CK_BASE_IMAGE=registry.access.redhat.com/ubi8/ubi-minimal:8.10

FROM $CK_BASE_IMAGE AS base

ARG IMAGE_ARCH

USER 1001:0

ADD build/_output/bin/camel-dashboard-${IMAGE_ARCH} /usr/local/bin/camel-dashboard
