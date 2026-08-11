package values

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/rand"
)

type Databases struct {
	Postgres    map[string]Postgres `yaml:"postgres"`
	Valkey      map[string]Valkey   `yaml:"valkey"`
	BackupStore BackupStore         `yaml:"backupStore"`
}

type BackupStore struct {
	Enabled         bool              `yaml:"enabled"`
	Endpoint        string            `yaml:"endpoint,omitempty"`
	DestinationPath string            `yaml:"destinationPath,omitempty"`
	Retention       string            `yaml:"retention,omitempty"`
	Schedule        string            `yaml:"schedule,omitempty"`
	Credentials     BackupCredentials `yaml:"credentials,omitempty"`
}

type BackupCredentials struct {
	ExistingSecret string           `yaml:"existingSecret,omitempty"`
	SecretKeys     BackupSecretKeys `yaml:"secretKeys,omitempty"`
}

type BackupSecretKeys struct {
	AccessKeyIDKey     string `yaml:"accessKeyIdKey,omitempty"`
	SecretAccessKeyKey string `yaml:"secretAccessKeyKey,omitempty"`
	RegionKey          string `yaml:"regionKey,omitempty"`
}

type Postgres struct {
	Instances   int               `yaml:"instances,omitempty"`
	Size        string            `yaml:"size,omitempty"`
	Version     string            `yaml:"version,omitempty"`
	PublicHost  string            `yaml:"publicHost,omitempty"`
	Resources   Resources         `yaml:"resources,omitempty"`
	Parameters  map[string]string `yaml:"parameters,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
	Backup      *PostgresBackup   `yaml:"backup,omitempty"`
	Restore     *PostgresRestore  `yaml:"restore,omitempty"`
}

type PostgresBackup struct {
	ServerName string `yaml:"serverName"`
}

type PostgresRestore struct {
	SourceServerName string `yaml:"sourceServerName"`
	TargetTime       string `yaml:"targetTime,omitempty"`
}

type DatabaseSpec struct {
	Version    string
	Size       resource.Quantity
	Resources  Resources
	Parameters map[string]string
	Restore    *PostgresRestore
}

func CreateDatabase(env *Env, name string, spec DatabaseSpec) error {
	if !isValidName(name) {
		return fmt.Errorf("invalid database name %q", name)
	}

	if _, ok := env.Databases.Postgres[name]; ok {
		// To keep this function idempotent, don't return an error if the database already exists.
		return nil
	}

	if env.Databases.Postgres == nil {
		env.Databases.Postgres = map[string]Postgres{}
	}

	env.Databases.Postgres[name] = Postgres{
		Version:    spec.Version,
		Size:       spec.Size.String(),
		Resources:  spec.Resources,
		Parameters: spec.Parameters,
		Labels:     map[string]string{labelDatabase: name},
		Restore:    spec.Restore,
	}

	return nil
}

func RestoreDatabase(env *Env, source, name string, spec DatabaseSpec, targetTime *time.Time) error {
	origin, ok := env.Databases.Postgres[source]

	if !ok {
		return fmt.Errorf("database %q not found", source)
	}

	if origin.Backup == nil || origin.Backup.ServerName == "" {
		return fmt.Errorf("database %q has no backups to restore from", source)
	}

	if _, ok := env.Databases.Postgres[name]; ok {
		return fmt.Errorf("database %q already exists", name)
	}

	spec.Restore = &PostgresRestore{
		SourceServerName: origin.Backup.ServerName,
	}

	if targetTime != nil {
		spec.Restore.TargetTime = targetTime.UTC().Format(time.RFC3339)
	}

	return CreateDatabase(env, name, spec)
}

func StripBackups(env *Env) {
	env.Databases.BackupStore = BackupStore{}

	for name, postgres := range env.Databases.Postgres {
		postgres.Backup = nil
		postgres.Restore = nil
		env.Databases.Postgres[name] = postgres
	}
}

func EnsureBackupServerNames(env *Env, namespace string) {
	for name, postgres := range env.Databases.Postgres {
		if postgres.Backup != nil && postgres.Backup.ServerName != "" {
			continue
		}

		postgres.Backup = &PostgresBackup{
			ServerName: fmt.Sprintf("%s-%s-%s", namespace, name, rand.String(6)),
		}

		env.Databases.Postgres[name] = postgres
	}
}

func SetDatabaseResources(env *Env, name string, resources Resources) error {
	return mutateDatabase(env, name, func(p *Postgres) {
		p.Resources = resources
	})
}

func SetDatabaseParameters(env *Env, name string, parameters map[string]string) error {
	return mutateDatabase(env, name, func(p *Postgres) {
		p.Parameters = parameters
	})
}

func SetDatabaseStorage(env *Env, name, size string) error {
	return mutateDatabase(env, name, func(p *Postgres) {
		p.Size = size
	})
}

func DeleteDatabase(env *Env, name string) error {
	if _, ok := env.Databases.Postgres[name]; !ok {
		return fmt.Errorf("database %q not found", name)
	}

	delete(env.Databases.Postgres, name)

	return nil
}

func ExposeDatabase(env *Env, name, host string) error {
	if !isValidHostname(host) {
		return fmt.Errorf("invalid hostname %q", host)
	}

	return mutateDatabase(env, name, func(p *Postgres) {
		p.PublicHost = host

		if p.Annotations == nil {
			p.Annotations = map[string]string{}
		}

		p.Annotations[annotationDatabaseHost] = host
	})
}

func UnexposeDatabase(env *Env, name string) error {
	return mutateDatabase(env, name, func(p *Postgres) {
		p.PublicHost = ""
		delete(p.Annotations, annotationDatabaseHost)
	})
}

func mutateDatabase(env *Env, name string, mutate func(*Postgres)) error {
	postgres, ok := env.Databases.Postgres[name]

	if !ok {
		return fmt.Errorf("database %q not found", name)
	}

	mutate(&postgres)
	env.Databases.Postgres[name] = postgres

	return nil
}
