package mcpserver

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *server) registerResource(m *mcp.Server) {
	mcp.AddTool(m, &mcp.Tool{
		Name:        "create_database",
		Description: "Provision a managed PostgreSQL database in an environment. size is the storage size as a Kubernetes quantity (e.g. 10Gi). cpu and memory (e.g. '500m'/'512Mi') size the database; pass both or omit both for platform defaults. It accepts connections about a minute after creation, once its status leaves PENDING; pass wait to get the result at that point instead of polling get_project. UPDATING then means it serves while a standby instance catches up. Wire its credentials into a service with set_variables refs from list_variables' available list; credential values are never returned.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptr(false)},
	}, s.createDatabase)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "create_kv_store",
		Description: "Provision a Redis-compatible key-value store in an environment. Pass wait to get the result once it accepts connections (status no longer PENDING).",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptr(false)},
	}, s.createKeyValueStore)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "create_bucket",
		Description: "Provision an S3-compatible object storage bucket in an environment.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptr(false)},
	}, s.createBucket)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "create_volume",
		Description: "Provision a persistent volume in an environment. size is a Kubernetes quantity between 10Gi and 128Gi (e.g. 10Gi); it can grow later but never shrink. Optionally mount it into a service at a path; a volume mounts into one service only, and that service must run a single replica.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptr(false)},
	}, s.createVolume)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "run_sql",
		Description: "Run a SQL statement against a database, for schema setup and quick checks. Returns up to 200 rows. Bulk dump imports need the database credentials, which are never shown here, so the user runs those from their own terminal.",
	}, s.runSQL)
}

type createDatabaseInput struct {
	Environment string `json:"environment" jsonschema:"environment id (workspace/project/environment)"`
	Name        string `json:"name" jsonschema:"what people read, any text up to 128 chars (e.g. 'Orders'); it can be renamed later"`
	ID          string `json:"id,omitempty" jsonschema:"optional id (2-16 chars: lowercase letters, digits and hyphens, starting and ending with a letter or digit); derived from name when omitted, with a counter added if taken. Fixed once created. Use the id this tool returns, never the name, to refer to it later"`
	Size        string `json:"size,omitempty" jsonschema:"storage size as a Kubernetes quantity (e.g. 10Gi)"`
	CPU         string `json:"cpu,omitempty" jsonschema:"CPU limit as a Kubernetes quantity (e.g. 500m); provide together with memory, or omit both for platform defaults"`
	Memory      string `json:"memory,omitempty" jsonschema:"memory limit as a Kubernetes quantity (e.g. 512Mi); provide together with cpu, or omit both for platform defaults"`
	Wait        int    `json:"wait,omitempty" jsonschema:"optional: seconds (1-300) to wait until it accepts connections (status no longer PENDING) before answering, instead of polling get_project"`
}

func (s *server) createDatabase(ctx context.Context, _ *mcp.CallToolRequest, input createDatabaseInput) (*mcp.CallToolResult, any, error) {
	environmentID, err := s.environmentID(input.Environment)
	if err != nil {
		return nil, nil, err
	}
	if (input.CPU == "") != (input.Memory == "") {
		return nil, nil, fmt.Errorf("cpu and memory must be provided together")
	}
	if err := validateWait(input.Wait); err != nil {
		return nil, nil, err
	}

	dbInput := map[string]any{"environment": environmentID, "name": input.Name}
	if input.ID != "" {
		dbInput["id"] = input.ID
	}
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
		CreateDatabase map[string]any `json:"createDatabase"`
	}
	if err := s.query(ctx, "create_database", mutation, map[string]any{"input": dbInput}, &out); err != nil {
		return nil, nil, err
	}

	database, settled, err := pollUntil(ctx, input.Wait, func() (map[string]any, bool, error) {
		if input.Wait == 0 {
			return out.CreateDatabase, false, nil
		}
		const query = `query($id: DatabaseID!) { database(id: $id) { id name status statusReason version size public resources { cpu memory } } }`
		var current struct {
			Database map[string]any `json:"database"`
		}
		if err := s.query(ctx, "create_database (status)", query, map[string]any{"id": out.CreateDatabase["id"]}, &current); err != nil {
			return nil, false, err
		}
		return current.Database, resourceSettled(current.Database), nil
	})
	if err != nil {
		return nil, nil, err
	}

	note := "PostgreSQL accepts connections about a minute after creation, once its status leaves PENDING (call get_project, or pass wait next time); wire its credentials into a service via set_variables refs from list_variables' available list right away"
	if settled {
		note = "wire its credentials into a service via set_variables refs from list_variables' available list"
	} else if pending := waitNote(input.Wait, settled, database["status"]); pending != "" {
		note = pending
	}
	return jsonResult(map[string]any{"database": database, "note": note})
}

func resourceSettled(resource map[string]any) bool {
	status, _ := resource["status"].(string)
	return status != "" && status != "PENDING"
}

type createKeyValueStoreInput struct {
	Environment string `json:"environment" jsonschema:"environment id (workspace/project/environment)"`
	Name        string `json:"name" jsonschema:"what people read, any text up to 128 chars (e.g. 'Sessions'); it can be renamed later"`
	ID          string `json:"id,omitempty" jsonschema:"optional id (2-16 chars: lowercase letters, digits and hyphens, starting and ending with a letter or digit); derived from name when omitted, with a counter added if taken. Fixed once created. Use the id this tool returns, never the name, to refer to it later"`
	Wait        int    `json:"wait,omitempty" jsonschema:"optional: seconds (1-300) to wait until it accepts connections (status no longer PENDING) before answering, instead of polling get_project"`
}

