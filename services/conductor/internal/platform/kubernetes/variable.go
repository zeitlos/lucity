package kubernetes

import (
	"context"
	"sort"

	"github.com/zeitlos/lucity/services/conductor/internal/platform"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
)

const (
	cnpgUserTypeLabel        = "cnpg.io/userType"
	cnpgAppUserType          = "app"
	objectStorageBucketLabel = "lucity.dev/object-storage-bucket"
)

var (
	cnpgAppSecretSelector = labels.SelectorFromSet(labels.Set{cnpgUserTypeLabel: cnpgAppUserType})
	sharedSecretSelector  = labels.SelectorFromSet(labels.Set{sharedVariablesLabel: "true"})
	keyValueStoreSelector = existsSelector(keyValueStoreLabel)
	objectStorageSelector = existsSelector(objectStorageBucketLabel)
)

func (c *Client) AvailableVariables(ctx context.Context, environmentID platform.EnvironmentID) ([]platform.Variable, error) {
	var variables []platform.Variable

	for _, source := range c.variableSources {
		found, err := source(ctx, environmentID)

		if err != nil {
			return nil, err
		}

		variables = append(variables, found...)
	}

	return variables, nil
}

type variableSource func(ctx context.Context, environmentID platform.EnvironmentID) ([]platform.Variable, error)

func newVariableSource(client kubernetes.Interface, selector labels.Selector, offeredKeys []string, source func(platform.EnvironmentID, *core.Secret) platform.VariableSource) variableSource {
	return func(ctx context.Context, environmentID platform.EnvironmentID) ([]platform.Variable, error) {
		secrets, err := client.CoreV1().Secrets(environmentID.Namespace()).List(ctx, meta.ListOptions{
			LabelSelector: selector.String(),
		})

		if err != nil {
			return nil, err
		}

		var variables []platform.Variable

		for i := range secrets.Items {
			secret := &secrets.Items[i]
			src := source(environmentID, secret)

			for _, key := range offeredNames(secret, offeredKeys) {
				variables = append(variables, platform.Variable{
					ID: platform.VariableID{
						Workspace:   environmentID.Workspace,
						Project:     environmentID.Project,
						Environment: environmentID.Name,
						Secret:      secret.Name,
						Name:        key,
					},
					Name:   key,
					Source: src,
				})
			}
		}

		return variables, nil
	}
}

func databaseSource(label string) func(platform.EnvironmentID, *core.Secret) platform.VariableSource {
	return func(environmentID platform.EnvironmentID, secret *core.Secret) platform.VariableSource {
		return platform.VariableSource{Database: &platform.DatabaseID{
			Workspace:   environmentID.Workspace,
			Project:     environmentID.Project,
			Environment: environmentID.Name,
			Name:        secret.Labels[label],
		}}
	}
}

func keyValueStoreSource(label string) func(platform.EnvironmentID, *core.Secret) platform.VariableSource {
	return func(environmentID platform.EnvironmentID, secret *core.Secret) platform.VariableSource {
		return platform.VariableSource{KeyValueStore: &platform.KeyValueStoreID{
			Workspace:   environmentID.Workspace,
			Project:     environmentID.Project,
			Environment: environmentID.Name,
			Name:        secret.Labels[label],
		}}
	}
}

func bucketSource(label string) func(platform.EnvironmentID, *core.Secret) platform.VariableSource {
	return func(environmentID platform.EnvironmentID, secret *core.Secret) platform.VariableSource {
		return platform.VariableSource{Bucket: &platform.BucketID{
			Workspace:   environmentID.Workspace,
			Project:     environmentID.Project,
			Environment: environmentID.Name,
			Name:        secret.Labels[label],
		}}
	}
}

func sharedSource(platform.EnvironmentID, *core.Secret) platform.VariableSource {
	return platform.VariableSource{Shared: true}
}

func offeredNames(secret *core.Secret, offeredKeys []string) []string {
	if len(offeredKeys) == 0 {
		names := make([]string, 0, len(secret.Data))

		for name := range secret.Data {
			names = append(names, name)
		}

		sort.Strings(names)

		return names
	}

	names := make([]string, 0, len(offeredKeys))

	for _, name := range offeredKeys {
		if _, ok := secret.Data[name]; ok {
			names = append(names, name)
		}
	}

	return names
}

func defaultVariableSources(client kubernetes.Interface) []variableSource {
	return []variableSource{
		newVariableSource(client, cnpgAppSecretSelector, []string{"fqdn-uri", "host", "port", "dbname", "user", "password"}, databaseSource(databaseLabel)),
		newVariableSource(client, sharedSecretSelector, nil, sharedSource),
		newVariableSource(client, keyValueStoreSelector, []string{"host", "port", "password", "uri"}, keyValueStoreSource(keyValueStoreLabel)),
		newVariableSource(client, objectStorageSelector, []string{"accessKeyId", "secretAccessKey", "endpoint", "region", "bucket"}, bucketSource(objectStorageBucketLabel)),
	}
}

func existsSelector(key string) labels.Selector {
	requirement, err := labels.NewRequirement(key, selection.Exists, nil)

	if err != nil {
		panic(err)
	}

	return labels.NewSelector().Add(*requirement)
}
