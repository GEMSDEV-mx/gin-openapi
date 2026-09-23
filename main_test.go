package openapi

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

type testRequest struct {
	Name      string       `json:"name"`
	Nickname  string       `json:"nickname,omitempty"`
	CreatedAt time.Time    `json:"createdAt"`
	Parent    *testRequest `json:"parent,omitempty"`
}

func TestGenerateOpenAPISpecProducesSchemasAndValidOperationKeys(t *testing.T) {
	routes := []Route{{
		Method:   "POST",
		Path:     "/items/{id}",
		Summary:  "Create an item",
		Body:     testRequest{},
		PathVars: []ParamSchema{{Name: "id", Type: "string"}},
		Response: []testRequest{},
	}}

	spec := GenerateOpenAPISpec(routes)
	path := spec.Paths["/items/{id}"].(map[string]interface{})
	operation, ok := path["post"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected lowercase post operation, got %#v", path)
	}
	params := operation["parameters"].([]map[string]interface{})
	if required, _ := params[0]["required"].(bool); !required {
		t.Fatal("OpenAPI path parameters must always be required")
	}

	bodySchema := operation["requestBody"].(map[string]interface{})["content"].(map[string]interface{})["application/json"].(map[string]interface{})["schema"].(map[string]interface{})
	properties := bodySchema["properties"].(map[string]interface{})
	if properties["createdAt"].(map[string]interface{})["format"] != "date-time" {
		t.Fatalf("expected date-time schema, got %#v", properties["createdAt"])
	}

	responseSchema := operation["responses"].(map[string]interface{})["200"].(map[string]interface{})["content"].(map[string]interface{})["application/json"].(map[string]interface{})["schema"].(map[string]interface{})
	if responseSchema["type"] != "array" {
		t.Fatalf("expected array response schema, got %#v", responseSchema)
	}
	if responseSchema["items"].(map[string]interface{})["type"] != "object" {
		t.Fatalf("expected object item schema, got %#v", responseSchema["items"])
	}

	if _, err := json.Marshal(spec); err != nil {
		t.Fatalf("spec must be JSON serializable: %v", err)
	}
}

func TestAddRouteTreatsGetCaseInsensitively(t *testing.T) {
	server := NewOpenAPIServer()
	server.AddRoute("get", "/items", "List", testRequest{}, nil, nil, []testRequest{})
	if server.Routes[0].Body != nil {
		t.Fatal("GET routes must not expose request bodies")
	}
}

func TestServeOpenAPI(t *testing.T) {
	server := NewOpenAPIServer()
	server.AddRoute(http.MethodGet, "/health", "Health", nil, nil, nil, map[string]interface{}{"type": "string"})
	if got := GenerateOpenAPISpec(server.Routes).Paths["/health"].(map[string]interface{})["get"]; got == nil {
		t.Fatal("expected GET operation")
	}
}
