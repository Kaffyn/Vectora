package ipc

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

func TestIPCFullCycle_Memory(t *testing.T) {
	// Setup custom Socket p/ teste isolado
	tmpDir, _ := os.MkdirTemp("", "vectora-ipc-*")
	defer os.RemoveAll(tmpDir)

	addr := filepath.Join(tmpDir, "test.sock")
	if runtime.GOOS == "windows" {
		addr = "127.0.0.1:42785" // Porta de Teste
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := &Server{
		addr:     addr,
		handlers: make(map[string]RouterFunc),
		clients:  make(map[net.Conn]bool),
		ctx:      ctx,
		cancel:   cancel,
	}

	// 1. Registra Mock Method "ping"
	server.Register("ping", func(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
		var req map[string]string
		json.Unmarshal(payload, &req)
		return map[string]string{"reply": "pong_" + req["msg"]}, nil
	})

	// 2. Start Server
	var l net.Listener
	var err error
	if runtime.GOOS == "windows" {
		l, err = net.Listen("tcp", addr)
	} else {
		l, err = net.Listen("unix", addr)
	}
	if err != nil {
		t.Fatalf("falha ao iniciar servidor teste: %v", err)
	}
	server.listener = l
	go func() {
		for {
			conn, err := server.listener.Accept()
			if err != nil { return }
			server.clientsLock.Lock()
			server.clients[conn] = true
			server.clientsLock.Unlock()
			go server.handleConnection(conn)
		}
	}()
	defer server.Shutdown()

	// 3. Client Connect
	client := &Client{
		addr:    addr,
		pending: make(map[string]chan IPCMessage),
		ctx:     ctx,
		cancel:  cancel,
	}
	if err := client.Connect(); err != nil {
		t.Fatalf("cliente falhou ao conectar: %v", err)
	}
	defer client.Close()

	// 4. Send Ping e valida Pong
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			var res map[string]string
			payload := map[string]string{"msg": "hi"}
			if err := client.Send(context.Background(), "ping", payload, &res); err != nil {
				t.Errorf("cliente recv falhou: %v", err)
				return
			}
			if res["reply"] != "pong_hi" {
				t.Errorf("Resposta errada no IPC: %s", res["reply"])
			}
		}(i)
	}
	wg.Wait()
}
