package app

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/MrSwed/go-musthave-metrics/internal/server/closer"
	"github.com/MrSwed/go-musthave-metrics/internal/server/config"
	"github.com/MrSwed/go-musthave-metrics/internal/server/constant"
	"github.com/MrSwed/go-musthave-metrics/internal/server/handler"
	myMigrate "github.com/MrSwed/go-musthave-metrics/internal/server/migrate"
	"github.com/MrSwed/go-musthave-metrics/internal/server/repository"
	"github.com/MrSwed/go-musthave-metrics/internal/server/service"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func buildInfo(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}

type BuildMetadata struct {
	Version string `json:"buildVersion"`
	Date    string `json:"buildDate"`
	Commit  string `json:"buildCommit"`
}

type app struct {
	ctx        context.Context
	stop       context.CancelFunc
	cfg        *config.Config
	build      BuildMetadata
	eg         *errgroup.Group
	http       *http.Server
	srv        *service.Service
	log        *zap.Logger
	db         *sqlx.DB
	closer     *closer.Closer
	isNewStore bool
	lockDB     chan struct{}
}

func NewApp(c context.Context, stop context.CancelFunc,
	metadata BuildMetadata, cfg *config.Config, log *zap.Logger) *app {
	g, ctx := errgroup.WithContext(c)
	a := app{
		ctx:        ctx,
		stop:       stop,
		build:      metadata,
		cfg:        cfg,
		eg:         g,
		isNewStore: true,
		closer:     &closer.Closer{},
		log:        log,
		lockDB:     make(chan struct{}),
	}

	a.log.Info("Init app", zap.Any(`Build info`, map[string]string{
		`Build version`: buildInfo(a.build.Version),
		`Build date`:    buildInfo(a.build.Date),
		`Build commit`:  buildInfo(a.build.Commit)}))

	a.maybeConnectDB()

	a.srv = service.NewService(repository.NewRepository(&a.cfg.StorageConfig, a.db), &a.cfg.StorageConfig)
	h := handler.NewHandler(a.srv, &a.cfg.WEB, a.log)

	a.maybeRestoreStore()
	a.maybeRunStoreSaver()

	a.http = &http.Server{Addr: a.cfg.Address, Handler: h.HTTPHandler()}

	return &a
}

func (a *app) maybeConnectDB() {
	if len(a.cfg.DatabaseDSN) > 0 {
		var err error
		if a.db, err = sqlx.Connect("postgres", a.cfg.DatabaseDSN); err != nil {
			a.log.Fatal("cannot connect db", zap.Error(err))
		}
		a.isNewStore = false
		a.log.Info("DB connected")
		versions, errM := myMigrate.Migrate(a.db.DB)
		switch {
		case errors.Is(errM, migrate.ErrNoChange):
			a.log.Info("DB migrate: ", zap.Any("info", errM), zap.Any("versions", versions))
		case errM == nil:
			a.log.Info("DB migrate: new applied ", zap.Any("versions", versions))
			a.isNewStore = versions[0] == 0
		default:
			a.log.Fatal("DB migrate: ", zap.Any("versions", versions), zap.Error(errM))
		}
	}
}

func (a *app) maybeRestoreStore() {
	if a.cfg.FileStoragePath != "" && a.cfg.StorageRestore {
		if a.isNewStore {
			if n, er := a.srv.RestoreFromFile(a.ctx); er != nil {
				a.log.Error("File storage restore", zap.Error(er))
			} else {
				a.log.Info("File storage restored success", zap.Any("records", n))
			}
		} else {
			a.log.Info("Storage not restored - it is not new db store used")
		}
	}
}

func (a *app) maybeRunStoreSaver() {
	if a.cfg.FileStoragePath != "" && a.cfg.FileStoreInterval > 0 {
		a.eg.Go(func() error {
			for {
				select {
				case <-time.After(time.Duration(a.cfg.FileStoreInterval) * time.Second):
					if n, er := a.srv.SaveToFile(a.ctx); er != nil {
						a.log.Error("Storage save", zap.Error(er))
					} else {
						a.log.Info("Storage saved", zap.Any("records", n))
					}
				case <-a.ctx.Done():
					a.log.Info("Store save on interval finished")
					return nil
				}
			}
		})
	}
}

func (a *app) shutdownFileStore(ctx context.Context) (err error) {
	defer close(a.lockDB)
	var n int64
	if n, err = a.srv.SaveToFile(ctx); err == nil {
		a.log.Info("Storage saved", zap.Any("records", n))
	}
	return

}

func (a *app) shutdownDBStore(_ context.Context) (err error) {
	if a.db != nil {
		<-a.lockDB
		if err = a.db.Close(); err == nil {
			a.log.Info("Db Closed")
		}
	}
	return
}

func (a *app) Run() {
	a.log.Info("Start server", zap.Any("Config", a.cfg))

	a.closer.Add("WEB", a.http.Shutdown)

	if a.cfg.FileStoragePath != "" {
		a.closer.Add("Storage save", a.shutdownFileStore)
	} else {
		close(a.lockDB)
	}

	if a.db != nil {
		a.closer.Add("DB Close", a.shutdownDBStore)
	}

	go func() {
		if err := a.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.log.Error("Start server", zap.Error(err))
			a.stop()
		}
	}()
	a.log.Info("Server started")
	<-a.ctx.Done()
}

func (a *app) Stop() {
	a.log.Info("Shutting down server gracefully")

	// wait FileStoreInterval
	if err := a.eg.Wait(); err != nil {
		a.log.Error("Service", zap.Error(err))
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), constant.ServerShutdownTimeout*time.Second)
	defer cancel()

	if err := a.closer.Close(shutdownCtx); err != nil {
		a.log.Error("Shutdown", zap.Error(err), zap.Any("timeout: ", constant.ServerShutdownTimeout))
	}

	a.log.Info("Server stopped")

	_ = a.log.Sync()
}
