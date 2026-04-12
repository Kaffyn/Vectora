// child_process stub for browser
export const spawn = () => ({
  on: () => {},
  stdout: { on: () => {} },
  stderr: { on: () => {} },
  kill: () => {},
});
export const exec = () => {};
export const execSync = () => "";
export const fork = () => {};

export default {
};
