export type Role = "user" | "agent";

export interface Message {
  id: string;
  role: Role;
  content: string;
  timestamp: number;
  tools?: ToolCall[];
}

export interface ToolCall {
  id: string;
  title: string;
  status: "in_progress" | "completed" | "failed";
}
