package cache

type (
	Cache interface {
		Get(key string) (string, bool)
		Set(key string, value any) error
		DeleteByPattern(key string) error
	}
)
