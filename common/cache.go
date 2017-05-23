package common

import (
	"github.com/astaxie/beego/cache"
)

var memcache cache.Cache

func GetMemoryCache() (c cache.Cache, err error) {
	if memcache == nil {
		memcache, err = cache.NewCache("memory", `{"interval":600}`)
	}
	return memcache, err
}
