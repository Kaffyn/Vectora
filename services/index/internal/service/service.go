package service

import (
	"database/sql"

	"github.com/Kaffyn/Vectora-Index/internal/db"
)

// Service contém todos os serviços de negócio
type Service struct {
	db *db.PostgresDB
}

// NewService cria uma nova instância de Service
func NewService(database *db.PostgresDB) *Service {
	return &Service{
		db: database,
	}
}

// GetDB retorna a conexão com o banco de dados para operações diretas
func (s *Service) GetDB() *sql.DB {
	return s.db.GetConnection()
}
