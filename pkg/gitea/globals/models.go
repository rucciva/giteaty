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
	"github.com/unknwon/com"
	"gopkg.in/ini.v1"
)

var (
	_ gitea.Models = gModels{}
)

type modelsOption func(*gModels) error

func ModelsOptions() []modelsOption {
	return make([]modelsOption, 0, 18)
}

func ModelsWithDBType(v string) modelsOption {
	return func(g *gModels) (err error) {
		g.cfg.Key("DB_TYPE").SetValue(v)
		return
	}
}

func ModelsWithDBHost(host string) modelsOption {
	return func(g *gModels) (err error) {
		g.cfg.Key("HOST").SetValue(host)
		return
	}
}
func ModelsWithDBName(v string) modelsOption {
	return func(g *gModels) (err error) {
		g.cfg.Key("NAME").SetValue(v)
		return
	}
}
func ModelsWithDBUser(v string) modelsOption {
	return func(g *gModels) (err error) {
		g.cfg.Key("USER").SetValue(v)
		return
	}
}
func ModelsWithDBPasswd(v string) modelsOption {
	return func(g *gModels) (err error) {
		g.cfg.Key("PASSWD").SetValue(v)
		return
	}
}
func ModelsWithDBSchema(v string) modelsOption {
	return func(g *gModels) (err error) {
		g.cfg.Key("SCHEMA").SetValue(v)
		return
	}
}
func ModelsWithDBSSLMode(v string) modelsOption {
	return func(g *gModels) (err error) {
		g.cfg.Key("SSL_MODE").SetValue(v)
		return
	}
}
func ModelsWithDBCharset(v string) modelsOption {
	return func(g *gModels) (err error) {
		g.cfg.Key("CHARSET").SetValue(v)
		return
	}
}
func ModelsWithDBPath(v string) modelsOption {
	return func(g *gModels) (err error) {
		g.cfg.Key("PATH").SetValue(v)
		return
	}
}
func ModelsWithDBSqliteTimeoutSecond(v int) modelsOption {
	return func(g *gModels) (err error) {
		g.cfg.Key("SQLITE_TIMEOUT").SetValue(strconv.FormatInt(int64(v), 10))
		return
	}
}
func ModelsWithDBMaxIdleConns(v int) modelsOption {
	return func(g *gModels) (err error) {
		g.cfg.Key("MAX_IDLE_CONNS").SetValue(strconv.FormatInt(int64(v), 10))
		return
	}
}
func ModelsWithDBConnMaxLifeTime(v time.Duration) modelsOption {
	return func(g *gModels) (err error) {
		g.cfg.Key("CONN_MAX_LIFE_TIME").SetValue(strconv.FormatInt(int64(v), 10))
		return
	}
}
func ModelsWithDBMaxOpenConns(v int) modelsOption {
	return func(g *gModels) (err error) {
		g.cfg.Key("MAX_OPEN_CONNS").SetValue(strconv.FormatInt(int64(v), 10))
		return
	}
}
func ModelsWithDBIterateBufferSize(v int) modelsOption {
	return func(g *gModels) (err error) {
		g.cfg.Key("ITERATE_BUFFER_SIZE").SetValue(strconv.FormatInt(int64(v), 10))
		return
	}
}
func ModelsWithDBLogSql(v bool) modelsOption {
	return func(g *gModels) (err error) {
		g.cfg.Key("LOG_SQL").SetValue(strconv.FormatBool(v))
		return
	}
}
func ModelsWithDBConnectRetries(v int) modelsOption {
	return func(g *gModels) (err error) {
		g.cfg.Key("DB_RETRIES").SetValue(strconv.FormatInt(int64(v), 10))
		return
	}
}

func ModelsWithDBConnectBackoff(v time.Duration) modelsOption {
	return func(g *gModels) (err error) {
		g.cfg.Key("DB_RETRY_BACKOFF").SetValue(strconv.FormatInt(int64(v), 10))
		return
	}
}

func ModelsWithGiteaConf(path string) modelsOption {
	return func(g *gModels) (err error) {
		cfg := ini.Empty()
		if !com.IsFile(path) {
			return fmt.Errorf("invalid path %s", path)
		}
		if err = cfg.Append(path); err != nil {
			return
		}
		g.cfg = cfg.Section("database")
		return
	}
}

func ModelsWithDBMigration() modelsOption {
	return func(g *gModels) (err error) {
		g.engineCreator = func() error {
			return models.NewEngine(context.Background(), migrations.Migrate)
		}
		return
	}
}

type gModels struct {
	cfg *ini.Section

	engineCreator func() error
}

var gm *gModels

func InitModels(opts ...modelsOption) (err error) {
	if gm != nil {
		return fmt.Errorf("already initialized")
	}

	setting.Cfg = ini.Empty()
	gm = &gModels{
		cfg:           setting.Cfg.Section("database"),
		engineCreator: models.SetEngine,
	}
	for _, opt := range opts {
		if err = opt(gm); err != nil {
			return
		}
	}
	setting.InitDBConfig()
	if err := gm.engineCreator(); err != nil {
		return fmt.Errorf("set engine failed: %v", err)
	}
	return
}

func Models() gitea.Models {
	return gm
}

func (gModels) UserSignIn(username, password string) (*models.User, error) {
	return models.UserSignIn(username, password)
}

func (gModels) GetUserTeams(userID int64, listOptions models.ListOptions) ([]*models.Team, error) {
	return models.GetUserTeams(userID, listOptions)
}

func (gModels) SearchUsers(opts *models.SearchUserOptions) (users []*models.User, count int64, err error) {
	return models.SearchUsers(opts)
}
