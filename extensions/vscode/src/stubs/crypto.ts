// crypto stub for browser
export const randomBytes = (size: number) => {
  const arr = new Uint8Array(size);
  window.crypto.getRandomValues(arr);
  return {
    toString: (encoding: string) =>
      Array.from(arr)
        .map((b) => b.toString(16).padStart(2, "0"))
        .join(""),
  };
};

export const createHash = (algorithm: string) => ({
  update: (data: string) => ({
    digest: (encoding: string) => "mock-hash",
  }),
});

export default {
};
