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
		Description: "Provision a managed PostgreSQL database in an environment. size is the storage size as a Kubernetes quantity (e.g. 10Gi). cpu and memory (e.g. '500m'/'512Mi') size the database; pass both or omit both for platform defaults. Provisioning takes ~1-2 min; poll get_project for status. Wire its credentials into a service with set_variables refs from list_variables' available list; credential values are never returned.",
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
		Description: "Provision a persistent volume in an environment. size is a Kubernetes quantity between 10Gi and 1Ti (e.g. 10Gi); it can grow later but never shrink. Optionally mount it into a service at a path; a volume mounts into one service only, and that service must run a single replica.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptr(false)},
	}, s.createVolume)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "run_sql",
		Description: "Run a SQL statement against a database, for schema setup and quick checks. Returns up to 200 rows. Bulk dump imports need the database credentials, which are never shown here, so the user runs those from their own terminal.",
	}, s.runSQL)
}

type createDatabaseInput struct {
	Environment string `json:"environment" jsonschema:"environment id (workspace/project/environment)"`
	Name        string `json:"name" jsonschema:"database name (2-16 chars, lowercase alphanumeric and hyphens)"`
	Size        string `json:"size,omitempty" jsonschema:"storage size as a Kubernetes quantity (e.g. 10Gi)"`
	CPU         string `json:"cpu,omitempty" jsonschema:"CPU limit as a Kubernetes quantity (e.g. 500m); provide together with memory, or omit both for platform defaults"`
	Memory      string `json:"memory,omitempty" jsonschema:"memory limit as a Kubernetes quantity (e.g. 512Mi); provide together with cpu, or omit both for platform defaults"`
}

func (s *server) createDatabase(ctx context.Context, _ *mcp.CallToolRequest, input createDatabaseInput) (*mcp.CallToolResult, any, error) {
	environmentID, err := s.environmentID(input.Environment)
	if err != nil {
		return nil, nil, err
	}
	if (input.CPU == "") != (input.Memory == "") {
		return nil, nil, fmt.Errorf("cpu and memory must be provided together")
	}

	dbInput := map[string]any{"environment": environmentID, "name": input.Name}
	if input.Size != "" {
		dbInput["size"] = input.Size
	}
	if input.CPU != "" {
		dbInput["resources"] = map[string]any{"cpu": input.CPU, "memory": input.Memory}
	}

	const mutation = `mutation($input: CreateDatabaseInput!) {
  createDatabase(input: $input) { id name status version size public resources { cpu memory } }
}`
	var out struct {
		CreateDatabase any `json:"createDatabase"`
	}
	if err := s.query(ctx, "create_database", mutation, map[string]any{"input": dbInput}, &out); err != nil {
		return nil, nil, err
	}
	return jsonResult(map[string]any{
		"database": withResourceID(out.CreateDatabase, environmentID+"/"+input.Name),
		"note":     "PostgreSQL provisioning takes ~1-2 min; poll get_project until status is HEALTHY, then wire its credentials into a service via set_variables refs from list_variables' available list",
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
		"note":          "Redis-compatible; wire its credentials into a service via set_variables refs from list_variables' available list",
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
		"note":   "S3-compatible; wire its credentials into a service via set_variables refs from list_variables' available list",
	})
}

type createVolumeInput struct {
	Environment  string `json:"environment" jsonschema:"environment id (workspace/project/environment)"`
	Name         string `json:"name" jsonschema:"volume name (2-16 chars, lowercase alphanumeric and hyphens)"`
	Size         string `json:"size" jsonschema:"size as a Kubernetes quantity, 10Gi minimum, 1Ti maximum (e.g. 10Gi)"`
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
