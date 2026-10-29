package kubernetes

import (
	"context"
	"maps"

	pkglabels "github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/pkg/platform"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	networkPolicyName = "lucity-namespace-isolation"
	linkLocalCIDR     = "169.254.0.0/16"
)

func (c *Client) ensureNetworkPolicy(ctx context.Context, id platform.EnvironmentID) error {
	namespace := id.Namespace()
	policy := buildNetworkPolicy(namespace, c.systemNamespace, c.podCIDR, c.serviceCIDR)

	existing, err := c.k8s.NetworkingV1().NetworkPolicies(namespace).Get(ctx, networkPolicyName, metav1.GetOptions{})

	if apierrors.IsNotFound(err) {
		if _, err := c.k8s.NetworkingV1().NetworkPolicies(namespace).Create(ctx, policy, metav1.CreateOptions{}); err != nil {
			return err
		}

		return nil
	}

	if err != nil {
		return err
	}

	if apiequality.Semantic.DeepEqual(existing.Spec, policy.Spec) && maps.Equal(existing.Labels, policy.Labels) {
		return nil
	}

	existing.Spec = policy.Spec
	existing.Labels = policy.Labels

	if _, err := c.k8s.NetworkingV1().NetworkPolicies(namespace).Update(ctx, existing, metav1.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}

func buildNetworkPolicy(namespace, platformNamespace, podCIDR, serviceCIDR string) *networkingv1.NetworkPolicy {
	dnsPort := intstr.FromInt(53)
	udp := corev1.ProtocolUDP
	tcp := corev1.ProtocolTCP

	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      networkPolicyName,
			Namespace: namespace,
			Labels: map[string]string{
				pkglabels.ManagedBy: pkglabels.ManagedByLucity,
			},
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{Key: pkglabels.Agent, Operator: metav1.LabelSelectorOpDoesNotExist},
				},
			},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
				networkingv1.PolicyTypeEgress,
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					From: []networkingv1.NetworkPolicyPeer{
						{PodSelector: &metav1.LabelSelector{}},
					},
				},
				{
					From: []networkingv1.NetworkPolicyPeer{
						{NamespaceSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"kubernetes.io/metadata.name": platformNamespace,
							},
						}},
					},
				},
			},
			Egress: []networkingv1.NetworkPolicyEgressRule{
				{
					To: []networkingv1.NetworkPolicyPeer{
						{NamespaceSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"kubernetes.io/metadata.name": "kube-system",
							},
						}},
					},
					Ports: []networkingv1.NetworkPolicyPort{
						{Protocol: &udp, Port: &dnsPort},
						{Protocol: &tcp, Port: &dnsPort},
					},
				},
				{
					To: []networkingv1.NetworkPolicyPeer{
						{PodSelector: &metav1.LabelSelector{}},
					},
				},
				{
					To: []networkingv1.NetworkPolicyPeer{
						{IPBlock: &networkingv1.IPBlock{
							CIDR:   "0.0.0.0/0",
							Except: []string{serviceCIDR, podCIDR, linkLocalCIDR},
						}},
					},
				},
			},
		},
	}
}
