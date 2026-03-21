// tossctl-mcp is a read-only MCP (Model Context Protocol) server for tossinvest-cli.
// It exposes portfolio, quote, order, and account data to AI agents via JSON-RPC over stdin/stdout.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/junghoonkye/tossinvest-cli/internal/client"
	"github.com/junghoonkye/tossinvest-cli/internal/config"
	"github.com/junghoonkye/tossinvest-cli/internal/session"
	"github.com/junghoonkye/tossinvest-cli/internal/version"
)

type jsonrpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonrpcResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id,omitempty"`
	Result  any       `json:"result,omitempty"`
	Error   *rpcError `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type server struct {
	client *client.Client
}

func main() {
	configDir, _ := os.UserConfigDir()
	sessionPath := filepath.Join(configDir, "tossctl", "session.json")

	store := session.NewFileStore(sessionPath)
	sess, err := store.Load(context.Background())
	if err != nil {
		resp := makeError(nil, -32603, fmt.Sprintf("session load failed: %v", err))
		writeJSON(resp)
		os.Exit(1)
	}

	c := client.New(client.Config{
		Session:       sess,
		TradingPolicy: config.Trading{},
	})

	s := &server{client: c}
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req jsonrpcRequest
		if err := json.Unmarshal(line, &req); err != nil {
			writeJSON(makeError(nil, -32700, "parse error"))
			continue
		}

		resp := handleMethod(s.client, req)
		if resp != nil {
			writeJSON(resp)
		}
	}
}

// handleMethod dispatches a JSON-RPC request and returns a response.
// Returns nil for notifications (no response needed).
func handleMethod(c *client.Client, req jsonrpcRequest) *jsonrpcResponse {
	switch req.Method {
	case "initialize":
		return makeResult(req.ID, buildInitializeResponse())
	case "notifications/initialized":
		return nil
	case "tools/list":
		return makeResult(req.ID, map[string]any{"tools": buildToolsList()})
	case "tools/call":
		return handleToolCall(c, req)
	default:
		return makeError(req.ID, -32601, fmt.Sprintf("method not found: %s", req.Method))
	}
}

func buildInitializeResponse() map[string]any {
	return map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{"tools": map[string]any{}},
		"serverInfo": map[string]any{
			"name":    "tossctl-mcp",
			"version": version.Current().Version,
		},
	}
}

func buildToolsList() []map[string]any {
	return []map[string]any{
		{
			"name":        "get_portfolio_positions",
			"description": "현재 보유 포지션(주식, ETF 등) 목록과 수익률을 조회합니다",
			"inputSchema": map[string]any{"type": "object", "properties": map[string]any{}},
		},
		{
			"name":        "get_account_summary",
			"description": "계좌 총자산, 수익률, 주문가능금액 등 요약 정보를 조회합니다",
			"inputSchema": map[string]any{"type": "object", "properties": map[string]any{}},
		},
		{
			"name":        "get_quote",
			"description": "특정 종목의 현재 시세를 조회합니다 (US/KR 모두 지원)",
			"inputSchema": map[string]any{
				"type":     "object",
				"required": []string{"symbol"},
				"properties": map[string]any{
					"symbol": map[string]any{
						"type":        "string",
						"description": "종목 심볼 (예: TSLL, 005930, AAPL)",
					},
				},
			},
		},
		{
			"name":        "list_pending_orders",
			"description": "현재 미체결(대기) 주문 목록을 조회합니다",
			"inputSchema": map[string]any{"type": "object", "properties": map[string]any{}},
		},
		{
			"name":        "list_completed_orders",
			"description": "이번 달 체결 완료된 주문 내역을 조회합니다",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"market": map[string]any{
						"type":        "string",
						"description": "시장 필터: all, us, kr (기본값: all)",
						"enum":        []string{"all", "us", "kr"},
					},
				},
			},
		},
		{
			"name":        "list_watchlist",
			"description": "관심 종목 목록을 조회합니다",
			"inputSchema": map[string]any{"type": "object", "properties": map[string]any{}},
		},
	}
}

func handleToolCall(c *client.Client, req jsonrpcRequest) *jsonrpcResponse {
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return makeError(req.ID, -32602, "invalid params")
	}

	if c == nil {
		return makeToolError(req.ID, "no active session")
	}

	ctx := context.Background()
	var result any
	var callErr error

	switch params.Name {
	case "get_portfolio_positions":
		result, callErr = c.ListPositions(ctx)
	case "get_account_summary":
		result, callErr = c.GetAccountSummary(ctx)
	case "get_quote":
		var args struct {
			Symbol string `json:"symbol"`
		}
		if err := json.Unmarshal(params.Arguments, &args); err != nil || args.Symbol == "" {
			return makeToolError(req.ID, "symbol is required")
		}
		result, callErr = c.GetQuote(ctx, args.Symbol)
	case "list_pending_orders":
		result, callErr = c.ListPendingOrders(ctx)
	case "list_completed_orders":
		var args struct {
			Market string `json:"market"`
		}
		_ = json.Unmarshal(params.Arguments, &args)
		if args.Market == "" {
			args.Market = "all"
		}
		result, callErr = c.ListCompletedOrders(ctx, args.Market)
	case "list_watchlist":
		result, callErr = c.ListWatchlist(ctx)
	default:
		return makeToolError(req.ID, fmt.Sprintf("unknown tool: %s", params.Name))
	}

	if callErr != nil {
		return makeToolError(req.ID, callErr.Error())
	}

	data, _ := json.Marshal(result)
	return makeResult(req.ID, map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": string(data)},
		},
	})
}

func makeResult(id any, result any) *jsonrpcResponse {
	return &jsonrpcResponse{JSONRPC: "2.0", ID: id, Result: result}
}

func makeError(id any, code int, message string) *jsonrpcResponse {
	return &jsonrpcResponse{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: code, Message: message}}
}

func makeToolError(id any, message string) *jsonrpcResponse {
	return makeResult(id, map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": message},
		},
		"isError": true,
	})
}

func writeJSON(resp *jsonrpcResponse) {
	data, _ := json.Marshal(resp)
	fmt.Fprintf(os.Stdout, "%s\n", data)
}
