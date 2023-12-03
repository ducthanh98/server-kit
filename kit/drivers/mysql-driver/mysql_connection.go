package mysql_driver

import (
	"fmt"
	"github.com/ducthanh98/server-kit/kit/logger"
	"strings"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MysqlDriver struct {
}

type MysqlConfiguration struct {
	Port      int
	Host      string
	Username  string
	Password  string
	Database  string
	Charset   string
	ParseTime bool
	Loc       string
	Debug     bool
}

func DefaultMysqlFromConfig(opts *MysqlConfiguration) *MysqlConfiguration {
	if opts == nil {
		opts = &MysqlConfiguration{
			Host:      viper.GetString("mysql.host"),
			Port:      viper.GetInt("mysql.port"),
			Username:  viper.GetString("mysql.username"),
			Password:  viper.GetString("mysql.password"),
			Charset:   viper.GetString("mysql.charset"),
			Database:  viper.GetString("mysql.database"),
			ParseTime: viper.GetBool("mysql.parse_time"),
			Debug:     viper.GetBool("mysql.debug"),
		}
	}

	return opts
}

func (d MysqlDriver) NewConnection(conn interface{}) (*gorm.DB, error) {
	opts := DefaultMysqlFromConfig(conn.(*MysqlConfiguration))
	logger.Log.Info("PARAM: ", buildMysqlConnectionParam(opts))
	db, err := gorm.Open(mysql.Open(buildMysqlConnectionParam(opts)), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if viper.GetBool("mysql.debug") {
		db.Debug()
	}

	return db, nil
}

func buildMysqlConnectionParam(opts *MysqlConfiguration) string {
	param := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s",
		opts.Username,
		opts.Password,
		opts.Host,
		opts.Port,
		opts.Database,
	)

	// Additional params
	additionParams := make(map[string]string)

	if len(opts.Charset) > 0 {
		additionParams["charset"] = opts.Charset
	}

	switch opts.ParseTime {
	case true:
		additionParams["parseTime"] = "True"
	case false:
		additionParams["parseTime"] = "False"
	}

	if len(opts.Loc) > 0 {
		additionParams["loc"] = opts.Loc
	}

	additionParamsStr := make([]string, 0)
	for k, v := range additionParams {
		additionParamsStr = append(additionParamsStr, fmt.Sprintf("%s=%s", k, v))
	}

	return fmt.Sprintf("%s?%s", param, strings.Join(additionParamsStr, "&"))
}
