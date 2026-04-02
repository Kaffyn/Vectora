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
// Operado via Pipes de StdIn/StdOut brutos para IDEs de terceiros plugar nosso "Cérebro Local".

func StartMCPServer(ctx context.Context) error {
	scanner := bufio.NewScanner(os.Stdin)
	
	// Limite alto devido a JSON-RPC payloads
	buf := make([]byte, 4*1024*1024)
	scanner.Buffer(buf, len(buf))

	for scanner.Scan() {
		line := scanner.Bytes()
		var req map[string]any
		
		if err := json.Unmarshal(line, &req); err != nil {
			continue // Silence. O Pipe pode ter sujeiras
		}

		// Checagem JSON-RPC 2.0 Base
		method, ok := req["method"].(string)
		id, _ := req["id"]

		if ok && method == "tools/list" {
			// Expõe a Tool Canônica "ask_vectora" nativamente via Cursor IDE!
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
			fmt.Println(string(b)) // Despeja no Stdout pra IDE ler e acoplar a Tool
		}
		
		if ok && method == "tools/call" {
			params, _ := req["params"].(map[string]any)
			name, _ := params["name"].(string)

			if name == "ask_vectora" {
				argsMap, _ := params["arguments"].(map[string]any)
				query, _ := argsMap["query"].(string)
				wsID, _ := argsMap["workspace_id"].(string)

				// Dispara Cliente IPC pra Conversar com o Nosso Daemon em Background
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
