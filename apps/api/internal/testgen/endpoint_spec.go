package testgen

import "strings"

// buildEndpointSpec translates a single method+path+fields description into
// the same map[string]interface{} shape an OpenAPI 3.0/3.1 document would
// produce for one path/operation. This lets GenerateFromEndpoint reuse
// GenerateDraftCases (openapi_parser.go) completely unchanged instead of
// duplicating any rule logic — the quick-generate form is just a different
// way to build a (one-endpoint) spec.
func buildEndpointSpec(method, path string, fields []EndpointField, requiresAuth bool) map[string]interface{} {
	var parameters []interface{}
	bodyProps := map[string]interface{}{}
	var bodyRequired []interface{}

	for _, f := range fields {
		name := strings.TrimSpace(f.Name)
		if name == "" {
			continue
		}
		schema := fieldSchema(f)

		if f.Location == "body" {
			bodyProps[name] = schema
			if f.Required {
				bodyRequired = append(bodyRequired, name)
			}
			continue
		}

		loc := f.Location
		if loc != "query" && loc != "path" && loc != "header" {
			loc = "query"
		}
		parameters = append(parameters, map[string]interface{}{
			"name":     name,
			"in":       loc,
			"required": f.Required,
			"schema":   schema,
		})
	}

	operation := map[string]interface{}{
		"parameters": parameters,
		"responses":  map[string]interface{}{"200": map[string]interface{}{"description": "OK"}},
	}
	if requiresAuth {
		operation["security"] = []interface{}{map[string]interface{}{"bearerAuth": []interface{}{}}}
	}
	if len(bodyProps) > 0 {
		bodySchema := map[string]interface{}{
			"type":       "object",
			"properties": bodyProps,
		}
		if len(bodyRequired) > 0 {
			bodySchema["required"] = bodyRequired
		}
		operation["requestBody"] = map[string]interface{}{
			"required": true,
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": bodySchema,
				},
			},
		}
	}

	return map[string]interface{}{
		"paths": map[string]interface{}{
			path: map[string]interface{}{
				strings.ToLower(method): operation,
			},
		},
	}
}

func fieldSchema(f EndpointField) map[string]interface{} {
	t := f.Type
	if t == "" {
		t = "string"
	}
	schema := map[string]interface{}{"type": t}
	if len(f.Enum) > 0 {
		enum := make([]interface{}, len(f.Enum))
		for i, v := range f.Enum {
			enum[i] = v
		}
		schema["enum"] = enum
	}
	return schema
}
