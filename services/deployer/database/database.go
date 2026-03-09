package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"

	"github.com/jackc/pgx/v5"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/zeitlos/lucity/pkg/labels"
)

// ErrNotReady indicates the database is still provisioning and credentials are not yet available.
var ErrNotReady = errors.New("database not ready")

// Credentials holds the connection info read from a CNPG secret.
type Credentials struct {
	Host     string
	Port     string
	DBName   string
	User     string
	Password string
}

// CredentialsFromSecret reads CNPG connection credentials from the K8s secret.
// CNPG cluster naming: {project}-pg-{dbname} (matches fullnameOverride = project)
// CNPG secret naming: {clustername}-app
func CredentialsFromSecret(ctx context.Context, k8s kubernetes.Interface, workspace, project, environment, database string) (*Credentials, error) {
	namespace := labels.NamespaceFor(workspace, project, environment)
	clusterName := project + "-pg-" + database
	secretName := clusterName + "-app"

	secret, err := k8s.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, ErrNotReady
		}
		return nil, fmt.Errorf("failed to get CNPG secret %q in namespace %q: %w", secretName, namespace, err)
	}

	// CNPG stores the short service name (e.g. "myns-lucity-app-pg-main-rw").
	// Qualify with namespace for cross-namespace DNS resolution.
	host := string(secret.Data["host"])
	if !strings.Contains(host, ".") {
		host = host + "." + namespace + ".svc.cluster.local"
	}

	return &Credentials{
		Host:     host,
		Port:     string(secret.Data["port"]),
		DBName:   string(secret.Data["dbname"]),
		User:     string(secret.Data["user"]),
		Password: string(secret.Data["password"]),
	}, nil
}

// Connect creates a short-lived connection to the database.
// Caller is responsible for closing: defer conn.Close(ctx)
//
// When the CNPG host (a K8s service name) is unresolvable — typical for local
// development — Connect falls back to localhost, expecting an active
// kubectl port-forward (see: make db-forward).
func Connect(ctx context.Context, creds *Credentials) (*pgx.Conn, error) {
	host := creds.Host
	if _, err := net.LookupHost(host); err != nil {
		slog.Debug("CNPG host unresolvable, falling back to localhost", "host", host)
		host = "localhost"
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		creds.User, creds.Password, host, creds.Port, creds.DBName)

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		if host == "localhost" {
			slog.Debug("database unreachable via localhost, port-forward may not be running",
				"originalHost", creds.Host,
				"hint", fmt.Sprintf("kubectl port-forward svc/%s %s:%s", creds.Host, creds.Port, creds.Port),
				"error", err,
			)
		}
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	slog.Debug("connected to database", "host", host, "dbname", creds.DBName)
	return conn, nil
}
