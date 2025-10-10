package persist

const (
	EGlobalManagerStateIdle   = 0 // 初始化
	EGlobalManagerStateNormal = 1 // 正常运行
	EGlobalManagerStatePanic  = 2 // 非法停止

	EGlobalTableStateDisk      = 0 // 导出
	EGlobalTableStateLoading   = 1 // 全导入开始
	EGlobalTableStateMemory    = 2 // 全导入完成
	EGlobalTableStateUnloading = 3 // 正在全导出

	EGlobalLoadStateDisk             = 0 // 不存在 or 导出
	EGlobalLoadStateLoading          = 1 // 导入开始
	EGlobalLoadStateMemory           = 2 // 导入完成
	EGlobalLoadStatePrepareUnloading = 3 // 准备导出
	EGlobalLoadStateUnloading        = 4 // 正在导出

	EGlobalOpInsert = 1 // 新建
	EGlobalOpUpdate = 2 // 修改
	EGlobalOpDelete = 3 // 删除
	EGlobalOpUnload = 4 // 导出

	EGlobalCollectStateNormal    = 0 // 正常
	EGlobalCollectStateSaveSync  = 1 // 开始退出, 清理同步队列
	EGlobalCollectStateSaveCache = 2 // 开始退出,清理缓存队列
	EGlobalCollectStateSaveDone  = 3 // 写回完成

)
const (
	EGlobalWordSize            = GlobalFieldIndex(64) // 每个单元的位数
	EGlobalLog2WordSize        = GlobalFieldIndex(6)  // 每个单元的位数的对数
	EGlobalAllBits      uint64 = 0xffffffffffffffff   // 全部位都为1
)

type GlobalFieldIndex = uint

// GlobalBitSet 全局位图管理
type GlobalBitSet[T any] struct {
	set []uint64
	t   T
}

func InitGlobalBitSet[T any]() GlobalBitSet[T] {
	t := new(T)
	return GlobalBitSet[T]{
		set: make([]uint64, (GetFieldNum(t)>>EGlobalLog2WordSize)+1),
		t:   *t,
	}
}

// Get 获取位 i 的值
func (b *GlobalBitSet[T]) Get(i GlobalFieldIndex) bool {
	globalFiledIndexLength := GlobalFieldIndex(GetFieldNum(b.t))
	if i >= globalFiledIndexLength {
		return false
	}
	return b.set[i>>EGlobalLog2WordSize]&(1<<(i&(EGlobalWordSize-1))) != 0
}

// Set 设置位 i
func (b *GlobalBitSet[T]) Set(i GlobalFieldIndex) *GlobalBitSet[T] {
	globalFiledIndexLength := GlobalFieldIndex(GetFieldNum(b.t))
	if i >= globalFiledIndexLength {
		return nil
	}
	b.set[i>>EGlobalLog2WordSize] |= 1 << (i & (EGlobalWordSize - 1))
	return b
}

// Clear 清除位 i
func (b *GlobalBitSet[T]) Clear(i GlobalFieldIndex) *GlobalBitSet[T] {
	globalFiledIndexLength := GlobalFieldIndex(GetFieldNum(b.t))
	if i >= globalFiledIndexLength {
		return b
	}
	b.set[i>>EGlobalLog2WordSize] &^= 1 << (i & (EGlobalWordSize - 1))
	return b
}

// Merge 合并两个位图
func (b *GlobalBitSet[T]) Merge(compare GlobalBitSet[T]) *GlobalBitSet[T] {
	for i, word := range b.set {
		b.set[i] = word | compare.set[i]
	}
	return b
}

// ClearAll 清除所有位
func (b *GlobalBitSet[T]) ClearAll() *GlobalBitSet[T] {
	if b != nil {
		for i := range b.set {
			b.set[i] = 0
		}
	}
	return b
}

// SetAll 设置所有位
func (b *GlobalBitSet[T]) SetAll() *GlobalBitSet[T] {
	if b != nil {
		for i := range b.set {
			b.set[i] = EGlobalAllBits
		}
	}
	return b
}

// IsSetAll 是否所有位都被设置
func (b *GlobalBitSet[T]) IsSetAll() bool {
	if b != nil {
		for i := range b.set {
			if b.set[i] == EGlobalAllBits {
				continue
			}
			return false
		}
	}
	return true
}
