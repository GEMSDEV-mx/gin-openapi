package openapi

import (
	"net/http"
	"reflect"

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

		// Ensure array responses are correctly wrapped
		responseSchema := processResponseSchema(route.Response)

		operation := map[string]interface{}{
			"summary": route.Summary,
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Success",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": responseSchema,
						},
					},
				},
			},
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
					"required":    param.Required,
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
						"schema": route.Body,
					},
				},
			}
		}

		spec.Paths[route.Path].(map[string]interface{})[route.Method] = operation
	}

	return spec
}

// processResponseSchema ensures arrays are correctly wrapped in OpenAPI.
func processResponseSchema(response interface{}) interface{} {
	if response == nil {
		return nil
	}

	// Detect if response is a slice (array)
	responseType := reflect.TypeOf(response)
	if responseType.Kind() == reflect.Slice {
		return map[string]interface{}{
			"type":  "array",
			"items": response,
		}
	}

	// Return as-is for single objects
	return response
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
	if method == "GET" {
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
