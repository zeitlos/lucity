package gitops

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// parseServiceInstanceMetas extracts service metadata from a YAML services map.
func parseServiceInstanceMetas(services map[string]any) []ServiceInstanceMeta {
	var result []ServiceInstanceMeta
	for svcName, svcRaw := range services {
		svcMap, ok := svcRaw.(map[string]any)
		if !ok {
			continue
		}

		meta := ServiceInstanceMeta{Name: svcName}
		if imageMap, ok := svcMap["image"].(map[string]any); ok {
			if tag, ok := imageMap["tag"].(string); ok {
				meta.ImageTag = tag
			}
		}
		if host, ok := svcMap["host"].(string); ok {
			meta.Host = host
		}

		result = append(result, meta)
	}
	return result
}

// parseStringMap extracts a map[string]string from a nested YAML map.
func parseStringMap(m map[string]any, key string) map[string]string {
	raw, ok := m[key].(map[string]any)
	if !ok {
		return nil
	}
	result := make(map[string]string, len(raw))
	for k, v := range raw {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result
}

// parseStringSlice extracts a []string from a nested YAML list.
func parseStringSlice(m map[string]any, key string) []string {
	raw, ok := m[key].([]any)
	if !ok {
		return nil
	}
	result := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

// parseDatabaseRefs extracts database references from a YAML service map.
func parseDatabaseRefs(m map[string]any) map[string]DatabaseRef {
	raw, ok := m["databaseRefs"].(map[string]any)
	if !ok {
		return nil
	}
	result := make(map[string]DatabaseRef, len(raw))
	for k, v := range raw {
		ref, ok := v.(map[string]any)
		if !ok {
			continue
		}
		result[k] = DatabaseRef{
			Database: fmt.Sprintf("%v", ref["database"]),
			Key:      fmt.Sprintf("%v", ref["key"]),
		}
	}
	return result
}

// parseServiceRefs extracts service references from a YAML service map.
func parseServiceRefs(m map[string]any) map[string]ServiceRef {
	raw, ok := m["serviceRefs"].(map[string]any)
	if !ok {
		return nil
	}
	result := make(map[string]ServiceRef, len(raw))
	for k, v := range raw {
		ref, ok := v.(map[string]any)
		if !ok {
			continue
		}
		result[k] = ServiceRef{
			Service: fmt.Sprintf("%v", ref["service"]),
		}
	}
	return result
}

// databaseRefsToAny converts database refs to map[string]any for YAML marshaling.
func databaseRefsToAny(refs map[string]DatabaseRef) map[string]any {
	result := make(map[string]any, len(refs))
	for k, v := range refs {
		result[k] = map[string]any{
			"database": v.Database,
			"key":      v.Key,
		}
	}
	return result
}

// serviceRefsToAny converts service refs to map[string]any for YAML marshaling.
func serviceRefsToAny(refs map[string]ServiceRef) map[string]any {
	result := make(map[string]any, len(refs))
	for k, v := range refs {
		result[k] = map[string]any{
			"service": v.Service,
		}
	}
	return result
}

// stringMapToAny converts map[string]string to map[string]any for YAML marshaling.
func stringMapToAny(m map[string]string) map[string]any {
	result := make(map[string]any, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

// projectYAMLData matches the structure of project.yaml for parsing.
type projectYAMLData struct {
	Name      string `yaml:"name"`
	CreatedAt string `yaml:"created_at"`
}

// parseProjectYAML parses a project.yaml file into ProjectMeta.
func parseProjectYAML(data []byte) (*ProjectMeta, error) {
	var d projectYAMLData
	if err := yaml.Unmarshal(data, &d); err != nil {
		return nil, fmt.Errorf("failed to unmarshal project.yaml: %w", err)
	}

	createdAt, _ := time.Parse(time.RFC3339, d.CreatedAt)

	return &ProjectMeta{
		Name:      d.Name,
		CreatedAt: createdAt,
	}, nil
}

// parseDatabaseDefs extracts database definitions from a YAML values map.
func parseDatabaseDefs(inner map[string]any) []DatabaseDef {
	databases, ok := inner["databases"].(map[string]any)
	if !ok {
		return nil
	}
	postgres, ok := databases["postgres"].(map[string]any)
	if !ok {
		return nil
	}

	var result []DatabaseDef
	for dbName, dbRaw := range postgres {
		dbMap, ok := dbRaw.(map[string]any)
		if !ok {
			continue
		}
		def := DatabaseDef{Name: dbName}
		if v, ok := dbMap["version"].(string); ok {
			def.Version = v
		}
		if v, ok := dbMap["instances"].(int); ok {
			def.Instances = v
		}
		if v, ok := dbMap["size"].(string); ok {
			def.Size = v
		}
		result = append(result, def)
	}
	return result
}
