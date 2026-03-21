package main

import (
	"encoding/json"
	"testing"
)

func TestBuildInitializeResponse(t *testing.T) {
	resp := buildInitializeResponse()
	if resp["protocolVersion"] != "2024-11-05" {
		t.Fatalf("expected protocol version 2024-11-05, got %v", resp["protocolVersion"])
	}
	serverInfo := resp["serverInfo"].(map[string]any)
	if serverInfo["name"] != "tossctl-mcp" {
		t.Fatalf("expected server name tossctl-mcp, got %v", serverInfo["name"])
	}
}

func TestBuildToolsList(t *testing.T) {
	tools := buildToolsList()
	if len(tools) != 6 {
		t.Fatalf("expected 6 tools, got %d", len(tools))
	}

	names := make(map[string]bool)
	for _, tool := range tools {
		names[tool["name"].(string)] = true
	}

	expected := []string{
		"get_portfolio_positions",
		"get_account_summary",
		"get_quote",
		"list_pending_orders",
		"list_completed_orders",
		"list_watchlist",
	}
	for _, name := range expected {
		if !names[name] {
			t.Fatalf("missing tool: %s", name)
		}
	}
}

func TestHandleInitialize(t *testing.T) {
	req := jsonrpcRequest{JSONRPC: "2.0", ID: float64(1), Method: "initialize"}
	resp := handleMethod(nil, req)
	if resp == nil {
		t.Fatal("expected response")
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
}

func TestHandleNotification(t *testing.T) {
	req := jsonrpcRequest{JSONRPC: "2.0", Method: "notifications/initialized"}
	resp := handleMethod(nil, req)
	if resp != nil {
		t.Fatal("expected nil response for notification")
	}
}

func TestHandleToolsList(t *testing.T) {
	req := jsonrpcRequest{JSONRPC: "2.0", ID: float64(2), Method: "tools/list"}
	resp := handleMethod(nil, req)
	if resp == nil || resp.Error != nil {
		t.Fatal("expected success response")
	}
	result := resp.Result.(map[string]any)
	tools := result["tools"].([]map[string]any)
	if len(tools) != 6 {
		t.Fatalf("expected 6 tools, got %d", len(tools))
	}
}

func TestHandleUnknownMethod(t *testing.T) {
	req := jsonrpcRequest{JSONRPC: "2.0", ID: float64(1), Method: "unknown/method"}
	resp := handleMethod(nil, req)
	if resp.Error == nil {
		t.Fatal("expected error for unknown method")
	}
	if resp.Error.Code != -32601 {
		t.Fatalf("expected -32601, got %d", resp.Error.Code)
	}
}

func TestHandleToolCallMissingSymbol(t *testing.T) {
	params, _ := json.Marshal(map[string]any{"name": "get_quote", "arguments": map[string]any{}})
	req := jsonrpcRequest{JSONRPC: "2.0", ID: float64(1), Method: "tools/call", Params: params}
	resp := handleMethod(nil, req)
	result := resp.Result.(map[string]any)
	if result["isError"] != true {
		t.Fatal("expected isError for missing symbol")
	}
}

func TestHandleToolCallUnknownTool(t *testing.T) {
	params, _ := json.Marshal(map[string]any{"name": "nonexistent", "arguments": map[string]any{}})
	req := jsonrpcRequest{JSONRPC: "2.0", ID: float64(1), Method: "tools/call", Params: params}
	resp := handleMethod(nil, req)
	result := resp.Result.(map[string]any)
	if result["isError"] != true {
		t.Fatal("expected isError for unknown tool")
	}
}

func TestHandleToolCallNoSession(t *testing.T) {
	params, _ := json.Marshal(map[string]any{"name": "get_portfolio_positions", "arguments": map[string]any{}})
	req := jsonrpcRequest{JSONRPC: "2.0", ID: float64(1), Method: "tools/call", Params: params}
	resp := handleMethod(nil, req)
	result := resp.Result.(map[string]any)
	if result["isError"] != true {
		t.Fatal("expected isError when client is nil")
	}
}

func TestMakeResult(t *testing.T) {
	resp := makeResult(float64(1), "hello")
	if resp.JSONRPC != "2.0" {
		t.Fatalf("expected 2.0, got %s", resp.JSONRPC)
	}
	if resp.Result != "hello" {
		t.Fatalf("expected hello, got %v", resp.Result)
	}
	if resp.Error != nil {
		t.Fatal("expected no error")
	}
}

func TestMakeError(t *testing.T) {
	resp := makeError(float64(1), -32600, "bad request")
	if resp.Error == nil {
		t.Fatal("expected error")
	}
	if resp.Error.Code != -32600 {
		t.Fatalf("expected -32600, got %d", resp.Error.Code)
	}
}
