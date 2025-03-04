# gin-openapi

gin-openapi is a lightweight and plug-and-play OpenAPI (Swagger) documentation generator for Go applications using [gin](https://github.com/gin-gonic/gin) or [gin-multi-server](https://github.com/GEMSDEV-mx/gin-multi-server). It dynamically builds an OpenAPI 3.0 JSON specification based on registered routes and DTOs.

## Features

- **Dynamic OpenAPI generation** based on registered routes.
- **Works with `gin` or `gin-multi-server`** seamlessly.
- **Plug-and-play** – add OpenAPI documentation with minimal code.
- **Supports request and response DTOs** for better schema generation.
- **Returns JSON-based OpenAPI spec** (no embedded Swagger UI).

## Installation

```sh
go get github.com/GEMSDEV-mx/gin-openapi
```

## Usage

### Basic Example with `gin-multi-server`

```go
package main

import (
	"log"

	server "github.com/GEMSDEV-mx/gin-multi-server"
	openapi "github.com/GEMSDEV-mx/gin-openapi"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	s := server.NewServer()
	apiDocs := openapi.NewOpenAPIServer()

	// Define API handlers
	s.MountEndpoint("POST", "/api/resource", CreateResourceHandler)
	apiDocs.AddRoute("POST", "/api/resource", "Creates a new resource", ResourceRequest{}, ResourceResponse{})

	s.MountEndpoint("GET", "/api/resource", GetResourceHandler)
	apiDocs.AddRoute("GET", "/api/resource", "Retrieves a resource", nil, ResourceResponse{})

	// Serve OpenAPI spec
	s.MountEndpoint("GET", "/openapi.json", apiDocs.ServeOpenAPIHandler)

	s.Serve("8080")
}
```

### Example with `gin`

```go
package main

import (
	"github.com/gin-gonic/gin"
	openapi "github.com/GEMSDEV-mx/gin-openapi"
)

func main() {
	r := gin.Default()
	apiDocs := openapi.NewOpenAPIServer()

	r.POST("/api/resource", func(c *gin.Context) {
		// Handle request
	})
	apiDocs.AddRoute("POST", "/api/resource", "Creates a new resource", ResourceRequest{}, ResourceResponse{})

	r.GET("/api/resource", func(c *gin.Context) {
		// Handle request
	})
	apiDocs.AddRoute("GET", "/api/resource", "Retrieves a resource", nil, ResourceResponse{})

	// Serve OpenAPI JSON
	r.GET("/openapi.json", apiDocs.ServeOpenAPI)

	r.Run(":8080")
}
```

## OpenAPI Output Example

When requesting `GET /openapi.json`, you'll receive:

```json
{
  "openapi": "3.0.0",
  "info": {
    "title": "API Documentation",
    "version": "1.0.0"
  },
  "paths": {
    "/api/resource": {
      "POST": {
        "summary": "Creates a new resource",
        "requestBody": {
          "content": {
            "application/json": {
              "schema": { "type": "object" }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Success",
            "content": {
              "application/json": {
                "schema": { "type": "object" }
              }
            }
          }
        }
      }
    }
  }
}
```

## License

MIT
