package plugin

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/slok/sloth/pkg/common/model"
	k8sutils "github.com/slok/sloth/pkg/common/utils/k8s"
	plugink8stransformv1 "github.com/slok/sloth/pkg/prometheus/plugin/k8stransform/v1"
)

const (
	PluginVersion = "prometheus/k8stransform/v1"
	PluginID      = "github.com/foax/sloth-plugins/k8s-transform-plugins/prom-operator-prometheus-rule/v1"
)

func NewPlugin() (plugink8stransformv1.Plugin, error) {
	return plugin{}, nil
}

type plugin struct{}

func (p plugin) TransformK8sObjects(ctx context.Context, kmeta model.K8sMeta, sloResult model.PromSLOGroupResult) (*plugink8stransformv1.K8sObjects, error) {
	u := &unstructured.Unstructured{}
	u.SetAPIVersion("monitoring.coreos.com/v1")
	u.SetKind("PrometheusRule")
	u.SetNamespace(kmeta.Namespace)
	u.SetName(kmeta.Name)
	u.SetLabels(kmeta.Labels)
	u.SetAnnotations(kmeta.Annotations)

	groups := []any{}
	for _, slo := range sloResult.SLOResults {

		ruleGroup := model.PromRuleGroup{
			Name:     "sloth_slo" + slo.PrometheusRules.SLIErrorRecRules.Name[len("sloth-slo-sli-recordings"):],
			Interval: slo.PrometheusRules.SLIErrorRecRules.Interval,
			Rules:    append(append(slo.PrometheusRules.SLIErrorRecRules.Rules, slo.PrometheusRules.MetadataRecRules.Rules...), slo.PrometheusRules.AlertRules.Rules...),
		}

		groups = append(groups, k8sutils.PromRuleGroupToUnstructuredPromOperator(ruleGroup))

		for _, extraRG := range slo.PrometheusRules.ExtraRules {
			// Skip empty extra rule groups.
			if len(extraRG.Rules) == 0 {
				continue
			}
			groups = append(groups,
				k8sutils.PromRuleGroupToUnstructuredPromOperator(extraRG),
			)
		}
	}

	u.Object["spec"] = map[string]any{
		"groups": groups,
	}

	return &plugink8stransformv1.K8sObjects{
		Items: []*unstructured.Unstructured{u},
	}, nil
}
