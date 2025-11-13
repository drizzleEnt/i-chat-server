package app

import (
	"chatsrv/internal/config"
	"chatsrv/internal/config/env"
	"chatsrv/internal/controller"
	chatctrl "chatsrv/internal/controller/chat"
	"chatsrv/internal/repository"
	chatrepository "chatsrv/internal/repository/chat"
	"chatsrv/internal/service"
	chatsrv "chatsrv/internal/service/chat"
	"context"
	"database/sql"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

type serviceProvider struct {
	logger *zap.Logger

	pgConfig config.PGConfig
	httpCfg  config.HttpConfig

	db   *sql.DB
	pool *pgxpool.Pool

	chatImpl controller.ChatController
	chatSrv  service.ChatService
	chatRepo repository.ChatRepository
}

func newServiceProvider() *serviceProvider {
	return &serviceProvider{}
}

func (sp *serviceProvider) HttpConfig() config.HttpConfig {
	if sp.httpCfg == nil {
		sp.httpCfg = env.NewHttpConfig()
	}
	return sp.httpCfg
}

func (s *serviceProvider) PGConfig() config.PGConfig {
	if s.pgConfig == nil {
		cfg, err := env.NewPGConfig()
		if err != nil {
			s.Logger(context.Background()).Error("failed to get pg config", zap.Error(err))
			os.Exit(1)
		}

		s.pgConfig = cfg
	}

	return s.pgConfig
}

func (s *serviceProvider) ChatRepository(ctx context.Context) repository.ChatRepository {
	if s.chatRepo == nil {
		s.chatRepo = chatrepository.NewChatRepository(s.DBClient(ctx))
	}

	return s.chatRepo
}

func (sp *serviceProvider) ChatService(ctx context.Context) service.ChatService {
	if sp.chatSrv == nil {
		sp.chatSrv = chatsrv.NewChatService(sp.ChatRepository(ctx))
	}
	return sp.chatSrv
}

func (sp *serviceProvider) ChatController(ctx context.Context) controller.ChatController {
	if sp.chatImpl == nil {
		sp.chatImpl = chatctrl.NewChatController(
			chatctrl.WithLogger(sp.Logger(ctx)),
			chatctrl.WithService(sp.ChatService(ctx)),
		)
	}
	return sp.chatImpl
}

func (sp *serviceProvider) Logger(ctx context.Context) *zap.Logger {
	if sp.logger == nil {
		logger := zap.New(getCore(getAtomicLevel()))
		sp.logger = logger
	}

	return sp.logger
}

func (s *serviceProvider) DBClient(ctx context.Context) *sql.DB {
	if s.db == nil {
		pool, err := pgxpool.New(ctx, s.PGConfig().DSN())
		if err != nil {
			s.Logger(context.Background()).Error("failed to connect to db", zap.Error(err))
			os.Exit(1)
		}
		s.pool = pool

		dbc := stdlib.OpenDBFromPool(pool)
		err = dbc.Ping()
		if err != nil {
			s.Logger(context.Background()).Error("failed to ping db", zap.Error(err))
			os.Exit(1)
		}
		s.db = dbc
	}

	return s.db
}
