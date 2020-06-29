package globals

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/ini.v1"
)

func TestModelsWithGiteaConf(t *testing.T) {
	g := &gModels{}
	err := ModelsWithGiteaConf(filepath.Join("testdata", "notexist"))(g)
	require.Error(t, err, "should return error when file not exist")

	cfg := ini.Empty().Section("database")
	g = &gModels{cfg: cfg}
	err = ModelsWithGiteaConf(filepath.Join("testdata", "gitea.ini"))(g)
	require.NoError(t, err, "should not return error")
	assert.Equal(t, cfg, g.cfg, "should not modify pointer")
	assert.Equal(t, "mysql", g.cfg.Key("DB_TYPE").String())
	assert.Equal(t, "/data/gitea/gitea.db", g.cfg.Key("PATH").String())
	assert.Equal(t, "mysql:3306", g.cfg.Key("HOST").String())
	assert.Equal(t, "gitea", g.cfg.Key("NAME").String())
	assert.Equal(t, "gitea", g.cfg.Key("USER").String())
	assert.Equal(t, "gitea", g.cfg.Key("PASSWD").String())
	assert.Equal(t, "disable", g.cfg.Key("SSL_MODE").String())
	assert.False(t, g.cfg.HasKey("LOG_SQL"))

	g = &gModels{cfg: cfg}
	_ = ModelsWithGiteaConf(filepath.Join("testdata", "gitea.ini"))(g)
	_ = ModelsWithDBLogSql(true)(g)
	assert.Equal(t, cfg, g.cfg, "should not modify pointer")
	assert.Equal(t, "mysql", g.cfg.Key("DB_TYPE").String())
	assert.Equal(t, "/data/gitea/gitea.db", g.cfg.Key("PATH").String())
	assert.Equal(t, "mysql:3306", g.cfg.Key("HOST").String())
	assert.Equal(t, "gitea", g.cfg.Key("NAME").String())
	assert.Equal(t, "gitea", g.cfg.Key("USER").String())
	assert.Equal(t, "gitea", g.cfg.Key("PASSWD").String())
	assert.Equal(t, "disable", g.cfg.Key("SSL_MODE").String())
	assert.Equal(t, true, g.cfg.Key("LOG_SQL").MustBool())

}
