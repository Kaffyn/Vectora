export interface MCPTool {
  name: string;
  description: string;
  inputSchema: Record<string, unknown>;
}

export interface MCPRequest {
  jsonrpc: "2.0";
  id: number;
  method: string;
  params?: Record<string, unknown>;
}

export interface MCPResponse {
  jsonrpc: "2.0";
  id: number;
  result?: unknown;
  error?: { code: number; message: string };
}

export interface ToolCallResult {
  content: Array<{ type: "text"; text: string }>;
  isError?: boolean;
}

export interface SearchResult {
  filename: string;
  content: string;
  relevance: number;
}

export interface AnalysisResult {
  patterns: string[];
  issues: Array<{
    severity: "high" | "medium" | "low";
    description: string;
    location?: string;
  }>;
  summary: string;
}
