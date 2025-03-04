package openapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// OpenAPI defines the OpenAPI schema
// This is a minimal example; expand as needed.
type OpenAPI struct {
	Openapi string                 `json:"openapi"`
	Info    map[string]string      `json:"info"`
	Paths   map[string]interface{} `json:"paths"`
}

// GenerateOpenAPISpec generates a basic OpenAPI spec
type Route struct {
	Method  string
	Path    string
	Summary string
	Request interface{}
	Response interface{}
}

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
		spec.Paths[route.Path].(map[string]interface{})[route.Method] = map[string]interface{}{
			"summary": route.Summary,
			"requestBody": map[string]interface{}{
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": route.Request,
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Success",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": route.Response,
						},
					},
				},
			},
		}
	}

	return spec
}

// ServeOpenAPIDocs serves the OpenAPI JSON
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

func (o *OpenAPIServer) AddRoute(method, path, summary string, request interface{}, response interface{}) {
	o.Routes = append(o.Routes, Route{Method: method, Path: path, Summary: summary, Request: request, Response: response})
}
