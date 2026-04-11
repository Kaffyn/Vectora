import * as vscode from 'vscode';
import * as fs from 'fs';
import * as path from 'path';
import * as os from 'os';
import * as https from 'https';
import * as cp from 'child_process';
// Directory where the Vectora binary will be installed
const VECTORA_BIN_DIR = path.join(os.homedir(), '.vectora', 'bin');

export class BinaryManager {
  private readonly installDir: string;
  private readonly binaryName: string;
  private readonly binaryPath: string;

  constructor() {
    this.installDir = VECTORA_BIN_DIR;

    // Detect binary name based on OS
    const isWin = os.platform() === 'win32';
    this.binaryName = isWin ? 'vectora.exe' : 'vectora';
    this.binaryPath = path.join(this.installDir, this.binaryName);
  }

  /**
   * Ensures the Vectora binary exists. Downloads it if necessary.
   * Returns the absolute path to the binary.
   */
  async ensureBinary(): Promise<string> {
    // 1. Check user-configured custom path
    const configPath = vscode.workspace.getConfiguration('vectora').get<string>('corePath');
    if (configPath && configPath !== 'vectora' && configPath !== this.binaryName) {
      if (await this.fileExists(configPath)) {
        return configPath;
      }
      throw new Error(`Vectora binary not found at configured path: ${configPath}`);
    }

    // 2. Check our managed directory (~/.vectora/bin)
    if (await this.fileExists(this.binaryPath)) {
      return this.binaryPath;
    }

     // 3. Check AppData\Local\Vectora on Windows (Standard install dir)
     if (os.platform() === 'win32') {
       const localAppData = process.env.LOCALAPPDATA || path.join(os.homedir(), 'AppData', 'Local');
       const installPath = path.join(localAppData, 'Vectora', this.binaryName);
       if (await this.fileExists(installPath)) {
         return installPath;
       }
     }

    // 4. Check system PATH
    const pathBinary = await this.findInPath();
    if (pathBinary) {
      return pathBinary;
    }

    // 5. Not found anywhere — offer to download
    const action = await vscode.window.showErrorMessage(
      'Vectora Core binary not found. Download it automatically?',
      'Download',
      'Cancel'
    );

    if (action === 'Download') {
      await this.downloadBinary();
      return this.binaryPath;
    }

    throw new Error('Vectora Core binary not found. Please install it or configure vectora.corePath.');
  }

  /**
   * Downloads the latest Vectora binary from GitHub Releases.
   */
  private async downloadBinary(): Promise<void> {
    if (!fs.existsSync(this.installDir)) {
      fs.mkdirSync(this.installDir, { recursive: true });
    }

    const platform = os.platform();
    const arch = os.arch();
    const version = 'v0.1.0';

    // Map Node arch to Go arch
    const goArch = arch === 'x64' ? 'amd64' : arch === 'arm64' ? 'arm64' : arch;

    let fileName: string;
    if (platform === 'win32') {
      fileName = `vectora-windows-${goArch}.exe`;
    } else if (platform === 'darwin') {
      fileName = `vectora-darwin-${goArch}`;
    } else {
      fileName = `vectora-linux-${goArch}`;
    }

    const downloadUrl = `https://github.com/Kaffyn/Vectora/releases/download/${version}/${fileName}`;

    return vscode.window.withProgress({
      location: vscode.ProgressLocation.Notification,
      title: 'Downloading Vectora Core...',
      cancellable: false,
    }, async () => {
      try {
        const file = fs.createWriteStream(this.binaryPath);
        await this.downloadFile(downloadUrl, file);

        // Set executable permission on Unix
        if (platform !== 'win32') {
          fs.chmodSync(this.binaryPath, 0o755);
        }

        vscode.window.showInformationMessage('Vectora Core downloaded successfully!');
      } catch (err: any) {
        // Clean up partial download
        try { fs.unlinkSync(this.binaryPath); } catch { /* ignore */ }
        throw new Error(`Download failed: ${err.message}`);
      }
    });
  }

  /**
   * Downloads a file from a URL to a write stream.
   */
  private downloadFile(url: string, dest: fs.WriteStream): Promise<void> {
    return new Promise((resolve, reject) => {
      https.get(url, (response) => {
        // Follow redirects
        if (response.statusCode === 302 || response.statusCode === 301) {
          this.downloadFile(response.headers.location!, dest).then(resolve).catch(reject);
          return;
        }

        if (response.statusCode !== 200) {
          reject(new Error(`HTTP ${response.statusCode}`));
          return;
        }

        response.pipe(dest);
        dest.on('finish', () => {
          dest.close();
          resolve();
        });
      }).on('error', (err) => {
        try { fs.unlinkSync(this.binaryPath); } catch { /* ignore */ }
        reject(err);
      });
    });
  }

  /**
   * Checks if a file exists at the given path.
   */
  private async fileExists(filePath: string): Promise<boolean> {
    try {
      await fs.promises.access(filePath, fs.constants.F_OK);
      return true;
    } catch {
      return false;
    }
  }

  /**
   * Searches for 'vectora' in the system PATH.
   */
  private async findInPath(): Promise<string | null> {
    const cmd = os.platform() === 'win32' ? 'where vectora' : 'which vectora';
    try {
      return new Promise((resolve) => {
        cp.exec(cmd, (error: Error | null, stdout: string) => {
          if (error || !stdout) {
            resolve(null);
          } else {
            resolve(stdout.trim().split('\n')[0]);
          }
        });
      });
    } catch {
      return null;
    }
  }
}
