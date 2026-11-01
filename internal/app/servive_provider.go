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
	"time"

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

func (s *serviceProvider) PGConfig() (config.PGConfig, error) {
	if s.pgConfig == nil {
		cfg, err := env.NewPGConfig()
		if err != nil {
			return nil, err
		}

		s.pgConfig = cfg
	}

	return s.pgConfig, nil
}

func (s *serviceProvider) initDB(ctx context.Context) error {
	cfg, err := s.PGConfig()
	if err != nil {
		return err
	}
	p, err := pgxpool.New(ctx, cfg.DSN())
	if err != nil {
		return err
	}
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := p.Ping(pingCtx); err != nil {
		p.Close()
		return err
	}
	s.pool = p
	s.db = stdlib.OpenDBFromPool(p)
	return nil
}

func (s *serviceProvider) ChatRepository(ctx context.Context) repository.ChatRepository {
	if s.chatRepo == nil {
		s.chatRepo = chatrepository.NewChatRepository(s.db)
	}

	return s.chatRepo
}

func (sp *serviceProvider) ChatService(ctx context.Context) service.ChatService {
	if sp.chatSrv == nil {
		sp.chatSrv = chatsrv.NewChatService(sp.ChatRepository(ctx), sp.Logger(ctx))
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

func (s *serviceProvider) DBClient() *sql.DB {
	return s.db
}
