import { expect, test, describe } from "bun:test";
import { renderChat } from "./Chat";

describe("Chat Component: 300% Proof", () => {
  test("HappyPath: should render user and assistant messages", () => {
    const messages = [
      { role: "user", content: "Query" },
      { role: "assistant", content: "Response" },
    ];
    const chat = renderChat(messages);
    expect(chat).toContain("Query");
    expect(chat).toContain("Response");
  });

  test("Negative: should handle empty message list", () => {
    const chat = renderChat([]);
    expect(chat).toContain("No messages yet");
  });

  test("EdgeCase: should handle 200 messages without crash", () => {
    const bulk = Array.from({ length: 200 }, (_, i) => ({
      role: i % 2 === 0 ? "user" : "assistant",
      content: `Message ${i}`,
    }));
    const chat = renderChat(bulk);
    expect(chat).toContain("Message 199");
  });
});
