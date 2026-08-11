package kubernetes

import (
	"context"
	"maps"

	pkglabels "github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/pkg/to"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
	"github.com/zeitlos/lucity/services/conductor/internal/resources"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) ensureQuota(ctx context.Context, id platform.EnvironmentID) error {
	namespace := id.Namespace()
	quota := buildQuota(namespace)

	existing, err := c.k8s.CoreV1().ResourceQuotas(namespace).Get(ctx, resourceQuotaName, metav1.GetOptions{})

	if apierrors.IsNotFound(err) {
		if _, err := c.k8s.CoreV1().ResourceQuotas(namespace).Create(ctx, quota, metav1.CreateOptions{}); err != nil {
			return err
		}

		return nil
	}

	if err != nil {
		return err
	}

	if resourceListEqual(existing.Spec.Hard, quota.Spec.Hard) && maps.Equal(existing.Labels, quota.Labels) {
		return nil
	}

	existing.Spec.Hard = quota.Spec.Hard
	existing.Labels = quota.Labels

	if _, err := c.k8s.CoreV1().ResourceQuotas(namespace).Update(ctx, existing, metav1.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}

func (c *Client) ensureLimitRange(ctx context.Context, id platform.EnvironmentID, tier platform.ResourceTier) error {
	namespace := id.Namespace()
	limitRange := buildLimitRange(namespace, tier)

	existing, err := c.k8s.CoreV1().LimitRanges(namespace).Get(ctx, limitRangeName, metav1.GetOptions{})

	if apierrors.IsNotFound(err) {
		if _, err := c.k8s.CoreV1().LimitRanges(namespace).Create(ctx, limitRange, metav1.CreateOptions{}); err != nil {
			return err
		}

		return nil
	}

	if err != nil {
		return err
	}

	if limitRangeSpecEqual(existing.Spec, limitRange.Spec) && maps.Equal(existing.Labels, limitRange.Labels) {
		return nil
	}

	existing.Spec = limitRange.Spec
	existing.Labels = limitRange.Labels

	if _, err := c.k8s.CoreV1().LimitRanges(namespace).Update(ctx, existing, metav1.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}

func resourceListEqual(a, b corev1.ResourceList) bool {
	if len(a) != len(b) {
		return false
	}

	for name, left := range a {
		right, ok := b[name]

		if !ok || left.Cmp(right) != 0 {
			return false
		}
	}

	return true
}

func limitRangeSpecEqual(a, b corev1.LimitRangeSpec) bool {
	if len(a.Limits) != len(b.Limits) {
		return false
	}

	for i, left := range a.Limits {
		right := b.Limits[i]

		if left.Type != right.Type {
			return false
		}

		if !resourceListEqual(left.Default, right.Default) {
			return false
		}

		if !resourceListEqual(left.DefaultRequest, right.DefaultRequest) {
			return false
		}

		if !resourceListEqual(left.Max, right.Max) || !resourceListEqual(left.Min, right.Min) {
			return false
		}
	}

	return true
}

func buildQuota(namespace string) *corev1.ResourceQuota {
	return &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceQuotaName,
			Namespace: namespace,
			Labels: map[string]string{
				pkglabels.ManagedBy: pkglabels.ManagedByLucity,
			},
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: corev1.ResourceList{
				corev1.ResourceRequestsCPU:     resources.DefaultCPUQuota,
				corev1.ResourceRequestsMemory:  resources.DefaultMemoryQuota,
				corev1.ResourceRequestsStorage: resources.DefaultStorageQuota,
			},
		},
	}
}

func buildLimitRange(namespace string, tier platform.ResourceTier) *corev1.LimitRange {
	return &corev1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Name:      limitRangeName,
			Namespace: namespace,
			Labels: map[string]string{
				pkglabels.ManagedBy: pkglabels.ManagedByLucity,
			},
		},
		Spec: corev1.LimitRangeSpec{
			Limits: []corev1.LimitRangeItem{{
				Type: corev1.LimitTypeContainer,
				Default: corev1.ResourceList{
					corev1.ResourceCPU:    resources.DefaultCPULimit,
					corev1.ResourceMemory: resources.DefaultMemoryLimit,
				},
				DefaultRequest: corev1.ResourceList{
					corev1.ResourceCPU:    to.Val(resources.Request(tier, resources.DefaultCPULimit)),
					corev1.ResourceMemory: to.Val(resources.Request(tier, resources.DefaultMemoryLimit)),
				},
			}},
		},
	}
}
