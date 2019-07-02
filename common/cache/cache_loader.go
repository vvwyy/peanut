package cache

type Loader interface {
	Load(key interface{}) (interface{}, error)
	//LoadAll(keys []interface{}) (map[string]interface{}, error)
}
