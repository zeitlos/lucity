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
	cnpgUserTypeLabel = "cnpg.io/userType"
	cnpgAppUserType   = "app"
)

var (
	cnpgAppSecretSelector = labels.SelectorFromSet(labels.Set{cnpgUserTypeLabel: cnpgAppUserType})
	sharedSecretSelector  = labels.SelectorFromSet(labels.Set{sharedVariablesLabel: "true"})
	keyValueStoreSelector = existsSelector(keyValueStoreLabel)
)

type variableSource func(ctx context.Context, environmentID platform.EnvironmentID) ([]platform.Variable, error)

func newVariableSource(client kubernetes.Interface, selector labels.Selector, offeredKeys []string) variableSource {
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

			for _, name := range offeredNames(secret, offeredKeys) {
				variables = append(variables, platform.Variable{
					ID: platform.VariableID{
						Workspace:   environmentID.Workspace,
						Project:     environmentID.Project,
						Environment: environmentID.Name,
						Secret:      secret.Name,
						Name:        name,
					},
					Name: name,
				})
			}
		}

		return variables, nil
	}
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
		newVariableSource(client, cnpgAppSecretSelector, []string{"uri", "host", "port", "dbname", "user", "password"}),
		newVariableSource(client, sharedSecretSelector, nil),
		newVariableSource(client, keyValueStoreSelector, []string{"host", "port", "password", "uri"}),
	}
}

func existsSelector(key string) labels.Selector {
	requirement, err := labels.NewRequirement(key, selection.Exists, nil)

	if err != nil {
		panic(err)
	}

	return labels.NewSelector().Add(*requirement)
}

func (c *Client) SharedVariables(ctx context.Context, environmentID platform.EnvironmentID) ([]platform.Variable, error) {
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
