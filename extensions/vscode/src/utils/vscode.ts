import type { WebviewApi } from "vscode-webview";

declare const acquireVsCodeApi: <T = unknown>() => WebviewApi<T>;

/**
 * A utility wrapper around the acquireVsCodeApi() function.
 */
class VSCodeAPIWrapper {
    private readonly vsCodeApi: WebviewApi<any> | undefined;

    constructor() {
        if (typeof acquireVsCodeApi === "function") {
            this.vsCodeApi = acquireVsCodeApi();
        }
    }

    public postMessage(message: any) {
        if (this.vsCodeApi) {
            this.vsCodeApi.postMessage(message);
        } else {
            console.log("Mock postMessage:", message);
        }
    }

    public getState(): any | undefined {
        if (this.vsCodeApi) {
            return this.vsCodeApi.getState();
        } else {
            const state = localStorage.getItem("vscodeState");
            return state ? JSON.parse(state) : undefined;
        }
    }

    public setState<T>(newState: T): T {
        if (this.vsCodeApi) {
            return this.vsCodeApi.setState(newState);
        } else {
            localStorage.setItem("vscodeState", JSON.stringify(newState));
            return newState;
        }
    }
}

export const vscode = new VSCodeAPIWrapper();
