package persist

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"os"
	"reflect"
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"xorm.io/xorm"
)

var (
	GlobalDBFiledMap []string // 全局数据库字段映射
)

type GlobalSync[T any] struct {
	Data   T               // 数据
	Op     int8            // 操作类型
	BitSet GlobalBitSet[T] // 位图
}

// GlobalManager 全局管理器
type GlobalManager[T any] struct {
	managerState int32 // 管理器状态
	loadState    int32 // 加载状态

	pool     *sync.Pool // 对象池
	modelNil *T

	syncQueue  *[]*GlobalSync[T] // 同步队列
	cacheQueue *[]*GlobalSync[T] // 缓存队列

	FailQueue   []*GlobalSync[T] // 失败队列
	InsertQueue []*GlobalSync[T] // 插入队列

	lastWriteBackTime time.Duration // 上次写回时间

	syncChan  chan *GlobalSync[T] // 同步通道
	syncBegin chan bool           // 同步开始
	syncEnd   chan bool           // 同步结束
	exitBegin chan bool           // 退出开始
	exitEnd   chan bool           // 退出结束

	//hashAuthId     GlobalHashAuthId
	//hashAuthIdType GlobalHashAuthIdType
	bitSetAll GlobalBitSet[T]

	engine *xorm.Engine // TODO 后期支持多种ORM数据库
}

// NewGlobalManager 创建全局管理器
func NewGlobalManager[T any](engine *xorm.Engine) *GlobalManager[T] {
	m := &GlobalManager[T]{
		engine: engine,
	}

	m.modelNil = new(T)
	m.syncChan = make(chan *GlobalSync[T], runtime.NumCPU()*2)
	tmpSyncQueue := make([]*GlobalSync[T], 0)
	m.syncQueue = &tmpSyncQueue
	m.syncEnd = make(chan bool)
	m.syncBegin = make(chan bool)
	m.exitBegin = make(chan bool)
	m.exitEnd = make(chan bool)
	tmpCacheQueue := make([]*GlobalSync[T], 0)
	m.cacheQueue = &tmpCacheQueue
	m.lastWriteBackTime = 1 * time.Millisecond
	m.pool = &sync.Pool{
		New: func() any {
			return m.modelNil
		},
	}

	m.bitSetAll = InitGlobalBitSet[T]()
	m.bitSetAll.SetAll()
	if engine != nil {
		fieldNames := GetFieldNames(m.modelNil)
		GlobalDBFiledMap = make([]string, len(fieldNames))
		for idx, name := range fieldNames {
			GlobalDBFiledMap[idx] = engine.GetColumnMapper().Obj2Table(name)
		}
	}

	return m
}

func (g *GlobalManager[T]) Sync(wg *sync.WaitGroup) (err error) {
	//TODO implement me
	panic("implement me")
}

func (g *GlobalManager[T]) Exit(wg *sync.WaitGroup) {
	//TODO implement me
	panic("implement me")
}

// Run 启动管理器
func (g *GlobalManager[T]) Run() (err error) {
	if atomic.CompareAndSwapInt32(&g.managerState, EGlobalManagerStateIdle, EGlobalManagerStateNormal) {
		// 启动失败读取崩溃恢复
		if err = g.LoadFile(); err != nil {
			return err
		}
		go g.Collect() // 启动数据收集协程
	} else if atomic.CompareAndSwapInt32(&g.managerState, EGlobalManagerStatePanic, EGlobalManagerStateNormal) {
		// 从崩溃状态恢复
		if err = g.LoadFile(); err != nil {
			return err
		}
		go g.Collect() // 启动数据收集协程
	}
	return nil
}

// Dead 管理器是否不可用
func (g *GlobalManager[T]) Dead() bool {
	return atomic.LoadInt32(&g.managerState) != EGlobalManagerStateNormal
}

// PersistName 获取持久化名称
func (g *GlobalManager[T]) PersistName() string {
	return reflect.TypeOf(g.modelNil).Name()
}

func (g *GlobalManager[T]) RecoverBomb(bomb []byte) (err error) {
	//TODO implement me
	panic("implement me")
}

func (g *GlobalManager[T]) SyncData(wg *sync.WaitGroup, sentryDebug bool) (err error) {
	//TODO implement me
	panic("implement me")
}

func (g *GlobalManager[T]) RecoverTrace(trace [][]byte) (err error) {
	//TODO implement me
	panic("implement me")
}

func (g *GlobalManager[T]) StringToPersistSyncInterface(data string) any {
	//TODO implement me
	panic("implement me")
}

