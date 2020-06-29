package globals

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModelsWithGiteaConf(t *testing.T) {
	g := &gModels{}

	err := ModelsWithGiteaConf(filepath.Join("testdata", "notexist"))(g)
	require.Error(t, err, "should return error when file not exist")

	err = ModelsWithGiteaConf(filepath.Join("testdata", "gitea.ini"))(g)
	require.NoError(t, err, "should not return error")
	assert.Equal(t, "mysql", g.cfg.Key("DB_TYPE").String())
	assert.Equal(t, "/data/gitea/gitea.db", g.cfg.Key("PATH").String())
	assert.Equal(t, "mysql:3306", g.cfg.Key("HOST").String())
	assert.Equal(t, "gitea", g.cfg.Key("NAME").String())
	assert.Equal(t, "gitea", g.cfg.Key("USER").String())
	assert.Equal(t, "gitea", g.cfg.Key("PASSWD").String())
	assert.Equal(t, "disable", g.cfg.Key("SSL_MODE").String())
	assert.False(t, g.cfg.HasKey("LOG_SQL"))
}
