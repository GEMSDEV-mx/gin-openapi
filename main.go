package openapi

import (
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// OpenAPI defines the OpenAPI schema.
type OpenAPI struct {
	Openapi string                 `json:"openapi"`
	Info    map[string]string      `json:"info"`
	Paths   map[string]interface{} `json:"paths"`
}

// Route represents an API route for OpenAPI documentation.
type Route struct {
	Method   string
	Path     string
	Summary  string
	Body     interface{}   // Request body schema (if applicable)
	Query    []ParamSchema // Query parameters
	PathVars []ParamSchema // Path parameters
	Response interface{}   // Response schema
}

// ParamSchema represents a query or path parameter in OpenAPI.
type ParamSchema struct {
	Name        string
	Description string
	Required    bool
	Type        string
}

// GenerateOpenAPISpec generates the OpenAPI documentation.
func GenerateOpenAPISpec(routes []Route) OpenAPI {
	spec := OpenAPI{
		Openapi: "3.0.0",
		Info: map[string]string{
			"title":   "API Documentation",
			"version": "1.0.0",
		},
		Paths: make(map[string]interface{}),
	}

	for _, route := range routes {
		if _, exists := spec.Paths[route.Path]; !exists {
			spec.Paths[route.Path] = make(map[string]interface{})
		}

		operation := map[string]interface{}{
			"summary": route.Summary,
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Success",
				},
			},
		}
		if route.Response != nil {
			operation["responses"].(map[string]interface{})["200"].(map[string]interface{})["content"] = map[string]interface{}{
				"application/json": map[string]interface{}{"schema": schemaFor(route.Response)},
			}
		}

		// Add query parameters
		if len(route.Query) > 0 {
			params := []map[string]interface{}{}
			for _, param := range route.Query {
				params = append(params, map[string]interface{}{
					"name":        param.Name,
					"in":          "query",
					"required":    param.Required,
					"schema":      map[string]string{"type": param.Type},
					"description": param.Description,
				})
			}
			operation["parameters"] = params
		}

		// Add path parameters
		if len(route.PathVars) > 0 {
			if operation["parameters"] == nil {
				operation["parameters"] = []map[string]interface{}{}
			}
			for _, param := range route.PathVars {
				operation["parameters"] = append(operation["parameters"].([]map[string]interface{}), map[string]interface{}{
					"name":        param.Name,
					"in":          "path",
					"required":    true,
					"schema":      map[string]string{"type": param.Type},
					"description": param.Description,
				})
			}
		}

		// Add request body if applicable
		if route.Body != nil {
			operation["requestBody"] = map[string]interface{}{
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": schemaFor(route.Body),
					},
				},
			}
		}

		spec.Paths[route.Path].(map[string]interface{})[strings.ToLower(route.Method)] = operation
	}

	return spec
}

// processResponseSchema ensures arrays are correctly wrapped in OpenAPI.
func schemaFor(value interface{}) interface{} {
	if value == nil {
		return nil
	}
	if schema, ok := value.(map[string]interface{}); ok {
		return schema
	}
	return schemaForType(reflect.TypeOf(value), make(map[reflect.Type]bool))
}

func schemaForType(t reflect.Type, visiting map[reflect.Type]bool) map[string]interface{} {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == reflect.TypeOf(time.Time{}) {
		return map[string]interface{}{"type": "string", "format": "date-time"}
	}

	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		return map[string]interface{}{
			"type":  "array",
			"items": schemaForType(t.Elem(), visiting),
		}
	case reflect.Struct:
		if visiting[t] {
			return map[string]interface{}{"type": "object"}
		}
		visiting[t] = true
		defer delete(visiting, t)

		properties := make(map[string]interface{})
		required := make([]string, 0)
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if !field.IsExported() {
				continue
			}
			name, options := jsonField(field)
			if name == "-" {
				continue
			}
			properties[name] = schemaForType(field.Type, visiting)
			if !options["omitempty"] && field.Type.Kind() != reflect.Pointer {
				required = append(required, name)
			}
		}
		schema := map[string]interface{}{"type": "object", "properties": properties}
		if len(required) > 0 {
			schema["required"] = required
		}
		return schema
	case reflect.Map:
		return map[string]interface{}{"type": "object", "additionalProperties": schemaForType(t.Elem(), visiting)}
	case reflect.Bool:
		return map[string]interface{}{"type": "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return map[string]interface{}{"type": "integer"}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]interface{}{"type": "integer", "minimum": 0}
	case reflect.Float32, reflect.Float64:
		return map[string]interface{}{"type": "number"}
	case reflect.Interface:
		return map[string]interface{}{}
	default:
		return map[string]interface{}{"type": "string"}
	}
}

func jsonField(field reflect.StructField) (string, map[string]bool) {
	parts := strings.Split(field.Tag.Get("json"), ",")
	name := parts[0]
	if name == "" {
		name = field.Name
	}
	options := make(map[string]bool, len(parts)-1)
	for _, option := range parts[1:] {
		options[option] = true
	}
	return name, options
}

// ServeOpenAPIDocs serves the OpenAPI JSON.
type OpenAPIServer struct {
	Routes []Route
}

func (o *OpenAPIServer) ServeOpenAPI(c *gin.Context) {
	spec := GenerateOpenAPISpec(o.Routes)
	c.JSON(http.StatusOK, spec)
}

func NewOpenAPIServer() *OpenAPIServer {
	return &OpenAPIServer{}
}

func (o *OpenAPIServer) AddRoute(method, path, summary string, body interface{}, queryParams []ParamSchema, pathParams []ParamSchema, response interface{}) {
	// Automatically prevent requestBody for GET
	if strings.EqualFold(method, http.MethodGet) {
		body = nil
	}

	o.Routes = append(o.Routes, Route{
		Method:   method,
		Path:     path,
		Summary:  summary,
		Body:     body,
		Query:    queryParams,
		PathVars: pathParams,
		Response: response,
	})
}
