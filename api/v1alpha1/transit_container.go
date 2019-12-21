/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"fmt"

	core "k8s.io/api/core/v1"
)

func transitContainer(ppl Pipeline) core.Container {
	return core.Container{
		Name:  "transit",
		Image: "mrsixw/concourse-rsync-resource",
		Env: []core.EnvVar{
			{Name: "transit.params.rsync_opts", Value: `["--delete", "--recursive"]`},
			{Name: "transit.source.server", Value: "${MY_IP}"},
			{Name: "transit.source.base_dir", Value: "/tmp"},
			{Name: "transit.source.user", Value: "root"},
			{Name: "transit.source.disable_version_path", Value: "true"},
			{Name: "transit.version", Value: `{"ref":"tmp"}`},
			{Name: "transit.source.private_key", ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					Key:                  rsyncSecretPrivateKey,
					LocalObjectReference: core.LocalObjectReference{Name: fmt.Sprintf(rsyncSecretFormatString, ppl.Name, ppl.Status.BuildNo)},
				},
			}},
		},
	}
}
