// path stub for browser
export const join = (...parts: string[]) => parts.filter(Boolean).join("/").replace(/\/+/g, "/");
export const dirname = (p: string) => p.split("/").slice(0, -1).join("/") || ".";
export const basename = (p: string) => p.split("/").pop() || "";
export const extname = (p: string) => {
  const b = basename(p);
  const i = b.lastIndexOf(".");
  return i === -1 ? "" : b.slice(i);
};
export const resolve = (...parts: string[]) => join(...parts);
export const relative = (from: string, to: string) => to;
export const sep = "/";
export const delimiter = ":";
export const isAbsolute = (p: string) => p.startsWith("/");

export default {
};