// BytesToPersistInterface bytes转化为persist
func (g *GlobalManager[T]) BytesToPersistInterface(data []byte) any {
	return g.BytesToPersist(data)
}

// PersistInterfaceToBytes persist转化为bytes
func (g *GlobalManager[T]) PersistInterfaceToBytes(i any) []byte {
	return g.PersistToBytes(i.(*T), g.bitSetAll)
}

// PersistInterfaceToPkStruct persist转化为主键
func (g *GlobalManager[T]) PersistInterfaceToPkStruct(i any) any {
	cls, ok := i.(*T)
	if !ok {
		return nil
	}

	pk, ok := GetFieldValueByTag(cls, "xorm", "pk")
	if !ok {
		return nil
	}

	return pk
}

func (g *GlobalManager[T]) LazyInit() (err error) {
	//TODO implement me
	panic("implement me")
}

func (g *GlobalManager[T]) Segmentation(wg *sync.WaitGroup) (err error) {
	//TODO implement me
	panic("implement me")
}

// PersistUserNilObjInterface 获取PersistUser对象数组的nil指针
func (g *GlobalManager[T]) PersistUserNilObjInterface() any {
	return g.modelNil
}

// PersistUserNilObjInterfaceList 返回persist any list
func (g *GlobalManager[T]) PersistUserNilObjInterfaceList() any {
	plist := make([]*T, 0)
	return &plist
}

// Collect 收集数据
func (g *GlobalManager[T]) Collect() {
	var persistSync *GlobalSync[T]
	var ok bool
	// 0:normal  1:exit begin, save sync  2:save cache  3:save done
	var state int8
	go g.Save()
	g.syncBegin <- true
	for {
		select {
		case persistSync, ok = <-g.syncChan:
			if ok {
				*g.cacheQueue = append(*g.cacheQueue, persistSync)
			}
		case _, ok = <-g.syncEnd:
			if ok {
				g.CheckOverload()
				g.cacheQueue, g.syncQueue = g.syncQueue, g.cacheQueue
				switch state {
				case EGlobalCollectStateNormal:
					//go m.AsyncSave()
					g.syncBegin <- true
				case EGlobalCollectStateSaveSync:
					//go m.AsyncSave()
					g.syncBegin <- true
					state = EGlobalCollectStateSaveCache
				case EGlobalCollectStateSaveCache:
					//go m.AsyncSave()
					g.syncBegin <- true
					state = EGlobalCollectStateSaveDone
				case EGlobalCollectStateSaveDone:
					g.syncBegin <- false
					<-g.syncEnd
					g.exitEnd <- true
					return
				}
			}
		case _, ok = <-g.exitBegin:
			if ok {
				state = EGlobalCollectStateSaveSync
			}
		}
	}
}

// LoadFile 文件读取写回失败数据
func (g *GlobalManager[T]) LoadFile() error {
	// TODO 指定路径恢复
	bombExist := DirExists("_./_Users_xt_go_servers_admin_backend_data/MenusGlobal.bomb")
	tmpExist := DirExists("_./_Users_xt_go_servers_admin_backend_data/MenusGlobal.tmp")

	if tmpExist {
		return EPersistErrorTempFileExist
	}

	if bombExist {
		data, err := os.ReadFile("_./_Users_xt_go_servers_admin_backend_data/MenusGlobal.bomb")
		if err != nil {
			return err
		}
		pos := bytes.IndexByte(data, byte(' '))
		if pos == -1 {
			return EPersistErrorInvalidBombFile
		}
		persistData := data[pos+1:]
		err = m.UnmarshalFailQueue(persistData, &g.FailQueue)
		if err != nil {
			return err
		}

		session := g.engine.NewSession()
		defer session.Close()

		var persistSync *GlobalSync[T]

		for i := range g.FailQueue {
			persistSync = g.FailQueue[i]
			err = g.SaveDB(session, persistSync)
			if err != nil {
				g.FailQueue = g.FailQueue[i:]
				g.SaveFile()
				return err
			}
		}
		g.FailQueue = g.FailQueue[0:0]
		g.RemoveFile()

	}
	return nil
}

