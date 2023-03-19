package db

// (TODO): Update this as and when required
type db interface {
	Put(table string, key string, value any) error
	Get(table string, key string) (any, error)
}
