import { expect, test } from "bun:test";
import { renderMessage } from "./Message";

test("should render message text", () => {
  const message = renderMessage("Hello World");
  expect(message).toContain("Hello World");
});
