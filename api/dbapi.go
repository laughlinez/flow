package api

type IDBReadAPI interface {
	Keys(prefix string) ([]string, error)
	Get(key string) (interface{}, error)
}

type IDBWriteAPI interface {
	Put(key string, value interface{}) error
}

type IDBReadWriteAPI interface {
	IDBReadAPI
	IDBWriteAPI
}
