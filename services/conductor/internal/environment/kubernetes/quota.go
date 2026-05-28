package kubernetes

import (
	"context"

	pkglabels "github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
	"github.com/zeitlos/lucity/services/conductor/internal/resources"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	defaultQuotaCPU     = resource.MustParse("4")
	defaultQuotaMemory  = resource.MustParse("8Gi")
	defaultQuotaStorage = resource.MustParse("40Gi")

	defaultContainerCPULimit    = resource.MustParse("500m")
	defaultContainerMemoryLimit = resource.MustParse("512Mi")
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

	existing.Spec = limitRange.Spec
	existing.Labels = limitRange.Labels

	if _, err := c.k8s.CoreV1().LimitRanges(namespace).Update(ctx, existing, metav1.UpdateOptions{}); err != nil {
		return err
	}

	return nil
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
				corev1.ResourceRequestsCPU:     defaultQuotaCPU,
				corev1.ResourceRequestsMemory:  defaultQuotaMemory,
				corev1.ResourceRequestsStorage: defaultQuotaStorage,
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
					corev1.ResourceCPU:    defaultContainerCPULimit,
					corev1.ResourceMemory: defaultContainerMemoryLimit,
				},
				DefaultRequest: corev1.ResourceList{
					corev1.ResourceCPU:    resources.Request(tier, defaultContainerCPULimit),
					corev1.ResourceMemory: resources.Request(tier, defaultContainerMemoryLimit),
				},
			}},
		},
	}
}
