package storage

import (
	"github.com/go-redis/redis"
	"github.com/gtfierro/hod/config"
	"github.com/pkg/errors"
	//logrus "github.com/sirupsen/logrus"
)

type RedisStorageProvider struct {
	*redis.Client
	name string
}

func (rds *RedisStorageProvider) Initialize(name string, cfg *config.Config) (err error) {
	rds.Client = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	rds.name = name

	if _, err := rds.Ping().Result(); err != nil {
		return err
	}

	return nil
}

func (rds *RedisStorageProvider) OpenTransaction() (Transaction, error) {
	return rds, nil
}

func (rtx *RedisStorageProvider) Has(bucket HodNamespace, key []byte) (exists bool, err error) {
	cmd := rtx.Exists(string(rtx.name) + string(bucket) + string(key))
	if cmd == nil {
		return false, errors.New("Cannot run command")
	}
	if cmd.Err() != nil {
		return false, cmd.Err()
	}
	return cmd.Val() == 1, nil
}

func (rtx *RedisStorageProvider) Get(bucket HodNamespace, key []byte) (value []byte, err error) {
	//logrus.Infof("GET %s %+v %s", bucket, key, string(key))
	cmd := rtx.Client.Get(string(rtx.name) + string(bucket) + string(key))
	if cmd.Err() == redis.Nil {
		return nil, ErrNotFound
	} else if cmd.Err() != nil {
		return nil, err
	}
	//logrus.Infof("Get> %s", cmd.String())
	bytes, err := cmd.Bytes()
	if len(bytes) == 0 {
		return nil, ErrNotFound
	}
	return bytes, nil
}

func (rtx *RedisStorageProvider) Put(bucket HodNamespace, key, value []byte) (err error) {
	val := rtx.Set(string(rtx.name)+string(bucket)+string(key), string(value), 0)
	//logrus.Warning("set", val.String())
	return val.Err()
}

func (rtx *RedisStorageProvider) Iterate(bucket HodNamespace) Iterator {
	return nil
}

func (rtx *RedisStorageProvider) Release() {
}

func (rtx *RedisStorageProvider) Commit() error {
	return nil
}

func (rds *RedisStorageProvider) OpenSnapshot() (Traversable, error) {
	return rds, nil
}
