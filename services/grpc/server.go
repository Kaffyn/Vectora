package grpc

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

// NewGRPCServer cria e retorna um novo servidor gRPC
func NewGRPCServer(opts ...grpc.ServerOption) *grpc.Server {
	// Adicionar opções padrão
	defaultOpts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(100 * 1024 * 1024), // 100 MB para uploads
		grpc.MaxSendMsgSize(100 * 1024 * 1024),
	}

	// Combinar com opções fornecidas
	allOpts := append(defaultOpts, opts...)

	return grpc.NewServer(allOpts...)
}

// StartServer inicia o servidor gRPC em um port específico
func StartServer(server *grpc.Server, host string, port int) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("erro ao criar listener TCP em %s: %w", addr, err)
	}

	log.Printf("[INFO] Servidor gRPC iniciando em %s\n", addr)

	// Serve é bloqueante
	return server.Serve(listener)
}