// SaveDB xorm写数据库
func (g *GlobalManager[T]) SaveDB(session *xorm.Session, persistSync *GlobalSync[T]) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if err == nil {
				err = errors.New("unknown error")
			}
		}
	}()
	switch persistSync.Op {
	case EGlobalOpInsert:

		_, err = session.Insert(persistSync.Data)

		if err != nil {
			log.Infoln("insert error ", err, "[sql error MenusGlobal]", g.PersistSyncToString(persistSync))
			return
		}

	case EGlobalOpUpdate:
		cls := persistSync.Data
		bitSet := persistSync.BitSet
		if bitSet.IsSetAll() {
			_, err = session.ID(core.NewPK(cls.AuthId)).AllCols().Update(cls)
			if err != nil {
				log.Infoln("update error ", err, "[sql error MenusGlobal]", m.PersistSyncToString(persistSync))
				return
			}
		} else {
			var nameList []string
			for idx, name := range MenusGlobalDBFiledMap {
				if bitSet.Get(MenusGlobalFieldIndex(idx)) {
					nameList = append(nameList, name)
				}
			}
			if nameList != nil {
				_, err = session.ID(core.NewPK(cls.AuthId)).Cols(nameList...).Update(cls)
				if err != nil {
					log.Infoln("update error ", err, "[sql error MenusGlobal]", m.PersistSyncToString(persistSync))
					return
				}
			} else {
				_, err = session.ID(core.NewPK(cls.AuthId)).AllCols().Update(cls)
				if err != nil {
					log.Infoln("update error ", err, "[sql error MenusGlobal]", m.PersistSyncToString(persistSync))
					return
				}
			}
		}

	case EGlobalOpDelete:
		cls := persistSync.Data
		_, err = session.ID(core.NewPK(cls.AuthId)).Delete(gMenusGlobalNil)
		if err != nil {
			log.Infoln("delete error ", err, "[sql error MenusGlobal]", m.PersistSyncToString(persistSync))
			return
		}

	}
	return
}

// PersistSyncToString 序列化2sync
func (g *GlobalManager[T]) PersistSyncToString(persistSync *GlobalSync[T]) (data string) {
	buf := g.PersistSyncToBytes(persistSync)
	if buf == nil {
		return ""
	}
	data = base64.StdEncoding.EncodeToString(buf)
	return
}

// PersistSyncToBytes 序列化sync
func (g *GlobalManager[T]) PersistSyncToBytes(persistSync *GlobalSync[T]) (data []byte) {
	var err error
	if persistSync == nil {
		return nil
	}
	defer func() {
		if r := recover(); r != nil {
			log.Infoln("recovered in ", r)
			log.Infoln("stack: ", string(debug.Stack()))
		}
		if err != nil {
			log.Infoln("PersistSyncToBytes Error", err.Error())
		}
	}()
	size := 0

	var bitSetSize = (int)((GlobalFieldIndex(GetFieldNum(g.modelNil))>>EGlobalLog2WordSize)+1) * 8

	pData := g.PersistToBytes(persistSync.Data, persistSync.BitSet)
	size += len(pData) + 1 + bitSetSize

	data = make([]byte, size)

	i := 0

	copy(data[i:], pData)
	i += len(pData)
	data[i] = uint8(persistSync.Op)
	i += 1
	for _, setItem := range persistSync.BitSet.set {
		binary.LittleEndian.PutUint64(data[i:], setItem)
		i += 8
	}

	return
}

// PersistToBytes 序列化
func (g *GlobalManager[T]) PersistToBytes(cls *T, bitSet GlobalBitSet[T]) (data []byte) {
	var err error
	if cls == nil {
		return nil
	}
	defer func() {
		if r := recover(); r != nil {
		}
		if err != nil {
			log.Infoln("PersistToBytes Error", err.Error())
		}
	}()
	size := 0

	return nil
}

// BytesToPersist 反序列化
func (g *GlobalManager[T]) BytesToPersist(data []byte) (cls *T) {
	cls = g.modelNil
	return cls
}

// Save 异步写回
func (g *GlobalManager[T]) Save() {
	var exit bool
	for {
		// 正常退出
		exit = g.AsyncSave()
		if exit {
			break
		}
	}
}

