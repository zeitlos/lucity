package values

import "github.com/zeitlos/lucity/pkg/usererror"

func withDisplayName(annotations map[string]string, displayName string) map[string]string {
	if displayName == "" {
		return annotations
	}

	if annotations == nil {
		annotations = map[string]string{}
	}

	annotations[annotationName] = displayName

	return annotations
}

func SetServiceDisplayName(env *Env, name, displayName string) error {
	return mutateService(env, name, func(s *Service) {
		s.Annotations = withDisplayName(s.Annotations, displayName)
	})
}

func SetDatabaseDisplayName(env *Env, name, displayName string) error {
	return mutateDatabase(env, name, func(p *Postgres) {
		p.Annotations = withDisplayName(p.Annotations, displayName)
	})
}

func SetKeyValueStoreDisplayName(env *Env, name, displayName string) error {
	store, ok := env.Databases.Valkey[name]

	if !ok {
		return usererror.Errorf("key-value store %q not found", name)
	}

	store.Annotations = withDisplayName(store.Annotations, displayName)
	env.Databases.Valkey[name] = store

	return nil
}

func SetVolumeDisplayName(env *Env, name, displayName string) error {
	volume, ok := env.Volumes[name]

	if !ok {
		return usererror.Errorf("volume %q not found", name)
	}

	volume.Annotations = withDisplayName(volume.Annotations, displayName)
	env.Volumes[name] = volume

	return nil
}
