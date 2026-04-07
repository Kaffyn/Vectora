package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// PostgresDB representa a conexão com banco de dados PostgreSQL/Supabase
type PostgresDB struct {
	conn *sql.DB
}

// NewPostgresDB cria e retorna uma nova instância de PostgresDB
func NewPostgresDB(dsn string) (*PostgresDB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("DSN (connection string) é obrigatório")
	}

	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir conexão com PostgreSQL: %w", err)
	}

	// Verificar conexão
	ctx, cancel := context.WithTimeout(context.Background(), 5*1000000000) // 5 segundos
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("erro ao conectar no PostgreSQL: %w", err)
	}

	// Configurar pool de conexões
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)

	return &PostgresDB{conn: conn}, nil
}

// Ping realiza health check da conexão
func (db *PostgresDB) Ping(ctx context.Context) error {
	return db.conn.PingContext(ctx)
}

// Close fecha a conexão com o banco de dados
func (db *PostgresDB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// GetConnection retorna a conexão SQL para queries diretas
func (db *PostgresDB) GetConnection() *sql.DB {
	return db.conn
}
