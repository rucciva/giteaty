package command

import (
	"github.com/rucciva/giteaty/pkg/gitea/globals"
	"github.com/urfave/cli/v2"
)

const (
	flagDBType                = "db-type"
	flagDBHost                = "db-host"
	flagDBName                = "db-name"
	flagDBUser                = "db-user"
	flagDBPasswd              = "db-passwd"
	flagDBSchema              = "db-schema"
	flagDBSslMode             = "db-ssl-mode"
	flagDBPath                = "db-path"
	flagDBLogSql              = "db-log-sql"
	flagDBCharset             = "db-charset"
	flagDBSqliteTimeoutSecond = "db-sqlite-timeout-second"
	flagDBConnectRetries      = "db-connect-retries"
	flagDBConnectBackoff      = "db-connect-backoff"
	flagDBMaxIdleConns        = "db-max-idle-conns"
	flagDBMaxOpenConns        = "db-max-open-conns"
	flagDBConnMaxLifetime     = "db-conn-max-lifetime"
	flagDBIterateBufferSize   = "db-iterate-buffer-size"
)

func modelsFlag() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     flagDBType,
			EnvVars:  []string{"DB_TYPE"},
			Required: true,
		},
		&cli.StringFlag{
			Name:    flagDBHost,
			EnvVars: []string{"DB_HOST"},
		},
		&cli.StringFlag{
			Name:    flagDBName,
			EnvVars: []string{"DB_NAME"},
		},
		&cli.StringFlag{
			Name:    flagDBUser,
			EnvVars: []string{"DB_USER"},
		},
		&cli.StringFlag{
			Name:    flagDBPasswd,
			EnvVars: []string{"DB_PASSWD"},
		},
		&cli.StringFlag{
			Name:    flagDBSchema,
			EnvVars: []string{"DB_SCHEMA"},
		},
		&cli.StringFlag{
			Name:    flagDBSslMode,
			EnvVars: []string{"DB_SSL_MODE"},
		},
		&cli.StringFlag{
			Name:    flagDBPath,
			EnvVars: []string{"DB_PATH"},
		},
		&cli.BoolFlag{
			Name:    flagDBLogSql,
			EnvVars: []string{"DB_LOG_SQL"},
			Value:   false,
		},
		&cli.StringFlag{
			Name:    flagDBCharset,
			EnvVars: []string{"DB_CHARSET"},
		},
		&cli.IntFlag{
			Name:    flagDBSqliteTimeoutSecond,
			EnvVars: []string{"DB_SQLITE_TIMEOUT_SECOND"},
		},
		&cli.IntFlag{
			Name:    flagDBConnectRetries,
			EnvVars: []string{"DB_CONNECT_RETRIES"},
		},
		&cli.DurationFlag{
			Name:    flagDBConnectBackoff,
			EnvVars: []string{"DB_CONNECT_BACKOFF"},
		},
		&cli.IntFlag{
			Name:    flagDBMaxIdleConns,
			EnvVars: []string{"DB_MAX_IDLE_CONNS"},
		},
		&cli.IntFlag{
			Name:    flagDBMaxOpenConns,
			EnvVars: []string{"DB_MAX_OPEN_CONNS"},
		},
		&cli.DurationFlag{
			Name:    flagDBConnMaxLifetime,
			EnvVars: []string{"DB_CONN_MAX_LIFETIME"},
		},
		&cli.IntFlag{
			Name:    flagDBIterateBufferSize,
			EnvVars: []string{"DB_ITERATE_BUFFER_SIZE"},
		},
	}
}

func initDB(c *cli.Context) (err error) {
	opts := globals.ModelsOptions()
	if c.IsSet(flagDBType) {
		opts = append(opts, globals.ModelsWithDBType(c.String(flagDBType)))
	}
	if c.IsSet(flagDBHost) {
		opts = append(opts, globals.ModelsWithDBHost(c.String(flagDBHost)))
	}
	if c.IsSet(flagDBName) {
		opts = append(opts, globals.ModelsWithDBName(c.String(flagDBName)))
	}
	if c.IsSet(flagDBUser) {
		opts = append(opts, globals.ModelsWithDBUser(c.String(flagDBUser)))
	}
	if c.IsSet(flagDBPasswd) {
		opts = append(opts, globals.ModelsWithDBPasswd(c.String(flagDBPasswd)))
	}
	if c.IsSet(flagDBSchema) {
		opts = append(opts, globals.ModelsWithDBSchema(c.String(flagDBSchema)))
	}
	if c.IsSet(flagDBSslMode) {
		opts = append(opts, globals.ModelsWithDBSSLMode(c.String(flagDBSslMode)))
	}
	if c.IsSet(flagDBPath) {
		opts = append(opts, globals.ModelsWithDBPath(c.String(flagDBPath)))
	}
	opts = append(opts, globals.ModelsWithDBLogSql(c.Bool(flagDBLogSql)))
	if c.IsSet(flagDBCharset) {
		opts = append(opts, globals.ModelsWithDBCharset(c.String(flagDBCharset)))
	}
	if c.IsSet(flagDBSqliteTimeoutSecond) {
		opts = append(opts, globals.ModelsWithDBSqliteTimeoutSecond(c.Int(flagDBSqliteTimeoutSecond)))
	}
	if c.IsSet(flagDBConnectRetries) {
		opts = append(opts, globals.ModelsWithDBConnectRetries(c.Int(flagDBConnectRetries)))
	}
	if c.IsSet(flagDBConnectBackoff) {
		opts = append(opts, globals.ModelsWithDBConnectBackoff(c.Duration(flagDBConnectBackoff)))
	}
	if c.IsSet(flagDBMaxIdleConns) {
		opts = append(opts, globals.ModelsWithDBMaxIdleConns(c.Int(flagDBMaxIdleConns)))
	}
	if c.IsSet(flagDBMaxOpenConns) {
		opts = append(opts, globals.ModelsWithDBMaxOpenConns(c.Int(flagDBMaxOpenConns)))
	}
	if c.IsSet(flagDBConnMaxLifetime) {
		opts = append(opts, globals.ModelsWithDBConnMaxLifeTime(c.Duration(flagDBConnMaxLifetime)))
	}
	if c.IsSet(flagDBIterateBufferSize) {
		opts = append(opts, globals.ModelsWithDBIterateBufferSize(c.Int(flagDBIterateBufferSize)))
	}
	return globals.InitModels(opts...)
}
