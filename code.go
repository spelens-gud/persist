package persist

const ELoadPollingTimeOut = 5
const EMarshalFlagPoint uint8 = 0b00000001
const EMarshalFlagBitSet uint8 = 0b10000000

const EPersistStateDisk = 0
const EPersistStateLoading = 1
const EPersistStateMemory = 2
const EPersistStatePrepareUnloading = 3
const EPersistStateUnloading = 4

// const EPersistErrorEngineNil = PersistError("persist: engine is nil")           // 启动关闭错误: 数据库连接失败
// const EPersistErrorTempFileExist = PersistError("persist: temp file exist")     // 启动关闭错误: 存在临时bomb文件
// const EPersistErrorInvalidBombFile = PersistError("persist: invalid bomb file") // 启动关闭错误: 无效的bomb文件
// const EPersistErrorUnknownError = PersistError("persist: unknown error")        // 导入导出错误: 未知错误, 可能是并发引起
// const EPersistErrorIncorrectState = PersistError("persist: incorrect state")    // 导入导出错误: 重复全导入或正在全导出
// const EPersistErrorUnloading = PersistError("persist: unloading state")         // 导入导出错误: 正在导出, 导出完成后方可导入
// const EPersistErrorAlreadyLoadAll = PersistError("persist: already load all")   // 导入导出错误: 已经全导入不能再按照key操作
// const EPersistErrorLoading = PersistError("persist: loading state")             // 导入导出错误: 正在导入, 导入完成后方可导出
// const EPersistErrorAlreadyLoad = PersistError("persist: already load")          // 导入导出错误: 重复导入
// const EPersistErrorAlreadyUnload = PersistError("persist: already unload")      // 导入导出错误: 重复导出
// const EPersistErrorNil = PersistError("persist: nil")                           // 增删改查错误: 非法的内存地址或空指针
// const EPersistErrorAlreadyExist = PersistError("persist: already exist")        // 增删改查错误: 对象已经存在
// const EPersistErrorNotInMemory = PersistError("persist: not in memory")         // 增删改查错误: 数据不在内存中
// const EPersistErrorOutOfDate = PersistError("persist: out of date")             // 增删改查错误: 数据过期, 应当重新查询
