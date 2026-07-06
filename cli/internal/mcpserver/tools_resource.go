package mcpserver

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func withResourceID(resource any, id string) any {
	if fields, ok := resource.(map[string]any); ok {
		if current, set := fields["id"].(string); !set || current == "" || strings.Trim(current, "/") == "" {
			fields["id"] = id
		}
		return fields
	}
	return resource
}

func (s *server) registerResource(m *mcp.Server) {
	mcp.AddTool(m, &mcp.Tool{
		Name:        "create_database",
		Description: "Provision a managed PostgreSQL database in an environment. size is a Kubernetes quantity (e.g. 10Gi). Provisioning takes ~1-2 min; poll get_project for status. Wire credentials into a service with set_variables (ref) or read them with get_credentials.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptr(false)},
	}, s.createDatabase)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "create_kv_store",
		Description: "Provision a Redis-compatible key-value store in an environment.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptr(false)},
	}, s.createKeyValueStore)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "create_bucket",
		Description: "Provision an S3-compatible object storage bucket in an environment.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptr(false)},
	}, s.createBucket)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "create_volume",
		Description: "Provision a persistent volume in an environment. size is a Kubernetes quantity (e.g. 5Gi). Optionally mount it into a service at a path.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptr(false)},
	}, s.createVolume)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "get_credentials",
		Description: "Read connection credentials for a database, key-value store, or bucket. For databases, expose_publicly first opens a public internet endpoint (public endpoints require TLS with SNI: psql 'sslmode=require', libpq >= 14).",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptr(false)},
	}, s.getCredentials)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "run_sql",
		Description: "Run a SQL statement against a database, for schema setup and quick checks. Returns up to 200 rows. For bulk imports prefer get_credentials(expose_publicly) plus a local psql.",
	}, s.runSQL)
}

type createDatabaseInput struct {
	Environment string `json:"environment" jsonschema:"environment id (workspace/project/environment)"`
	Name        string `json:"name" jsonschema:"database name (2-16 chars, lowercase alphanumeric and hyphens)"`
	Size        string `json:"size,omitempty" jsonschema:"storage size as a Kubernetes quantity (e.g. 10Gi)"`
}

func (s *server) createDatabase(ctx context.Context, _ *mcp.CallToolRequest, input createDatabaseInput) (*mcp.CallToolResult, any, error) {
	environmentID, err := s.environmentID(input.Environment)
	if err != nil {
		return nil, nil, err
	}

	dbInput := map[string]any{"environment": environmentID, "name": input.Name}
	if input.Size != "" {
		dbInput["size"] = input.Size
	}

	const mutation = `mutation($input: CreateDatabaseInput!) {
  createDatabase(input: $input) { id name status version size public }
}`
	var out struct {
		CreateDatabase any `json:"createDatabase"`
	}
	if err := s.query(ctx, "create_database", mutation, map[string]any{"input": dbInput}, &out); err != nil {
		return nil, nil, err
	}
	return jsonResult(map[string]any{
		"database": withResourceID(out.CreateDatabase, environmentID+"/"+input.Name),
		"note":     "PostgreSQL provisioning takes ~1-2 min; poll get_project until status is HEALTHY, then wire credentials via set_variables (ref) or read them with get_credentials",
	})
}

type createKeyValueStoreInput struct {
	Environment string `json:"environment" jsonschema:"environment id (workspace/project/environment)"`
	Name        string `json:"name" jsonschema:"store name (2-16 chars, lowercase alphanumeric and hyphens)"`
}

func (s *server) createKeyValueStore(ctx context.Context, _ *mcp.CallToolRequest, input createKeyValueStoreInput) (*mcp.CallToolResult, any, error) {
	environmentID, err := s.environmentID(input.Environment)
	if err != nil {
		return nil, nil, err
	}

	const mutation = `mutation($input: CreateKeyValueStoreInput!) {
  createKeyValueStore(input: $input) { id name status version size }
}`
	var out struct {
		CreateKeyValueStore any `json:"createKeyValueStore"`
	}
	kvInput := map[string]any{"environment": environmentID, "name": input.Name}
	if err := s.query(ctx, "create_kv_store", mutation, map[string]any{"input": kvInput}, &out); err != nil {
		return nil, nil, err
	}
	return jsonResult(map[string]any{
		"keyValueStore": withResourceID(out.CreateKeyValueStore, environmentID+"/"+input.Name),
		"note":          "Redis-compatible; read credentials with get_credentials(kind=kv_store)",
	})
}

