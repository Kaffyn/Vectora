import { expect, test, describe } from "bun:test";
import { handleSearch } from "./api";

describe("Search API: 300% Proof", () => {
  test("HappyPath: should return results for valid query", async () => {
    const query = "test";
    const response = await handleSearch(query);
    expect(response.status).toBe(200);
    expect(response.results!.length).toBeGreaterThan(0);
  });

  test("Negative: should return 400 for empty query", async () => {
    const query = "";
    const response = await handleSearch(query);
    expect(response.status).toBe(400);
    expect(response.error).toBe("Query required");
  });

  test("EdgeCase: should handle special characters", async () => {
    const query = "!@#$%^&*()";
    const response = await handleSearch(query);
    expect(response.status).toBe(200);
    expect(response.results![0].content).toContain(query);
  });
});
