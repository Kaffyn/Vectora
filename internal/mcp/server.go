package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Kaffyn/vectora/internal/core"
	"github.com/Kaffyn/vectora/internal/ipc"
)

// MCP Server: Model Context Protocol.
// Operated via raw StdIn/StdOut Pipes for third-party IDEs to plug into our "Local Brain".

func StartMCPServer(ctx context.Context) error {
	scanner := bufio.NewScanner(os.Stdin)
	
	// High limit due to JSON-RPC payloads
	buf := make([]byte, 4*1024*1024)
	scanner.Buffer(buf, len(buf))

	for scanner.Scan() {
		line := scanner.Bytes()
		var req map[string]any
		
		if err := json.Unmarshal(line, &req); err != nil {
			continue // Silence. The Pipe might have noise.
		}

		// Checagem JSON-RPC 2.0 Base
		method, ok := req["method"].(string)
		id, _ := req["id"]

		if ok && method == "tools/list" {
			// Exposes the canonical "ask_vectora" tool natively via Cursor IDE!
			resp := map[string]any{
				"jsonrpc": "2.0",
				"id":      id,
				"result": map[string]any{
					"tools": []map[string]any{
						{
							"name":        "ask_vectora",
							"description": "Realiza RAG nos seus diretórios privados via Chromem-Go.",
							"inputSchema": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"query":        map[string]string{"type": "string"},
									"workspace_id": map[string]string{"type": "string"},
								},
								"required": []string{"query", "workspace_id"},
							},
						},
					},
				},
			}
			
			b, _ := json.Marshal(resp)
			fmt.Println(string(b)) // Dumps to Stdout for the IDE to read and attach the Tool.
		}
		
		if ok && method == "tools/call" {
			params, _ := req["params"].(map[string]any)
			name, _ := params["name"].(string)

			if name == "ask_vectora" {
				argsMap, _ := params["arguments"].(map[string]any)
				query, _ := argsMap["query"].(string)
				wsID, _ := argsMap["workspace_id"].(string)

				// Triggers IPC Client to talk to our Background Daemon.
				ipcClient, _ := ipc.NewClient()
				if ipcClient != nil && ipcClient.Connect() == nil {
					var queryResp core.QueryResponse
					reqStruct := core.QueryRequest{Query: query, WorkspaceID: wsID}
					ipcClient.Send(ctx, "workspace.query", reqStruct, &queryResp)

					// Responde JSON RPC V2 pro VSCode
					escAns, _ := json.Marshal(queryResp.Answer)
					resp := fmt.Sprintf(`{"jsonrpc": "2.0", "id": "%v", "result": {"content": [{"type": "text", "text": %s}]}}`, id, string(escAns))
					fmt.Println(resp)
					ipcClient.Close()
				}
			}
		}
	}

	return nil
}
