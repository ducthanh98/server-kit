package cache

type Callback func() (interface{}, error)

type Cache interface {
	Get(key string, callback Callback) (interface{}, error)
	Set(key string, value interface{})
	Delete(key string)
}
