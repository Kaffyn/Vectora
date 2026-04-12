// fs-promises stub for browser
export const readFile = async () => "";
export const writeFile = async () => {};
export const stat = async () => ({ isDirectory: () => false, isFile: () => true });
export const readdir = async () => [];
export const readlink = async () => "";
export const lstat = async () => ({ isDirectory: () => false, isFile: () => true, isSymbolicLink: () => false });
export const mkdir = async () => {};
export const rm = async () => {};
export const existsSync = () => false;
export const lstatSync = () => ({ isDirectory: () => false, isFile: () => true, isSymbolicLink: () => false });
export const readFileSync = () => "";
export const mkdirSync = () => {};
export const writeFileSync = () => {};
export const readdirSync = () => [];
export const unlinkSync = () => {};
export const statSync = () => ({ isDirectory: () => false, isFile: () => true });
export const chmodSync = () => {};
export const constants = {};

export default {
  chmodSync: () => {},
};
