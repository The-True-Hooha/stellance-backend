package config

import (
	"context"
	"log/slog"
	"sync"

	database "github.com/The-True-Hooha/stellance-backend.git/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AppContainer struct {
	Database *pgxpool.Pool
	Log      *slog.Logger
}

var (
	container *AppContainer
	once      sync.Once
)

func InitializeContainer(ctx context.Context, dbConfig database.PostgresConfig, log *slog.Logger) error {
	var initError error

	once.Do(func() {
		pool, err := database.CreateNewPostgresConnection(ctx, dbConfig)
		if err != nil {
			log.Error("failed to initialize database pool", "error", err.Error())
			initError = err
			return
		}
		log.Info("::: database connection pool initialized successfully ::===>")
		container = &AppContainer{
			Database: pool.Pool,
			Log:      log,
		}
	})
	return initError
}

func GetAppContainer() *AppContainer {
	if container == nil {
		panic("App container not initialized. Call Initialize container first")
	}
	return container
}

func Shutdown() {
	if container != nil && container.Database != nil {
		container.Log.Info("shutting down... closing the database connection pool")
		container.Database.Close()
	}
}
