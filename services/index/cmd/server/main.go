package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/Kaffyn/Vectora-Index/internal/db"
	grpcserver "github.com/Kaffyn/Vectora-Index/internal/grpc"
	"github.com/Kaffyn/Vectora-Index/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Carregar configurações do ambiente
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "3000"
	}

	host := os.Getenv("GRPC_HOST")
	if host == "" {
		host = "0.0.0.0"
	}

	// Obter DSN do Supabase
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatalf("[ERROR] DATABASE_URL environment variable is required")
	}

	log.Printf("[INFO] Iniciando Vectora Index Service em %s:%s\n", host, port)

	// Conectar ao banco de dados
	database, err := db.NewPostgresDB(dsn)
	if err != nil {
		log.Fatalf("[ERROR] Erro ao conectar no banco de dados: %v\n", err)
	}
	defer database.Close()

	// Verificar conexão
	ctx, cancel := context.WithTimeout(context.Background(), 5*1000000000)
	if err := database.Ping(ctx); err != nil {
		cancel()
		log.Fatalf("[ERROR] Erro ao fazer ping no banco de dados: %v\n", err)
	}
	cancel()

	log.Printf("[INFO] Conexão com banco de dados estabelecida com sucesso\n")

	// Criar serviço
	svc := service.NewService(database)

	// Criar servidor gRPC
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(100 * 1024 * 1024), // 100 MB para uploads
		grpc.MaxSendMsgSize(100 * 1024 * 1024),
	}

	grpcSrv := grpcserver.NewGRPCServer(opts...)

	// Registrar reflection para inspeção (desenvolvimento)
	reflection.Register(grpcSrv)

	// TODO: Registrar IndexServiceServer quando protobuf for gerado
	indexServer := grpcserver.NewIndexServiceServer(svc)
	_ = indexServer // Evitar unused warning

	// Converter port para int
	portNum, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("[ERROR] PORT inválida: %v\n", err)
	}

	// Canal para sinais de shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Goroutine para escutar sinais
	go func() {
		sig := <-sigChan
		log.Printf("[INFO] Recebido sinal: %v\n", sig)
		log.Printf("[INFO] Encerrando servidor gRPC...\n")
		grpcSrv.GracefulStop()
	}()

	// Iniciar servidor (bloqueante)
	if err := grpcserver.StartServer(grpcSrv, host, portNum); err != nil {
		log.Fatalf("[ERROR] Servidor parou com erro: %v\n", err)
	}

	log.Printf("[INFO] Servidor encerrado com sucesso\n")
}
