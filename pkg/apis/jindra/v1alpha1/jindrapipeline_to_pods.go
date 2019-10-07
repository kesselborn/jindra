package v1alpha1

import (
	"fmt"

	core "k8s.io/api/core/v1"
)

// JindraPipelineRunConfigs keeps pod configurations for the pipeline run
type JindraPipelineRunConfigs []map[string]core.Pod

func pipelineConfigs(p JindraPipeline, buildNo int) (JindraPipelineRunConfigs, error) {
	config := JindraPipelineRunConfigs{}
	for i, stage := range p.Spec.Stages {
		if stage.Kind == "" {
			stage.Kind = "Pod"
		}
		if stage.APIVersion == "" {
			stage.APIVersion = "v1"
		}
		if stage.Labels == nil {
			stage.Labels = map[string]string{}
		}
		if stage.Labels["jindra.io/uid"] == "" {
			stage.Labels["jindra.io/uid"] = "${MY_UID}"
		}
		// stage.ObjectMeta.Name = fmt.Sprintf("${MY_NAME}.%02d-%s", buildNo, stage.ObjectMeta.Name)

		nodeAffinity := core.NodeAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []core.PreferredSchedulingTerm{
				{
					Preference: core.NodeSelectorTerm{
						MatchExpressions: []core.NodeSelectorRequirement{
							{
								Key:      "kubernetes.io/hostname",
								Operator: core.NodeSelectorOpIn,
								Values:   []string{"${MY_NODE_NAME}"},
							},
						},
					},
					Weight: 1,
				},
			},
		}

		stage.Spec.Affinity = &core.Affinity{NodeAffinity: &nodeAffinity}

		name := fmt.Sprintf("jindra.%s-%02d.%02d-%s.yaml", p.ObjectMeta.Name, buildNo, i+1, stage.ObjectMeta.Name)
		config = append(config, map[string]core.Pod{name: stage})
	}

	return config, nil
}
