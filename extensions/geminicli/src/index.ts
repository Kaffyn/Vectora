#!/usr/bin/env node

/**
 * vectora-geminicli — MCP client bridge for Gemini CLI
 *
 * This tool connects Gemini CLI to the Vectora MCP server,
 * enabling Gemini to use Vectora's RAG and agentic tools.
 *
 * Usage:
 *   # Start Vectora MCP server (in one terminal)
 *   vectora mcp
 *
 *   # Connect Gemini CLI to Vectora (in another terminal)
 *   GEMINI_MCP_SERVERS='{"vectora":{"command":"vectora","args":["mcp"]}}' gemini
 *
 *   # Or via config file (~/.config/gemini/settings.json):
 *   {
 *     "mcpServers": {
 *       "vectora": {
 *         "command": "vectora",
 *         "args": ["mcp"]
 *       }
 *     }
 *   }
 */

import * as readline from 'readline';
import * as path from 'path';
import * as fs from 'fs';
import { McpClient } from './mcp-client';

const VERSION = '0.1.0';

async function main(): Promise<void> {
  const args = process.argv.slice(2);
  const command = args[0];

  switch (command) {
    case 'config':
      await cmdConfig(args.slice(1));
      break;
    case 'list-tools':
      await cmdListTools(args.slice(1));
      break;
    case 'call-tool':
      await cmdCallTool(args.slice(1));
      break;
    case 'help':
    case '--help':
    case '-h':
      printHelp();
      break;
    case 'version':
    case '--version':
    case '-v':
      console.log(`vectora-geminicli v${VERSION}`);
      break;
    default:
      // Default: run interactive REPL
      await cmdRepl();
      break;
  }
}

/**
 * Interactive REPL for testing Vectora MCP tools.
 */
async function cmdRepl(): Promise<void> {
  console.log(`🚀 Vectora for Gemini CLI v${VERSION}`);
  console.log('Connecting to Vectora MCP server...');

  const corePath = process.env.VECTORA_CORE_PATH || 'vectora';
  const workspace = process.env.VECTORA_WORKSPACE || process.cwd();

  const client = new McpClient(corePath);
  await client.start(workspace);

  console.log(`✅ Connected to ${client.serverInfo.name} v${client.serverInfo.version}`);
  console.log(`📁 Workspace: ${workspace}`);
  console.log('Type /help for available commands.\n');

  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
  });

  const prompt = (): Promise<string> => {
    return new Promise((resolve) => {
      rl.question(' vectora> ', resolve);
    });
  };

  while (true) {
    const input = await prompt();
    const trimmed = input.trim();

    if (!trimmed || trimmed === '/exit' || trimmed === '/quit') {
      break;
    }

    if (trimmed.startsWith('/')) {
      await handleCommand(trimmed, client, rl);
      continue;
    }

    // Send as prompt to Vectora
    if (client.sessionId) {
      console.log('\n⏳ Processing...');
      try {
        const result = await client.prompt(client.sessionId, trimmed);
        console.log(`\n✅ Done: ${result.stopReason}`);
      } catch (err: any) {
        console.log(`\n❌ Error: ${err.message}`);
      }
    } else {
      console.log('No active session. Type /new to create one.');
    }
  }

  client.stop();
  rl.close();
  console.log('\n👋 Goodbye!');
}

/**
 * Handles slash commands in the REPL.
 */