// AsyncSave 异步写回
func (g *GlobalManager[T]) AsyncSave() (exit bool) {
	var persistSync *GlobalSync[T]
	var err error
	var queueEmpty bool
	bTime := time.Now().UnixNano()
	defer func() {
		if r := recover(); r != nil {
			//log.Infoln("recovered in ", r)
			//log.Infoln("stack: ", string(debug.Stack()))
			if !queueEmpty {
				//log.Infoln("save failed: incrementalSave")
			}
		} else {
			if !queueEmpty {
				if err == nil {
					//log.Infoln("save success: incrementalSave")
				} else {
					//log.Infoln("save failed: incrementalSave")
				}
			}
		}
		g.DataToFailQueue()
		g.lastWriteBackTime = time.Duration(time.Now().UnixNano() - bTime)
		g.syncEnd <- true
	}()

	needCollect := <-g.syncBegin
	if len(*g.syncQueue) == 0 {
		queueEmpty = true
		if needCollect {
			time.Sleep(time.Millisecond * 100)
		} else {
			exit = true
		}
		return
	}
	session := g.engine.NewSession()
	defer session.Close()

	//log.Infoln("begin incrementalSave", bTime)

	if len(g.FailQueue) > 0 {
		tmpQueue := make([]*GlobalSync[T], len(g.FailQueue)+len(*g.syncQueue))
		copy(tmpQueue, g.FailQueue)
		copy(tmpQueue[len(g.FailQueue):], *g.syncQueue)
		insertQueue, otherQueue := g.MergeQueue(tmpQueue, true)
		g.syncQueue = &otherQueue
		g.InsertQueue = insertQueue
		g.FailQueue = g.FailQueue[0:0]
	} else {
		insertQueue, otherQueue := g.MergeQueue(*g.syncQueue, true)
		g.syncQueue = &otherQueue
		g.InsertQueue = insertQueue
	}

	multiInsertFn := func() bool {
		var err error
		defer func() {
			if r := recover(); r != nil {
				_ = session.Rollback()
			} else {
				if err == nil {
					m.InsertQueue = m.InsertQueue[0:0]
				} else {
					_ = session.Rollback()
				}
			}
		}()

		if len(m.InsertQueue) <= 0 {
			return true
		}
		err = session.Begin()
		if err != nil {
			return false
		}

		const num = 100
		var insertArray [num]*model.MenusGlobal
		length := len(m.InsertQueue)
		quotient := length / num
		remainder := length % num
		for i := 0; i < quotient; i++ {
			//fmt.Println("queue->(", i*num, "-", (i+1)*num, "): ", m.InsertQueue[i*num:(i+1)*num])
			for j := 0; j < num; j++ {
				insertArray[j] = m.InsertQueue[i*num+j].Data
			}

			_, err = session.InsertMulti(insertArray[:])

			if err != nil {
				log.Infoln("InsertMulti error ", err)
				return false
			}
		}
		if remainder != 0 {
			//fmt.Println("queue->(", quotient*num, "-", length, "): ", m.InsertQueue[quotient*num:length])

			insertArray = [num]*model.MenusGlobal{}
			for j := 0; j < remainder; j++ {
				insertArray[j] = m.InsertQueue[quotient*num+j].Data
			}

			_, err = session.InsertMulti(insertArray[:remainder])

			if err != nil {
				log.Infoln("InsertMulti error ", err)
				return false
			}
		}
		err = session.Commit()
		if err != nil {
			return false
		}
		return true
	}

	multiInsertSuccess := multiInsertFn()

	// 批量插入失败, 改为单条插入
	if !multiInsertSuccess {
		for idx, persistSync := range g.InsertQueue {
			err = g.SaveDB(session, persistSync)
			if err != nil {
				g.InsertQueue = g.InsertQueue[idx:]
				g.SaveFile()
				return
			}
		}
		g.InsertQueue = m.InsertQueue[0:0]
	}

	for i := 0; i < len(*g.syncQueue); i++ {
		persistSync = (*m.syncQueue)[i]
		err = g.SaveDB(session, persistSync)
		if err != nil {
			*m.syncQueue = (*m.syncQueue)[i:]
			m.SaveFile()
			return
		}
	}
	*g.syncQueue = (*g.syncQueue)[0:0]
	g.RemoveFile()
	return
}

// DataToFailQueue 未写入成功数据, 添加到失败队列
func (g *GlobalManager[T]) DataToFailQueue() {
	var persistSync *GlobalSync[T]

	// 插入队列数据添加到失败队列
	g.FailQueue = append(g.FailQueue, g.InsertQueue...)
	// 清空插入队列
	g.InsertQueue = g.InsertQueue[0:0]

	// 一旦失败标记所有的数据都是失败, 不允许导出
	for i := 0; i < len(*g.syncQueue); i++ {
		persistSync = (*g.syncQueue)[i]
		switch persistSync.Op {
		case EGlobalOpInsert, EGlobalOpUpdate, EGlobalOpDelete:
			g.FailQueue = append(g.FailQueue, persistSync)
		default:
		}
	}
	// 清空同步队列
	*g.syncQueue = (*g.syncQueue)[0:0]
}
