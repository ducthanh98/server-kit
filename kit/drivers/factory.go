package drivers

import (
	"github.com/ducthanh98/server-kit/kit/constants"
	mysqldriver "github.com/ducthanh98/server-kit/kit/drivers/mysql-driver"
	postgresqldriver "github.com/ducthanh98/server-kit/kit/drivers/postgresql-driver"
	"github.com/ducthanh98/server-kit/kit/utils/types"
)

func InitConnection(driver types.Driver) SQLDriver {
	switch driver {
	case constants.MysqlDriver:
		return &mysqldriver.MysqlDriver{}
	case constants.PostgresqlDriver:
		return &postgresqldriver.PostgresqlDriver{}
	}

	return nil
}
