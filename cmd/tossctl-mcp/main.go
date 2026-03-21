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
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id,omitempty"`
	Result  any    `json:"result,omitempty"`
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
		writeError(nil, -32603, fmt.Sprintf("session load failed: %v", err))
		os.Exit(1)
	}

	c := client.New(client.Config{
		Session:       sess,
		TradingPolicy: config.Trading{}, // read-only, all trading disabled
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
			writeError(nil, -32700, "parse error")
			continue
		}

		s.handle(req)
	}
}

func (s *server) handle(req jsonrpcRequest) {
	switch req.Method {
	case "initialize":
		writeResult(req.ID, map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
			"serverInfo": map[string]any{
				"name":    "tossctl-mcp",
				"version": version.Current().Version,
			},
		})
	case "notifications/initialized":
		// no response needed for notifications
	case "tools/list":
		writeResult(req.ID, map[string]any{
			"tools": []map[string]any{
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
			},
		})
	case "tools/call":
		s.handleToolCall(req)
	default:
		writeError(req.ID, -32601, fmt.Sprintf("method not found: %s", req.Method))
	}
}

func (s *server) handleToolCall(req jsonrpcRequest) {
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		writeError(req.ID, -32602, "invalid params")
		return
	}

	ctx := context.Background()
	var result any
	var callErr error

	switch params.Name {
	case "get_portfolio_positions":
		result, callErr = s.client.ListPositions(ctx)
	case "get_account_summary":
		result, callErr = s.client.GetAccountSummary(ctx)
	case "get_quote":
		var args struct {
			Symbol string `json:"symbol"`
		}
		if err := json.Unmarshal(params.Arguments, &args); err != nil || args.Symbol == "" {
			writeToolError(req.ID, "symbol is required")
			return
		}
		result, callErr = s.client.GetQuote(ctx, args.Symbol)
	case "list_pending_orders":
		result, callErr = s.client.ListPendingOrders(ctx)
	case "list_completed_orders":
		var args struct {
			Market string `json:"market"`
		}
		_ = json.Unmarshal(params.Arguments, &args)
		if args.Market == "" {
			args.Market = "all"
		}
		result, callErr = s.client.ListCompletedOrders(ctx, args.Market)
	case "list_watchlist":
		result, callErr = s.client.ListWatchlist(ctx)
	default:
		writeToolError(req.ID, fmt.Sprintf("unknown tool: %s", params.Name))
		return
	}

	if callErr != nil {
		writeToolError(req.ID, callErr.Error())
		return
	}

	data, _ := json.Marshal(result)
	writeResult(req.ID, map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": string(data)},
		},
	})
}

func writeResult(id any, result any) {
	resp := jsonrpcResponse{JSONRPC: "2.0", ID: id, Result: result}
	data, _ := json.Marshal(resp)
	fmt.Fprintf(os.Stdout, "%s\n", data)
}

func writeError(id any, code int, message string) {
	resp := jsonrpcResponse{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: code, Message: message}}
	data, _ := json.Marshal(resp)
	fmt.Fprintf(os.Stdout, "%s\n", data)
}

func writeToolError(id any, message string) {
	writeResult(id, map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": message},
		},
		"isError": true,
	})
}
