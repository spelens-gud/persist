package persist

import (
	"errors"
	"fmt"
	"sync"
)

var (
	gPersistMap     = make(map[string]IPersist)     // 所有注册的persist
	gPersistMapLazy = make(map[string]IPersist)     // 所有惰性注册的persist
	gPersistUserMap = make(map[string]IPersistUser) // 所有注册的用户相关persist
)

// IPersist 所有persist必须实现接口
type IPersist interface {
	Sync(wg *sync.WaitGroup) (err error)                       // 启动同步表结构
	Exit(wg *sync.WaitGroup)                                   // 退出
	Run() (err error)                                          // 启动
	Dead() bool                                                // 是否死亡
	PersistName() string                                       // 获取结构名
	RecoverBomb(bomb []byte) (err error)                       // 恢复数据通过 bomb数据
	SyncData(wg *sync.WaitGroup, sentryDebug bool) (err error) // 检查内存数据并同步到数据库
	RecoverTrace(trace [][]byte) (err error)                   // 恢复数据通过 trace数据
	StringToPersistSyncInterface(data string) any              // string类型数据 转化成 PersistSync结构
	BytesToPersistInterface(data []byte) any                   // bytes类型数据 转化成 Persist结构
	PersistInterfaceToBytes(i any) []byte                      // Persist结构 转化成 bytes类型数据
	PersistInterfaceToPkStruct(i any) any                      // Persist结构 转成 主键结构体interface
	LazyInit() (err error)                                     // 惰性创建注册初始化
	Segmentation(wg *sync.WaitGroup) (err error)               // 切换表名并创建新表
}

// IPersistUser 用户相关persist必须实现接口
type IPersistUser interface {
	IPersist
	Load(Uid int32) (err error)                           // 导入用户UID的数据
	Unload(Uid int32) (err error)                         // 导出用户UID的数据
	SetLoadState2Memory(Uid int32)                        // 将用户UID的数据强制设置为导入到内存中的状态
	LoadState(Uid int32) int32                            // 查询用户UID的导入状态
	SyncUserData(Uid int32, sentryDebug bool) (err error) // 检查用户UID的内存数据并同步到数据库
	PersistUserNilObjInterface() any                      // 获取PersistUser对象的nil指针
	PersistUserNilObjInterfaceList() any                  // 获取PersistUser对象数组的nil指针
}

// RegisterPersistLazy 惰性注册
func RegisterPersistLazy(persist IPersist) {
	name := persist.PersistName()
	if _, ok := gPersistMapLazy[name]; ok {
		panic(errors.New("repeated register lazy persist " + name))
	}
	gPersistMapLazy[name] = persist
}

// RegisterPersist 注册
func RegisterPersist(persist IPersist) {
	name := persist.PersistName()
	if _, ok := gPersistMap[name]; ok {
		panic(errors.New("repeated register persist " + name))
	}
	gPersistMap[name] = persist

	if persistUser, ok := persist.(IPersistUser); ok {
		if _, ok := gPersistUserMap[name]; ok {
			fmt.Printf("%#v\n", gPersistUserMap)
			panic(errors.New("repeated register persist user " + name))
		}
		gPersistUserMap[name] = persistUser
	}
}

// ChangeRegister 非注册 更换已经注册的
func ChangeRegister(persist IPersist) {
	name := persist.PersistName()
	if _, ok := gPersistMap[name]; ok {
		gPersistMap[name] = persist
	}
}

// GetIPersistByName 通过名字获取persist
func GetIPersistByName(name string) IPersist {
	if persist, ok := gPersistMap[name]; ok {
		return persist
	} else {
		return nil
	}
}

// Load 按照用户uid导入
func Load(uid int32) (err error) {
	for _, persist := range gPersistUserMap {
		err = persist.Load(uid)
		if err != nil {
			return errors.New(persist.PersistName() + err.Error())
		}
	}
	return
}

// SetLoadState2Memory 确定数据一致性前提下，强制设置用户数据已导入
func SetLoadState2Memory(uid int32) {
	for _, persist := range gPersistUserMap {
		persist.SetLoadState2Memory(uid)
	}
	return
}

