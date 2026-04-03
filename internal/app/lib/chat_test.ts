// Unit test for chat utility (placeholder logic for now)

export const formatMessage = (msg: string): string => {
  return msg.trim();
};

if (import.meta.vitest) {
  const { it, expect } = import.meta.vitest;
  it("should trim messages", () => {
    expect(formatMessage("  hello  ")).toBe("hello");
  });
}
