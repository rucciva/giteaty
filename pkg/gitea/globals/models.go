package globals

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/models/migrations"
	"code.gitea.io/gitea/modules/setting"
	"github.com/rucciva/giteaty/pkg/gitea"
	"gopkg.in/ini.v1"
)

var (
	_ gitea.Models = gModels{}
)

type modelsOption func(*gModels)

func ModelsWithDBType(v string) modelsOption {
	return func(g *gModels) {
		g.cfg["DB_TYPE"] = v
	}
}

func ModelsWithDBHost(host string) modelsOption {
	return func(g *gModels) {
		g.cfg["HOST"] = host
	}
}
func ModelsWithDBName(v string) modelsOption {
	return func(g *gModels) {
		g.cfg["NAME"] = v
	}
}
func ModelsWithDBUser(v string) modelsOption {
	return func(g *gModels) {
		g.cfg["USER"] = v
	}
}
func ModelsWithDBPasswd(v string) modelsOption {
	return func(g *gModels) {
		g.cfg["PASSWD"] = v
	}
}
func ModelsWithDBSchema(v string) modelsOption {
	return func(g *gModels) {
		g.cfg["SCHEMA"] = v
	}
}
func ModelsWithDBSSLMode(v string) modelsOption {
	return func(g *gModels) {
		g.cfg["SSL_MODE"] = v
	}
}
func ModelsWithDBCharset(v string) modelsOption {
	return func(g *gModels) {
		g.cfg["CHARSET"] = v
	}
}
func ModelsWithDBPath(v string) modelsOption {
	return func(g *gModels) {
		g.cfg["PATH"] = v
	}
}
func ModelsWithDBSqliteTimeoutSecond(v int) modelsOption {
	return func(g *gModels) {
		g.cfg["SQLITE_TIMEOUT"] = strconv.FormatInt(int64(v), 10)
	}
}
func ModelsWithDBMaxIdleConns(v int) modelsOption {
	return func(g *gModels) {
		g.cfg["MAX_IDLE_CONNS"] = strconv.FormatInt(int64(v), 10)
	}
}
func ModelsWithDBMaxConnLifeTime(v time.Duration) modelsOption {
	return func(g *gModels) {
		g.cfg["CONN_MAX_LIFE_TIME"] = strconv.FormatInt(int64(v), 10)
	}
}
func ModelsWithDBMaxOpenConns(v int) modelsOption {
	return func(g *gModels) {
		g.cfg["MAX_OPEN_CONNS"] = strconv.FormatInt(int64(v), 10)
	}
}
func ModelsWithDBIterateBuffSize(v int) modelsOption {
	return func(g *gModels) {
		g.cfg["ITERATE_BUFFER_SIZE"] = strconv.FormatInt(int64(v), 10)
	}
}
func ModelsWithDBLogSql(v bool) modelsOption {
	return func(g *gModels) {
		g.cfg["LOG_SQL"] = strconv.FormatBool(v)
	}
}
func ModelsWithDBRetries(v int) modelsOption {
	return func(g *gModels) {
		g.cfg["DB_RETRIES"] = strconv.FormatInt(int64(v), 10)
	}
}
func ModelsWithDBRetryBackoff(v time.Duration) modelsOption {
	return func(g *gModels) {
		g.cfg["DB_RETRY_BACKOFF"] = strconv.FormatInt(int64(v), 10)
	}
}
func ModelsWithDBMigration() modelsOption {
	return func(g *gModels) {
		g.engineCreator = func() error {
			return models.NewEngine(context.Background(), migrations.Migrate)
		}
	}
}

type gModels struct {
	cfg map[string]string

	engineCreator func() error
}

var gmodels *gModels

func InitModels(opts ...modelsOption) (err error) {
	if gmodels != nil {
		return fmt.Errorf("already initialized")
	}

	gmodels = &gModels{
		cfg:           make(map[string]string),
		engineCreator: func() error { return models.SetEngine() },
	}
	for _, opt := range opts {
		opt(gmodels)
	}

	setting.Cfg = ini.Empty()
	for k, v := range gmodels.cfg {
		setting.Cfg.Section("database").Key(k).SetValue(v)
	}
	setting.InitDBConfig()
	if err := gmodels.engineCreator(); err != nil {
		return fmt.Errorf("set engine failed: %v", err)
	}
	return
}

func Models() gitea.Models {
	return gmodels
}

func (gModels) UserSignIn(username, password string) (*models.User, error) {
	return models.UserSignIn(username, password)
}

func (gModels) GetOrgsByUserID(userID int64, showAll bool) ([]*models.User, error) {
	return models.GetOrgsByUserID(userID, showAll)
}
func (gModels) GetUserTeams(userID int64, listOptions models.ListOptions) ([]*models.Team, error) {

	return models.GetUserTeams(userID, listOptions)
}
func (gModels) SearchUsers(opts *models.SearchUserOptions) (users []*models.User, count int64, err error) {
	return models.SearchUsers(opts)
}
