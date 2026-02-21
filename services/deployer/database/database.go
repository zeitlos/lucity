package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/zeitlos/lucity/pkg/labels"
)

// Credentials holds the connection info read from a CNPG secret.
type Credentials struct {
	Host     string
	Port     string
	DBName   string
	User     string
	Password string
}

// CredentialsFromSecret reads CNPG connection credentials from the K8s secret.
// CNPG cluster naming: {namespace}-lucity-app-pg-{dbname}
// CNPG secret naming: {clustername}-app
func CredentialsFromSecret(ctx context.Context, k8s kubernetes.Interface, project, environment, database string) (*Credentials, error) {
	namespace := labels.NamespaceFor(project, environment)
	clusterName := namespace + "-lucity-app-pg-" + database
	secretName := clusterName + "-app"

	secret, err := k8s.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get CNPG secret %q in namespace %q: %w", secretName, namespace, err)
	}

	return &Credentials{
		Host:     string(secret.Data["host"]),
		Port:     string(secret.Data["port"]),
		DBName:   string(secret.Data["dbname"]),
		User:     string(secret.Data["user"]),
		Password: string(secret.Data["password"]),
	}, nil
}

// Connect creates a short-lived connection to the database.
// Caller is responsible for closing: defer conn.Close(ctx)
func Connect(ctx context.Context, creds *Credentials) (*pgx.Conn, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		creds.User, creds.Password, creds.Host, creds.Port, creds.DBName)

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	slog.Debug("connected to database", "host", creds.Host, "dbname", creds.DBName)
	return conn, nil
}
