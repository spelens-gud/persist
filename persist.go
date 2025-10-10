package persist

import (
	"log"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"xorm.io/xorm"
)

var (
	gEngine    *xorm.Engine
	engineOnce sync.Once
)

// GetDatabaseDB 获取数据库连接
func GetDatabaseDB() *xorm.Engine {
	// TODO 初始化配置
	engineOnce.Do(func() {
		// TODO 初始化数据库
		var err error
		gEngine, err = xorm.NewEngine("mysql", "root:123456@tcp(127.0.0.1:3306)/persistence?charset=utf8mb4")
		if err != nil {
			log.Fatal(err)
		}

		gEngine.SetMaxIdleConns(2)            // 设置连接池中的保持连接的最大连接数
		gEngine.SetMaxOpenConns(4)            // 设置连接池的打开的最大连接数
		gEngine.SetConnMaxLifetime(time.Hour) // 设置连接超时时间
	})
	return gEngine
}

// ExitPersists 退出并保存所有持久化数据
func ExitPersists() error {
	ExitPersist()
	if err := SyncDataPersist(true); err != nil {
		return err
	}

	return nil
}

// InitPersists 初始化所有持久化数据
func InitPersists() error {
	engine := GetDatabaseDB()
	if engine == nil {
		panic("GetDB Error")
	}
	if err := SyncPersist(); err != nil {
		return err
	}

	return nil
}
