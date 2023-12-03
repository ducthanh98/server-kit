package drivers

import (
	"gorm.io/gorm"
)

type SQLDriver interface {
	NewConnection(conn interface{}) (*gorm.DB, error)
}