type createBucketInput struct {
	Environment string `json:"environment" jsonschema:"environment id (workspace/project/environment)"`
	Name        string `json:"name" jsonschema:"bucket name (2-16 chars, lowercase alphanumeric and hyphens)"`
}

func (s *server) createBucket(ctx context.Context, _ *mcp.CallToolRequest, input createBucketInput) (*mcp.CallToolResult, any, error) {
	environmentID, err := s.environmentID(input.Environment)
	if err != nil {
		return nil, nil, err
	}

	const mutation = `mutation($input: CreateBucketInput!) {
  createBucket(input: $input) { id name status region endpoint public }
}`
	var out struct {
		CreateBucket any `json:"createBucket"`
	}
	bucketInput := map[string]any{"environment": environmentID, "name": input.Name}
	if err := s.query(ctx, "create_bucket", mutation, map[string]any{"input": bucketInput}, &out); err != nil {
		return nil, nil, err
	}
	return jsonResult(map[string]any{
		"bucket": withResourceID(out.CreateBucket, environmentID+"/"+input.Name),
		"note":   "S3-compatible; read credentials with get_credentials(kind=bucket)",
	})
}

type createVolumeInput struct {
	Environment  string `json:"environment" jsonschema:"environment id (workspace/project/environment)"`
	Name         string `json:"name" jsonschema:"volume name (2-16 chars, lowercase alphanumeric and hyphens)"`
	Size         string `json:"size" jsonschema:"size as a Kubernetes quantity (e.g. 5Gi)"`
	MountService string `json:"mount_service,omitempty" jsonschema:"optional service id to mount the volume into"`
	MountPath    string `json:"mount_path,omitempty" jsonschema:"mount path inside the container; required when mount_service is set"`
}

func (s *server) createVolume(ctx context.Context, _ *mcp.CallToolRequest, input createVolumeInput) (*mcp.CallToolResult, any, error) {
	environmentID, err := s.environmentID(input.Environment)
	if err != nil {
		return nil, nil, err
	}

	const mutation = `mutation($environment: EnvironmentID!, $name: String!, $size: String!) {
  createVolume(environment: $environment, name: $name, size: $size) { id name size mount { service path } }
}`
	var out struct {
		CreateVolume struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Size  string `json:"size"`
			Mount any    `json:"mount"`
		} `json:"createVolume"`
	}
	if err := s.query(ctx, "create_volume", mutation, map[string]any{"environment": environmentID, "name": input.Name, "size": input.Size}, &out); err != nil {
		return nil, nil, err
	}

	volumeID := environmentID + "/" + input.Name
	out.CreateVolume.ID = volumeID
	result := map[string]any{"volume": out.CreateVolume}

	if input.MountService != "" {
		if input.MountPath == "" {
			return nil, nil, fmt.Errorf("mount_path is required when mount_service is set")
		}
		serviceID, err := s.serviceID(input.MountService)
		if err != nil {
			return nil, nil, err
		}
		const mountMutation = `mutation($volume: VolumeID!, $service: ServiceID!, $path: String!) {
  mountVolume(volume: $volume, service: $service, path: $path) { id name size mount { service path } }
}`
		var mounted struct {
			MountVolume any `json:"mountVolume"`
		}
		if err := s.query(ctx, "create_volume (mount)", mountMutation, map[string]any{"volume": volumeID, "service": serviceID, "path": input.MountPath}, &mounted); err != nil {
			return nil, nil, err
		}
		result["volume"] = mounted.MountVolume
		result["note"] = "mounted: the service rolls out with its current image to attach the volume (no rebuild)"
	}

	return jsonResult(result)
}

type getCredentialsInput struct {
	Kind           string `json:"kind" jsonschema:"one of database, kv_store, bucket"`
	ID             string `json:"id" jsonschema:"resource id (workspace/project/environment/name)"`
	ExposePublicly bool   `json:"expose_publicly,omitempty" jsonschema:"database only: open a public internet endpoint before reading credentials"`
}

