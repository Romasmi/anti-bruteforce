package api

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Romasmi/anti-bruteforce/internal/logger"
	"github.com/Romasmi/anti-bruteforce/internal/repository/iplist"
	grpcserver "github.com/Romasmi/anti-bruteforce/internal/server/grpc"
	internalhttp "github.com/Romasmi/anti-bruteforce/internal/server/http"
	"github.com/Romasmi/anti-bruteforce/internal/usecases"
	"github.com/Romasmi/anti-bruteforce/migrations"
	"github.com/Romasmi/anti-bruteforce/pkg/ratelimiter"
	leakybucket "github.com/Romasmi/anti-bruteforce/pkg/ratelimiter/algorithms/leackybucket"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
)

type App struct {
	config      Config
	logger      *logger.Logger
	db          *sql.DB
	redisClient *redis.Client
	grpcServer  *grpcserver.Server
	httpServer  *internalhttp.Server
	IPRepo      *iplist.IPRepo
}

func New(conf Config, l *logger.Logger) *App {
	return &App{
		config: conf,
		logger: l,
	}
}

func (a *App) Init(ctx context.Context) error {
	database, err := openDB(a.config.DB.DSN)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	a.db = database

	if err := migrations.Migrate(a.db); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	if a.config.Redis.Addr != "" {
		rc := redis.NewClient(&redis.Options{Addr: a.config.Redis.Addr})
		if err := rc.Ping(ctx).Err(); err != nil {
			_ = rc.Close()
			return fmt.Errorf("ping redis: %w", err)
		}
		a.redisClient = rc
		a.logger.Info("using Redis-backed rate-limit buckets")
	} else {
		a.logger.Info("no Redis configured — using in-memory rate-limit buckets")
	}

	a.IPRepo = iplist.NewIPRepo(a.db)
	limiter := buildRateLimiter(a.config.RateLimiter, a.IPRepo, a.redisClient)
	ucs := usecases.NewUsecases(a.logger, limiter, a.IPRepo)

	a.grpcServer = grpcserver.NewServer(a.logger, ucs)
	grpcAddr := fmt.Sprintf("%s:%s", a.config.GRPC.Host, a.config.GRPC.Port)
	a.httpServer = internalhttp.NewServer(a.logger, grpcAddr, a.config.HTTP.Host, a.config.HTTP.Port)

	return nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 2)

	go func() {
		if err := a.grpcServer.Start(a.config.GRPC.Host, a.config.GRPC.Port); err != nil {
			errCh <- fmt.Errorf("grpc server: %w", err)
		}
	}()

	go func() {
		if err := a.httpServer.Start(ctx); err != nil {
			errCh <- fmt.Errorf("http server: %w", err)
		}
	}()

	a.logger.Info("anti-bruteforce is running...")

	select {
	case <-ctx.Done():
		a.logger.Info("Stopping servers...")
	case err := <-errCh:
		a.logger.Error("Server error: " + err.Error())
	}

	stopCtx, stopCancel := context.WithTimeout(context.Background(), time.Second*3)
	defer stopCancel()

	if err := a.httpServer.Stop(stopCtx); err != nil {
		a.logger.Error("failed to stop http server: " + err.Error())
	}
	a.grpcServer.Stop()

	if err := a.db.Close(); err != nil {
		a.logger.Error("failed to close db: " + err.Error())
	}

	if a.redisClient != nil {
		if err := a.redisClient.Close(); err != nil {
			a.logger.Error("failed to close redis: " + err.Error())
		}
	}

	return nil
}

func openDB(dsn string) (*sql.DB, error) {
	database, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := database.PingContext(ctx); err != nil {
		_ = database.Close()
		return nil, err
	}
	return database, nil
}

func buildRateLimiter(conf RateLimiterConf, ipRepo leakybucket.Repository, rc *redis.Client) ratelimiter.RateLimiter {
	ttl := func(c BucketConf) time.Duration {
		return 2 * time.Duration(c.WindowSeconds) * time.Second
	}
	leakRate := func(c BucketConf) float64 { return c.Capacity / c.WindowSeconds }

	if rc != nil {
		return ratelimiter.NewRateLimiter(ratelimiter.AlgorithmMap{
			usecases.StrategyLogin: leakybucket.NewPersistentLeakyBucket(leakybucket.PersistentLeakyBucketParams{
				Client: rc, Prefix: usecases.StrategyLogin,
				Capacity: conf.Login.Capacity, LeakRate: leakRate(conf.Login), TTL: ttl(conf.Login),
			}),
			usecases.StrategyPassword: leakybucket.NewPersistentLeakyBucket(leakybucket.PersistentLeakyBucketParams{
				Client: rc, Prefix: usecases.StrategyPassword,
				Capacity: conf.Password.Capacity, LeakRate: leakRate(conf.Password), TTL: ttl(conf.Password),
			}),
			usecases.StrategyIP: leakybucket.NewPersistentLeakyBucket(leakybucket.PersistentLeakyBucketParams{
				Client: rc, Prefix: usecases.StrategyIP, Repo: ipRepo,
				Capacity: conf.IP.Capacity, LeakRate: leakRate(conf.IP), TTL: ttl(conf.IP),
			}),
		})
	}

	return ratelimiter.NewRateLimiter(ratelimiter.AlgorithmMap{
		usecases.StrategyLogin:    newInMemoryBucket(conf.Login, nil),
		usecases.StrategyPassword: newInMemoryBucket(conf.Password, nil),
		usecases.StrategyIP:       newInMemoryBucket(conf.IP, ipRepo),
	})
}

func newInMemoryBucket(conf BucketConf, repo leakybucket.Repository) *leakybucket.LeakyBucket {
	return leakybucket.NewLeakyBucket(leakybucket.LeakyBucketParams{
		Capacity: conf.Capacity,
		LeakRate: conf.Capacity / conf.WindowSeconds,
		TTL:      2 * time.Duration(conf.WindowSeconds) * time.Second,
		Repo:     repo,
	})
}
