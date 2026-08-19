package server

import (
	"github.com/fauzie/golang-sekeleton/internal/config"
	"github.com/fauzie/golang-sekeleton/internal/repository"
	"github.com/fauzie/golang-sekeleton/internal/repository/cache"
	"github.com/fauzie/golang-sekeleton/internal/repository/database"
	"github.com/fauzie/golang-sekeleton/internal/repository/mq"
	"github.com/fauzie/golang-sekeleton/pkg/logger"
)

// buildRepository assembles every infrastructure dependency behind
// repository.RepoFactory: the database (MySQL or PostgreSQL, per
// cfg.Database.Driver), the master cache and, if configured, a read-only
// replica, and one dynamic MQ publisher client per configured topic
// (Kafka/RabbitMQ/Pub/Sub, per each topic's own type).
func buildRepository(cfg *config.Config, log *logger.Logger) (*repository.Repository, error) {
	factories := []repository.RepoFactory{
		&database.Conf{
			Driver:                 cfg.Database.Driver,
			Host:                   cfg.Database.Host,
			Port:                   cfg.Database.Port,
			User:                   cfg.Database.User,
			Password:               cfg.Database.Password,
			Name:                   cfg.Database.Name,
			SSLMode:                cfg.Database.SSLMode,
			QueryTimeoutSeconds:    cfg.Database.QueryTimeoutSeconds,
			MaxOpenConns:           cfg.Database.MaxOpenConns,
			MaxIdleConns:           cfg.Database.MaxIdleConns,
			ConnMaxLifetimeMinutes: cfg.Database.ConnMaxLifetimeMinutes,
		},
	}

	if cfg.Cache.Enabled {
		factories = append(factories, &cache.Conf{
			Mode:     cfg.Cache.Mode,
			Addrs:    cfg.Cache.Addrs,
			Password: cfg.Cache.Password,
			DB:       cfg.Cache.DB,
		})

		if cfg.Cache.Replica != nil {
			factories = append(factories, &cache.Conf{
				Mode:     cfg.Cache.Replica.Mode,
				Addrs:    cfg.Cache.Replica.Addrs,
				Password: cfg.Cache.Replica.Password,
				DB:       cfg.Cache.Replica.DB,
				Replica:  true,
			})
		}
	}

	if len(cfg.MessagePublishers.Topics) > 0 {
		factories = append(factories, &mq.PublisherConf{
			Topics: cfg.ResolvedPublishers(),
			Logger: log,
		})
	}

	return repository.NewRepository(factories, log)
}