func (s *server) createKeyValueStore(ctx context.Context, _ *mcp.CallToolRequest, input createKeyValueStoreInput) (*mcp.CallToolResult, any, error) {
	environmentID, err := s.environmentID(input.Environment)
	if err != nil {
		return nil, nil, err
	}
	if err := validateWait(input.Wait); err != nil {
		return nil, nil, err
	}

	const mutation = `mutation($input: CreateKeyValueStoreInput!) {
  createKeyValueStore(input: $input) { id name status version size }
}`
	var out struct {
		CreateKeyValueStore map[string]any `json:"createKeyValueStore"`
	}
	kvInput := map[string]any{"environment": environmentID, "name": input.Name}
	if input.ID != "" {
		kvInput["id"] = input.ID
	}
	if err := s.query(ctx, "create_kv_store", mutation, map[string]any{"input": kvInput}, &out); err != nil {
		return nil, nil, err
	}

	store, settled, err := pollUntil(ctx, input.Wait, func() (map[string]any, bool, error) {
		if input.Wait == 0 {
			return out.CreateKeyValueStore, false, nil
		}
		const query = `query($id: KeyValueStoreID!) { keyValueStore(id: $id) { id name status version size } }`
		var current struct {
			KeyValueStore map[string]any `json:"keyValueStore"`
		}
		if err := s.query(ctx, "create_kv_store (status)", query, map[string]any{"id": out.CreateKeyValueStore["id"]}, &current); err != nil {
			return nil, false, err
		}
		return current.KeyValueStore, resourceSettled(current.KeyValueStore), nil
	})
	if err != nil {
		return nil, nil, err
	}

	note := "Redis-compatible; wire its credentials into a service via set_variables refs from list_variables' available list"
	if pending := waitNote(input.Wait, settled, store["status"]); pending != "" {
		note = pending + ". " + note
	}
	return jsonResult(map[string]any{"keyValueStore": store, "note": note})
}

type createBucketInput struct {
	Environment string `json:"environment" jsonschema:"environment id (workspace/project/environment)"`
	Name        string `json:"name" jsonschema:"what people read, any text up to 128 chars (e.g. 'Uploads'); it can be renamed later"`
	ID          string `json:"id,omitempty" jsonschema:"optional id (2-16 chars: lowercase letters, digits and hyphens, starting and ending with a letter or digit); derived from name when omitted, with a counter added if taken. Fixed once created. Use the id this tool returns, never the name, to refer to it later"`
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
	if input.ID != "" {
		bucketInput["id"] = input.ID
	}
	if err := s.query(ctx, "create_bucket", mutation, map[string]any{"input": bucketInput}, &out); err != nil {
		return nil, nil, err
	}
	return jsonResult(map[string]any{
		"bucket": out.CreateBucket,
		"note":   "S3-compatible; wire its credentials into a service via set_variables refs from list_variables' available list",
	})
}

type createVolumeInput struct {
	Environment  string `json:"environment" jsonschema:"environment id (workspace/project/environment)"`
	Name         string `json:"name" jsonschema:"what people read, any text up to 128 chars (e.g. 'Data'); it can be renamed later"`
	ID           string `json:"id,omitempty" jsonschema:"optional id (2-16 chars: lowercase letters, digits and hyphens, starting and ending with a letter or digit); derived from name when omitted, with a counter added if taken. Fixed once created. Use the id this tool returns, never the name, to refer to it later"`
	Size         string `json:"size" jsonschema:"size as a Kubernetes quantity, 10Gi minimum, 128Gi maximum (e.g. 10Gi)"`
	MountService string `json:"mount_service,omitempty" jsonschema:"optional service id to mount the volume into"`
	MountPath    string `json:"mount_path,omitempty" jsonschema:"mount path inside the container; required when mount_service is set"`
}

func (s *server) createVolume(ctx context.Context, _ *mcp.CallToolRequest, input createVolumeInput) (*mcp.CallToolResult, any, error) {
	environmentID, err := s.environmentID(input.Environment)
	if err != nil {
		return nil, nil, err
	}

	if input.MountService != "" && input.MountPath == "" {
		return nil, nil, fmt.Errorf("mount_path is required when mount_service is set")
	}

	const mutation = `mutation($environment: EnvironmentID!, $name: String!, $id: String, $size: String!) {
  createVolume(environment: $environment, name: $name, id: $id, size: $size) { id name size mount { service path } }
}`
	var out struct {
		CreateVolume struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Size  string `json:"size"`
			Mount any    `json:"mount"`
		} `json:"createVolume"`
	}
	variables := map[string]any{"environment": environmentID, "name": input.Name, "size": input.Size}
	if input.ID != "" {
		variables["id"] = input.ID
	}
	if err := s.query(ctx, "create_volume", mutation, variables, &out); err != nil {
		return nil, nil, err
	}

	volumeID := out.CreateVolume.ID
	result := map[string]any{"volume": out.CreateVolume}

	if input.MountService != "" {
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
	Query    string `json:"query" jsonschema:"SQL statement to execute, e.g. select count(*) from users"`
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