func (s *server) getCredentials(ctx context.Context, _ *mcp.CallToolRequest, input getCredentialsInput) (*mcp.CallToolResult, any, error) {
	id, err := s.resourceID(input.ID)
	if err != nil {
		return nil, nil, err
	}

	switch input.Kind {
	case "database":
		if input.ExposePublicly {
			const mutation = `mutation($database: DatabaseID!) { exposeDatabase(database: $database) { id public } }`
			if err := s.query(ctx, "get_credentials (expose)", mutation, map[string]any{"database": id}, nil); err != nil {
				return nil, nil, err
			}
		}
		const query = `query($database: DatabaseID!) {
  databaseCredentials(database: $database) { type host port dbname user password uri }
}`
		var out struct {
			DatabaseCredentials any `json:"databaseCredentials"`
		}
		if err := s.query(ctx, "get_credentials", query, map[string]any{"database": id}, &out); err != nil {
			return nil, nil, err
		}
		result := map[string]any{"credentials": out.DatabaseCredentials}
		if input.ExposePublicly {
			result["note"] = "public endpoints (type PLATFORM) require TLS with SNI: connect with 'sslmode=require' and libpq >= 14"
		}
		return jsonResult(result)

	case "kv_store":
		if input.ExposePublicly {
			return nil, nil, fmt.Errorf("expose_publicly is only valid for kind=database")
		}
		const query = `query($keyValueStore: KeyValueStoreID!) {
  keyValueStoreCredentials(keyValueStore: $keyValueStore) { type host port password uri }
}`
		var out struct {
			KeyValueStoreCredentials any `json:"keyValueStoreCredentials"`
		}
		if err := s.query(ctx, "get_credentials", query, map[string]any{"keyValueStore": id}, &out); err != nil {
			return nil, nil, err
		}
		return jsonResult(map[string]any{"credentials": out.KeyValueStoreCredentials})

	case "bucket":
		if input.ExposePublicly {
			return nil, nil, fmt.Errorf("expose_publicly is only valid for kind=database")
		}
		const query = `query($bucket: BucketID!) {
  bucketCredentials(bucket: $bucket) { endpoint region bucket accessKeyId secretAccessKey }
}`
		var out struct {
			BucketCredentials any `json:"bucketCredentials"`
		}
		if err := s.query(ctx, "get_credentials", query, map[string]any{"bucket": id}, &out); err != nil {
			return nil, nil, err
		}
		return jsonResult(map[string]any{"credentials": out.BucketCredentials})

	default:
		return nil, nil, fmt.Errorf("invalid kind %q — use one of database, kv_store, bucket", input.Kind)
	}
}

type runSQLInput struct {
	Database string `json:"database" jsonschema:"database id (workspace/project/environment/name)"`
	Query    string `json:"query" jsonschema:"SQL statement to execute"`
}

func (s *server) runSQL(ctx context.Context, _ *mcp.CallToolRequest, input runSQLInput) (*mcp.CallToolResult, any, error) {
	databaseID, err := s.resourceID(input.Database)
	if err != nil {
		return nil, nil, err
	}

	const mutation = `mutation($database: DatabaseID!, $query: String!) {
  executeQuery(database: $database, query: $query) { columns rows affectedRows }
}`
	var out struct {
		ExecuteQuery struct {
			Columns      []string    `json:"columns"`
			Rows         [][]*string `json:"rows"`
			AffectedRows int         `json:"affectedRows"`
		} `json:"executeQuery"`
	}
	if err := s.query(ctx, "run_sql", mutation, map[string]any{"database": databaseID, "query": input.Query}, &out); err != nil {
		return nil, nil, err
	}

	total := len(out.ExecuteQuery.Rows)
	rows := out.ExecuteQuery.Rows
	result := map[string]any{
		"columns":      out.ExecuteQuery.Columns,
		"affectedRows": out.ExecuteQuery.AffectedRows,
	}
	if total > 200 {
		rows = rows[:200]
		result["note"] = fmt.Sprintf("truncated: showing 200 of %d rows", total)
	}
	result["rows"] = rows
	return jsonResult(result)
}
