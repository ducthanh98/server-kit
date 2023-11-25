package postgresql_driver

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"strings"
)

type PostgresqlDriver struct {
}

type postgresqlConfiguration struct {
	Port     int
	Host     string
	Username string
	Password string
	Database string
	Debug    bool
}

func DefaultPostgresqlFromConfig(opts *postgresqlConfiguration) *postgresqlConfiguration {
	if opts == nil {
		opts = &postgresqlConfiguration{
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

func (d PostgresqlDriver) NewConnection() (*gorm.DB, error) {
	opts := DefaultPostgresqlFromConfig(nil)
	log.Info("PARAM: ", buildPostgresConnectionParam(opts))
	db, err := gorm.Open(postgres.Open(buildPostgresConnectionParam(opts)), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if opts.Debug {
		db.Debug()
	}

	return db, nil
}

func buildPostgresConnectionParam(opts *postgresqlConfiguration) string {
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
