package server_test

import (
	"testing"

	"github.com/Kaffyn/Vectora/bridge/mcp/server"
)

func TestServerInit(t *testing.T) {
	t.Run("should initialize mcp server", func(t *testing.T) {
		s := server.NewServer("Vectora-Bridge")
		if s.Name != "Vectora-Bridge" {
			t.Error("server name mismatch")
		}
	})
}
