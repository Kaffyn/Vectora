package tools

import (
	"unicode/utf8"
)

// SanitizeResult força o cast limpo garantindo conformidade JSON-RPC contra dados binários acidentais ou overload massivo
func SanitizeResult(res *ToolResult) *ToolResult {
	if !utf8.ValidString(res.Output) {
		res.Output = "Binary or Invalid UTF-8 Content Blocked"
		res.IsError = true
		return res
	}

	if len(res.Output) > 100000 {
		res.Output = res.Output[:100000] + "\n... [OUTPUT TRUNCATED DUE TO EXTREME LENGTH]"
	}

	clean := make([]rune, 0, len(res.Output))
	for _, r := range res.Output {
		if r == '\n' || r == '\t' || (r >= 32 && r <= 126) || utf8.RuneLen(r) > 1 {
			clean = append(clean, r)
		}
	}
	res.Output = string(clean)

	return res
}
