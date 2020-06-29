package ldaphandler

import (
	"fmt"
	"os"
	"testing"

	"github.com/rucciva/giteaty/pkg/gitea/globals"
)

func NoDB() bool {
	_, tok := os.LookupEnv("DB_TYPE")
	_, hok := os.LookupEnv("DB_HOST")
	return !tok || !hok
}

func TestMain(t *testing.M) {
	if NoDB() {
		os.Exit(t.Run())
		return
	}

	err := globals.InitModels(
		globals.ModelsWithDBType(os.Getenv("DB_TYPE")),
		globals.ModelsWithDBHost(os.Getenv("DB_HOST")),
		globals.ModelsWithDBName(os.Getenv("DB_NAME")),
		globals.ModelsWithDBUser(os.Getenv("DB_USER")),
		globals.ModelsWithDBPasswd(os.Getenv("DB_PASSWD")),
		globals.ModelsWithDBLogSql(false),
		globals.ModelsWithDBMigration(),
	)
	if err != nil {
		fmt.Printf("init db failed: %v", err)
		os.Exit(1)
	}

	os.Exit(t.Run())
}
