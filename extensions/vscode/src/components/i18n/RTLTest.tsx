/**
 * RTL Testing Component
 * Phase 11: RTL Languages Support & i18n Enhancement
 * Visual testing and validation tool for RTL languages
 */

import React, { useMemo } from "react";
import { useRTL, useRTLStyles } from "../../hooks/useRTL";
import type { SupportedLanguage } from "../../i18n/types";

interface RTLTestProps {
  language: SupportedLanguage;
}

/**
 * RTL Test Component
 */
export const RTLTest: React.FC<RTLTestProps> = ({ language }) => {
  const { direction, isRTL, shouldFlipIcon, getFlippedIcon } = useRTL();
  const { getMargin, getPadding, getTextAlign } = useRTLStyles();

  const testCases = useMemo(
    () => [
      {
        name: "Text Direction",
        test: direction === "rtl" ? "✓ RTL" : "✓ LTR",
        expected: direction,
      },
      {
        name: "Language Code",
        test: language,
        expected: language,
      },
      {
        name: "Document Direction",
        test: document.documentElement.dir || "not set",
        expected: direction,
      },
      {
        name: "Document Language",
        test: document.documentElement.lang || "not set",
        expected: language,
      },
    ],
    [direction, language],
  );

  const iconTests = useMemo(
    () => [
      { original: "arrow-left", flipped: getFlippedIcon("arrow-left") },
      { original: "arrow-right", flipped: getFlippedIcon("arrow-right") },
      { original: "chevron-left", flipped: getFlippedIcon("chevron-left") },
      { original: "chevron-right", flipped: getFlippedIcon("chevron-right") },
    ],
    [getFlippedIcon],
  );

  return (
    <div
      dir={direction}
      style={{
        padding: "2rem",
        maxWidth: "800px",
        margin: "0 auto",
        textAlign: direction === "rtl" ? "right" : "left",
      }}
    >
      <h2>RTL Testing Dashboard</h2>

      {/* Basic Tests */}
      <section style={{ marginBottom: "2rem" }}>
        <h3>Basic Configuration</h3>
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))",
            gap: "1rem",
          }}
        >
          {testCases.map((test, idx) => (
            <div
              key={idx}
              style={{
                padding: "1rem",
                border: "1px solid #ccc",
                borderRadius: "8px",
                backgroundColor: test.test === test.expected ? "#d4edda" : "#f8d7da",
              }}
            >
              <div style={{ fontWeight: "bold", marginBottom: "0.5rem" }}>
                {test.name}
              </div>
              <div style={{ fontSize: "0.9rem" }}>
                <strong>Result:</strong> {test.test}
              </div>
              <div style={{ fontSize: "0.9rem" }}>
                <strong>Expected:</strong> {test.expected}
              </div>
            </div>
          ))}
        </div>
      </section>

      {/* Icon Tests */}
      <section style={{ marginBottom: "2rem" }}>
        <h3>Icon Flipping Tests</h3>
        <div style={{ display: "grid", gridTemplateColumns: "repeat(2, 1fr)", gap: "1rem" }}>
          {iconTests.map((test, idx) => (
            <div key={idx} style={{ padding: "1rem", border: "1px solid #ccc" }}>
              <div style={{ fontWeight: "bold" }}>
                Original: {test.original}
              </div>
              <div style={{ fontSize: "0.9rem", marginTop: "0.5rem" }}>
                Flipped: {test.flipped}
              </div>
              {shouldFlipIcon(test.original) && (
                <div style={{ fontSize: "0.85rem", color: "green", marginTop: "0.25rem" }}>
                  ✓ Should flip
                </div>
              )}
            </div>
          ))}
        </div>
      </section>

      {/* Layout Tests */}
      <section style={{ marginBottom: "2rem" }}>
        <h3>Layout Direction Tests</h3>

        {/* Margin Test */}
        <div style={{ marginBottom: "1rem" }}>
          <h4>Margin Handling</h4>
          <div
            style={{
              margin: getMargin("0", "1rem", "0", "0.5rem"),
              padding: "1rem",
              border: "2px solid blue",
              backgroundColor: "#e7f3ff",
            }}
          >
            This element has dynamic margins (top: 0, right/left: 1rem, bottom: 0)
          </div>
        </div>

        {/* Text Alignment Test */}
        <div style={{ marginBottom: "1rem" }}>
          <h4>Text Alignment</h4>
          <div
            style={{
              textAlign: getTextAlign("left") as any,
              padding: "1rem",
              border: "2px solid green",
              backgroundColor: "#f0fff0",
            }}
          >
            This text should be {getTextAlign("left")} aligned
          </div>
        </div>

        {/* Flex Direction Test */}
        <div style={{ marginBottom: "1rem" }}>
          <h4>Flex Direction</h4>
          <div
            style={{
              display: "flex",
              flexDirection: direction === "rtl" ? "row-reverse" : "row",
              gap: "1rem",
              padding: "1rem",
              border: "2px solid orange",
              backgroundColor: "#fffacd",
            }}
          >
            <div
              style={{
                flex: 1,
                padding: "0.5rem",
                backgroundColor: "rgba(255, 99, 71, 0.2)",
              }}
            >
              Box 1
            </div>
            <div
              style={{
                flex: 1,
                padding: "0.5rem",
                backgroundColor: "rgba(0, 100, 200, 0.2)",
              }}
            >
              Box 2
            </div>
            <div
              style={{
                flex: 1,
                padding: "0.5rem",
                backgroundColor: "rgba(34, 139, 34, 0.2)",
              }}
            >
              Box 3
            </div>
          </div>
        </div>
      </section>

      {/* Multilingual Test */}
      <section style={{ marginBottom: "2rem" }}>
        <h3>Multilingual Content Test</h3>
        <div
          style={{
            padding: "1rem",
            border: "1px solid #ccc",
            borderRadius: "8px",
            backgroundColor: "#f9f9f9",
          }}
        >
          <p>English: Hello, this is a test message.</p>
          <p>Arabic: السلام عليكم، هذه رسالة اختبار.</p>
          <p>Hebrew: שלום, זו הודעת בדיקה.</p>
          <p>Persian: سلام، این یک پیام تستی است.</p>
          <p>Urdu: السلام علیکم، یہ ایک ٹیسٹ پیغام ہے۔</p>
          <p>Chinese: 你好，这是一条测试信息。</p>
          <p>Japanese: こんにちは、これはテストメッセージです。</p>
          <p>Korean: 안녕하세요, 이것은 테스트 메시지입니다.</p>
        </div>
      </section>

      {/* CSS Logical Properties Test */}
      <section style={{ marginBottom: "2rem" }}>
        <h3>CSS Logical Properties</h3>
        <div
          style={{
            paddingInlineStart: "2rem",
            paddingInlineEnd: "1rem",
            marginInlineStart: "1rem",
            marginInlineEnd: "2rem",
            borderInlineStart: "4px solid red",
            padding: "1rem",
            backgroundColor: "#ffe4e1",
          }}
        >
          This element uses CSS Logical Properties that automatically adjust for
          {direction === "rtl" ? " RTL" : " LTR"} text direction
        </div>
      </section>

      {/* Status Summary */}
      <section style={{ marginBottom: "2rem" }}>
        <h3>Status Summary</h3>
        <div
          style={{
            padding: "1rem",
            borderRadius: "8px",
            backgroundColor: isRTL ? "#ffebee" : "#e8f5e9",
            borderLeft: `4px solid ${isRTL ? "#f44336" : "#4caf50"}`,
          }}
        >
          <p>
            <strong>Current Mode:</strong> {direction.toUpperCase()}
          </p>
          <p>
            <strong>Language:</strong> {language}
          </p>
          <p>
            <strong>RTL Enabled:</strong> {isRTL ? "Yes" : "No"}
          </p>
          <p>
            <strong>Document Direction:</strong> {document.documentElement.dir}
          </p>
          <p>
            <strong>Document Language:</strong> {document.documentElement.lang}
          </p>
        </div>
      </section>
    </div>
  );
};

export default RTLTest;
