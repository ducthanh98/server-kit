package postgresql_driver

import (
	"fmt"
	"github.com/ducthanh98/server-kit/kit/logger"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"strings"
)

type PostgresqlDriver struct {
}

type PostgresqlConfiguration struct {
	Port     int
	Host     string
	Username string
	Password string
	Database string
	Debug    bool
}

func DefaultPostgresqlFromConfig(con interface{}) *PostgresqlConfiguration {
	var opts *PostgresqlConfiguration
	if con != nil {
		opts = con.(*PostgresqlConfiguration)
	}
	if opts == nil {
		opts = &PostgresqlConfiguration{
			Host:     viper.GetString("postgresql.host"),
			Port:     viper.GetInt("postgresql.port"),
			Username: viper.GetString("postgresql.username"),
			Password: viper.GetString("postgresql.password"),
			Database: viper.GetString("postgresql.database"),
			Debug:    viper.GetBool("postgresql.debug"),
		}
	}

	return opts
}

func (d PostgresqlDriver) NewConnection(con interface{}) (*gorm.DB, error) {
	opts := DefaultPostgresqlFromConfig(con)
	logger.Log.Info("PARAM: ", buildPostgresConnectionParam(opts))
	db, err := gorm.Open(postgres.Open(buildPostgresConnectionParam(opts)), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if opts.Debug {
		db.Debug()
	}

	return db, nil
}

func buildPostgresConnectionParam(opts *PostgresqlConfiguration) string {
	param := fmt.Sprintf(
		"postgresql://%v:%v@%v:%v/%v",
		opts.Username,
		opts.Password,
		opts.Host,
		opts.Port,
		opts.Database,
	)

	// Additional params
	additionParams := make(map[string]string)

	additionParamsStr := make([]string, 0)
	for k, v := range additionParams {
		additionParamsStr = append(additionParamsStr, fmt.Sprintf("%s=%s", k, v))
	}

	return fmt.Sprintf("%s?%s", param, strings.Join(additionParamsStr, "&"))
}
