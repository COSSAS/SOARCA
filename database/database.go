package database

type Database interface {
	Read(string) (any, error)
	Find(map[string]string, ...interface{}) ([]any, error)
	Create(interface{}) error
	Update(string, interface{}) error
	Delete(string) error
}
type FindOptions interface {
	GetIds() interface{}
}
