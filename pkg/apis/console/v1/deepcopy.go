/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import "k8s.io/apimachinery/pkg/runtime"

func (in *ConsolePlugin) DeepCopyObject() runtime.Object {
	if in == nil {
		return nil
	}

	out := new(ConsolePlugin)
	in.DeepCopyInto(&out.ObjectMeta)
	out.TypeMeta = in.TypeMeta
	out.Spec = in.Spec

	if in.Spec.Backend.Service != nil {
		svc := *in.Spec.Backend.Service
		out.Spec.Backend.Service = &svc
	}

	return out
}

func (in *ConsolePluginList) DeepCopyObject() runtime.Object {
	if in == nil {
		return nil
	}

	out := new(ConsolePluginList)
	out.TypeMeta = in.TypeMeta
	in.DeepCopyInto(&out.ListMeta)

	if in.Items != nil {
		out.Items = make([]ConsolePlugin, len(in.Items))
		for i := range in.Items {
			cp, ok := in.Items[i].DeepCopyObject().(*ConsolePlugin)
			if !ok {
				continue
			}

			out.Items[i] = *cp
		}
	}

	return out
}
