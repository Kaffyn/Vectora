#!/usr/bin/env node
/**
 * Vectora postinstall script
 * Detecta o SO, baixa o binário correto do GitHub Releases e instala em ~/.vectora/bin/
 */

const { execSync, spawnSync } = require("child_process");
const { createWriteStream, mkdirSync, chmodSync, existsSync } = require("fs");
const { join, dirname } = require("path");
const { homedir, platform, arch } = require("os");
const https = require("https");

const REPO = "Kaffyn/Vectora";
const BINARY_NAME = "vectora";
const INSTALL_DIR = join(homedir(), ".vectora", "bin");
const VERSION = require("../package.json").vectoraVersion || "latest";

// Mapear plataforma Node → nome do binário no GitHub Releases
function getBinaryName() {
  const p = platform(); // win32, linux, darwin
  const a = arch();     // x64, arm64, ia32

  const osMap = { win32: "windows", linux: "linux", darwin: "darwin" };
  const archMap = { x64: "amd64", arm64: "arm64", ia32: "386" };

  const os = osMap[p];
  const ar = archMap[a];

  if (!os || !ar) {
    throw new Error(`Plataforma não suportada: ${p}/${a}`);
  }

  const suffix = p === "win32" ? ".exe" : "";
  return `vectora-${os}-${ar}${suffix}`;
}

function getInstallPath() {
  const suffix = platform() === "win32" ? ".exe" : "";
  return join(INSTALL_DIR, `${BINARY_NAME}${suffix}`);
}

// Verifica se o vectora já está instalado e é recente
function isAlreadyInstalled() {
  const installPath = getInstallPath();
  if (existsSync(installPath)) {
    console.log(`✓ Vectora já instalado em: ${installPath}`);
    return true;
  }

  // Checa se está no PATH
  const result = spawnSync(BINARY_NAME, ["--version"], { encoding: "utf8" });
  if (result.status === 0) {
    console.log(`✓ Vectora encontrado no PATH: ${result.stdout.trim()}`);
    return true;
  }

  return false;
}

// Obtém a URL do release mais recente via GitHub API
async function getDownloadUrl(binaryName) {
  return new Promise((resolve, reject) => {
    const apiUrl = VERSION === "latest"
      ? `https://api.github.com/repos/${REPO}/releases/latest`
      : `https://api.github.com/repos/${REPO}/releases/tags/${VERSION}`;

    const options = {
      hostname: "api.github.com",
      path: new URL(apiUrl).pathname,
      headers: {
        "User-Agent": "vectora-claude-code-postinstall",
        Accept: "application/vnd.github.v3+json",
      },
    };

    https.get(options, (res) => {
      let data = "";
      res.on("data", (chunk) => (data += chunk));
      res.on("end", () => {
        try {
          const release = JSON.parse(data);
          const asset = release.assets?.find((a) => a.name === binaryName);
          if (!asset) {
            reject(new Error(`Binário "${binaryName}" não encontrado no release ${release.tag_name}`));
            return;
          }
          resolve({ url: asset.browser_download_url, version: release.tag_name });
        } catch (e) {
          reject(new Error(`Falha ao parsear resposta da API GitHub: ${e.message}`));
        }
      });
    }).on("error", reject);
  });
}

// Download com follow de redirects
function download(url, destPath) {
  return new Promise((resolve, reject) => {
    const file = createWriteStream(destPath);

    function get(url) {
      https.get(url, (res) => {
        if (res.statusCode === 301 || res.statusCode === 302) {
          get(res.headers.location);
          return;
        }
        if (res.statusCode !== 200) {
          reject(new Error(`HTTP ${res.statusCode} ao baixar ${url}`));
          return;
        }

        const total = parseInt(res.headers["content-length"] || "0");
        let downloaded = 0;

        res.on("data", (chunk) => {
          downloaded += chunk.length;
          if (total > 0) {
            const pct = Math.round((downloaded / total) * 100);
            process.stdout.write(`\r  Baixando... ${pct}% (${Math.round(downloaded / 1024)}KB)`);
          }
        });

        res.pipe(file);
        file.on("finish", () => {
          file.close();
          process.stdout.write("\n");
          resolve();
        });
      }).on("error", reject);
    }

    get(url);
  });
}

async function main() {
  console.log("\n🔧 Vectora — configurando binário...\n");

  if (isAlreadyInstalled()) {
    console.log("✓ Nenhuma ação necessária.\n");
    writeConfig();
    return;
  }

  const binaryName = getBinaryName();
  const installPath = getInstallPath();

  console.log(`→ Sistema detectado: ${platform()}/${arch()}`);
  console.log(`→ Binário necessário: ${binaryName}`);
  console.log(`→ Destino: ${installPath}\n`);

  // Criar diretório de instalação
  mkdirSync(INSTALL_DIR, { recursive: true });

  try {
    // Buscar URL do release
    process.stdout.write("→ Consultando GitHub Releases...");
    const { url, version } = await getDownloadUrl(binaryName);
    console.log(` ${version}`);
    console.log(`→ URL: ${url}\n`);

    // Download
    await download(url, installPath);

    // Tornar executável (Unix)
    if (platform() !== "win32") {
      chmodSync(installPath, 0o755);
    }

    console.log(`\n✅ Vectora instalado com sucesso em: ${installPath}`);
    console.log(`\n   Adicione ao seu PATH:\n   export PATH="$PATH:${INSTALL_DIR}"\n`);

    writeConfig(installPath);
  } catch (err) {
    console.error(`\n⚠️  Falha ao instalar automaticamente: ${err.message}`);
    console.error(`\n   Instale manualmente em: https://github.com/${REPO}/releases`);
    console.error(`   E certifique-se que 'vectora' está no PATH.\n`);
    // Não falha o npm install — só avisa
  }
}

// Gera o .mcp.json de exemplo para o usuário
function writeConfig(binaryPath) {
  const configPath = join(INSTALL_DIR, "..", "mcp-example.json");
  const binary = binaryPath || "vectora";
  const config = {
    mcpServers: {
      vectora: {
        command: binary,
        args: ["mcp", "<CAMINHO_DO_SEU_WORKSPACE>"],
        description: "Vectora MCP Server — RAG + análise de código",
      },
    },
  };

  const { writeFileSync } = require("fs");
  writeFileSync(configPath, JSON.stringify(config, null, 2));
  console.log(`\n📄 Exemplo de .mcp.json gerado em: ${configPath}`);
  console.log(`   Copie para seu projeto e ajuste o workspace.\n`);
}

main().catch((e) => {
  console.error(e);
  process.exit(0); // Não bloqueia npm install
});