// Unload 按照用户uid导出
func Unload(uid int32) (err error) {
	for _, persist := range gPersistUserMap {
		err = persist.Unload(uid)
		if err != nil {
			return errors.New(persist.PersistName() + err.Error())
		}
	}
	return
}

// LoadState 所有用户数据导入状态
func LoadState(uid int32) (stateList []int32) {
	for _, persist := range gPersistUserMap {
		stateList = append(stateList, persist.LoadState(uid))
	}
	return
}

// SyncPersist 所有Persist同步结构
func SyncPersist() error {
	for name, persist := range gPersistMapLazy {
		if err := persist.LazyInit(); err != nil {
			return err
		} else {
			delete(gPersistMapLazy, name)
		}
	}

	errMap := map[string]error{}

	var wg sync.WaitGroup
	for key := range gPersistMap {
		wg.Go(func() {
			err := gPersistMap[key].Sync(&wg)
			if err != nil {
				errMap[key] = err
				panic(errors.New(key + err.Error()))
			}
		})
	}
	wg.Wait()

	for _, err := range errMap {
		if err != nil {
			return err
		}
	}
	return nil
}

// RunPersist 运行所有Persist
func RunPersist() error {
	for _, persist := range gPersistMap {
		if err := persist.Run(); err != nil {
			return errors.New(persist.PersistName() + err.Error())
		}
	}
	return nil
}

// DeadPersist 是否存在异常状态Persist
func DeadPersist() bool {
	for _, persist := range gPersistMap {
		if dead := persist.Dead(); dead {
			return true
		}
	}
	return false
}

// ExitPersist 退出所有Persist
func ExitPersist() {
	var wg sync.WaitGroup
	for key := range gPersistMap {
		wg.Go(func() {
			gPersistMap[key].Exit(&wg)
		})
	}
	wg.Wait()
}

// SyncDataPersist 所有Persist, 不安全的方式强制同步数据, 调用后不允许再修改数据
func SyncDataPersist(sentryDebug bool) error {
	var wg sync.WaitGroup
	ch := make(chan error, len(gPersistMap))
	for name, persist := range gPersistMap {
		wg.Go(func() {
			err := persist.SyncData(&wg, sentryDebug)
			if err != nil {
				ch <- errors.New(name + err.Error())
			}
		})
	}
	wg.Wait()

	// TODO 后期加上sentry
	// sentry.Flush(time.Second * 5)
	select {
	case err, ok := <-ch:
		if ok {
			return err
		} else {
			return nil
		}
	default:
		return nil
	}
}

// SyncUserDataPersist 用户相关Persist, 不安全的方式强制同步数据, 调用后不允许再修改数据
func SyncUserDataPersist(uid int32, sentryDebug bool) (err error) {
	for _, persist := range gPersistUserMap {
		err = persist.SyncUserData(uid, sentryDebug)
		if err != nil {
			return errors.New(persist.PersistName() + err.Error())
		}
	}
	// TODO 后期加上sentry
	//sentry.Flush(time.Second * 5)
	return
}

// GetPersistList 注册的IPersist列表
func GetPersistList() (list []IPersist) {
	for _, persist := range gPersistMap {
		list = append(list, persist)
	}
	return
}

// GetPersistUserList 注册的IPersistUser列表
func GetPersistUserList() (list []IPersistUser) {
	for _, persist := range gPersistUserMap {
		list = append(list, persist)
	}
	return
}

// GetGPersistUserMap 获取 gPersistUserMap
func GetGPersistUserMap() map[string]IPersistUser {
	return gPersistUserMap
}

// SegmentationPersist 检查IPersist 配置切换写入表名
// 定时任务调用 实现切表
func SegmentationPersist() {
	var wg sync.WaitGroup
	ch := make(chan error, len(gPersistMap))
	for name, persist := range gPersistMap {
		wg.Go(func() {
			err := persist.Segmentation(&wg)
			if err != nil {
				ch <- errors.New(name + err.Error())
			}
		})
	}
	wg.Wait()
}
