package drivers

import "gorm.io/gorm"

type SQLDriver interface {
	NewConnection() (*gorm.DB, error)
}

type Queue interface {
}
