package dbrepo

import (
	"database/sql"

	"github.com/mohammedsalman999/silsila/internal/config"
	"github.com/mohammedsalman999/silsila/internal/repository"
)

type mysqlDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}



func NewMysqlRepo(conn *sql.DB, a *config.AppConfig) repository.DatabaseRepo {
	return &mysqlDBRepo{
		App: a,
		DB:  conn,
	}
}