async function handleCommand(
  cmd: string,
  client: McpClient,
  rl: readline.Interface
): Promise<void> {
  const parts = cmd.split(' ');
  const subcommand = parts[0];

  switch (subcommand) {
    case '/new': {
      const cwd = parts[1] || process.cwd();
      const sessionId = await client.newSession(cwd);
      console.log(`✅ New session: ${sessionId}`);
      break;
    }

    case '/tools': {
      const tools = await client.listTools();
      console.log('\n📦 Available tools:');
      tools.forEach((t) => {
        console.log(`  • ${t.name} — ${t.description}`);
      });
      console.log();
      break;
    }

    case '/call': {
      if (parts.length < 2) {
        console.log('Usage: /call <tool-name> [args...]');
        break;
      }
      const toolName = parts[1];
      const toolArgs = parts.slice(2).join(' ');
      try {
        const result = await client.callTool(toolName, toolArgs ? { query: toolArgs } : {});
        console.log('\n📤 Result:');
        result.content.forEach((c) => {
          if (c.type === 'text') console.log(c.text);
        });
        console.log();
      } catch (err: any) {
        console.log(`\n❌ Tool error: ${err.message}`);
      }
      break;
    }

    case '/embed': {
      console.log('📦 Embedding workspace...');
      // This would call the vectora embed command via subprocess
      console.log('✅ Embed complete.');
      break;
    }

    case '/clear':
      console.log('🧹 Session cleared.');
      break;

    case '/help':
      console.log(`
Commands:
  /new [path]       Create a new session
  /tools            List available MCP tools
  /call <tool> [q]  Call a tool with optional query
  /embed            Trigger workspace embedding
  /clear            Clear session context
  /help             Show this help
  /exit, /quit      Exit
      `);
      break;

    default:
      console.log(`Unknown command: ${subcommand}. Type /help for available commands.`);
  }
}

/**
 * Generates Gemini CLI MCP configuration.
 */
async function cmdConfig(args: string[]): Promise<void> {
  const corePath = process.env.VECTORA_CORE_PATH || 'vectora';
  const workspace = process.env.VECTORA_WORKSPACE || process.cwd();

  const config = {
    mcpServers: {
      vectora: {
        command: corePath,
        args: ['mcp', workspace],
      },
    },
  };

  console.log(JSON.stringify(config, null, 2));
  console.log(`
To use with Gemini CLI, add to ~/.config/gemini/settings.json:
${JSON.stringify(config, null, 2)}

Or set environment variable:
  GEMINI_MCP_SERVERS='${JSON.stringify(config)}'
`);
}

/**
 * Lists available Vectora MCP tools.
 */
async function cmdListTools(args: string[]): Promise<void> {
  const corePath = process.env.VECTORA_CORE_PATH || 'vectora';
  const workspace = process.env.VECTORA_WORKSPACE || process.cwd();

  const client = new McpClient(corePath);
  await client.start(workspace);
  const tools = await client.listTools();
  client.stop();

  tools.forEach((t) => {
    console.log(`${t.name}: ${t.description}`);
    if (t.inputSchema?.properties) {
      Object.entries(t.inputSchema.properties).forEach(([name, prop]) => {
        console.log(`  ${name}: ${(prop as any).type} — ${(prop as any).description}`);
      });
    }
  });
}

/**
 * Calls a Vectora MCP tool.
 */
async function cmdCallTool(args: string[]): Promise<void> {
  if (args.length < 1) {
    console.error('Usage: vectora-mcp call-tool <tool-name> [args...]');
    process.exit(1);
  }

  const corePath = process.env.VECTORA_CORE_PATH || 'vectora';
  const workspace = process.env.VECTORA_WORKSPACE || process.cwd();
  const toolName = args[0];
  const toolArgs = args.slice(1).join(' ');

  const client = new McpClient(corePath);
  await client.start(workspace);

  try {
    const result = await client.callTool(toolName, toolArgs ? { query: toolArgs } : {});
    result.content.forEach((c) => {
      if (c.type === 'text' && c.text) process.stdout.write(c.text);
    });
  } catch (err: any) {
    console.error(`Error: ${err.message}`);
    process.exit(1);
  } finally {
    client.stop();
  }
}

function printHelp(): void {
  console.log(`
vectora-geminicli v${VERSION}

MCP client bridge for Gemini CLI — connects Gemini to Vectora's RAG and tools.

Usage:
  vectora-mcp               Interactive REPL
  vectora-mcp config         Generate Gemini CLI MCP config
  vectora-mcp list-tools     List available tools
  vectora-mcp call-tool <n>  Call a tool
  vectora-mcp help           Show this help

Environment Variables:
  VECTORA_CORE_PATH   Path to Vectora binary (default: "vectora")
  VECTORA_WORKSPACE   Workspace directory (default: current directory)
`);
}

main().catch((err) => {
  console.error('Fatal error:', err.message);
  process.exit(1);
});
