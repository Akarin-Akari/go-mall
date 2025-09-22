package idgen

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// IDGenerator 分布式ID生成器接口
type IDGenerator interface {
	GenerateOrderNo() string
	GenerateID() int64
}

// UUIDGenerator UUID生成器
type UUIDGenerator struct{}

// NewUUIDGenerator 创建UUID生成器
func NewUUIDGenerator() *UUIDGenerator {
	return &UUIDGenerator{}
}

// GenerateOrderNo 生成订单号
func (g *UUIDGenerator) GenerateOrderNo() string {
	// 生成UUID并转换为订单号格式
	id := uuid.New()
	// 使用时间前缀 + UUID后8位，确保可读性和唯一性
	timestamp := time.Now().Format("20060102150405")
	uuidSuffix := fmt.Sprintf("%08X", id.ID())
	return fmt.Sprintf("ORD%s%s", timestamp, uuidSuffix)
}

// GenerateID 生成数字ID
func (g *UUIDGenerator) GenerateID() int64 {
	id := uuid.New()
	return int64(id.ID())
}

// SnowflakeGenerator 雪花算法生成器
type SnowflakeGenerator struct {
	mutex     sync.Mutex
	machineID int64 // 机器ID (0-1023)
	sequence  int64 // 序列号 (0-4095)
	lastTime  int64 // 上次生成时间戳

	// 雪花算法配置
	epoch        int64 // 起始时间戳 (2020-01-01 00:00:00)
	machineBits  int64 // 机器ID位数
	sequenceBits int64 // 序列号位数

	// 位移量
	machineShift int64
	timeShift    int64

	// 最大值
	maxMachine  int64
	maxSequence int64
}

// NewSnowflakeGenerator 创建雪花算法生成器
func NewSnowflakeGenerator(machineID int64) (*SnowflakeGenerator, error) {
	// 2020-01-01 00:00:00 UTC
	epoch := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()

	machineBits := int64(10)  // 支持1024台机器
	sequenceBits := int64(12) // 每毫秒支持4096个序列号

	maxMachine := int64(-1) ^ (int64(-1) << machineBits)
	maxSequence := int64(-1) ^ (int64(-1) << sequenceBits)

	if machineID < 0 || machineID > maxMachine {
		return nil, fmt.Errorf("机器ID必须在0-%d之间", maxMachine)
	}

	return &SnowflakeGenerator{
		machineID:    machineID,
		epoch:        epoch,
		machineBits:  machineBits,
		sequenceBits: sequenceBits,
		machineShift: sequenceBits,
		timeShift:    sequenceBits + machineBits,
		maxMachine:   maxMachine,
		maxSequence:  maxSequence,
	}, nil
}

// GenerateID 生成雪花ID
func (g *SnowflakeGenerator) GenerateID() int64 {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	now := time.Now().UnixMilli()

	if now < g.lastTime {
		// 时钟回拨，等待到上次时间
		time.Sleep(time.Duration(g.lastTime-now) * time.Millisecond)
		now = time.Now().UnixMilli()
	}

	if now == g.lastTime {
		// 同一毫秒内，序列号递增
		g.sequence = (g.sequence + 1) & g.maxSequence
		if g.sequence == 0 {
			// 序列号溢出，等待下一毫秒
			for now <= g.lastTime {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		// 新的毫秒，序列号重置
		g.sequence = 0
	}

	g.lastTime = now

	// 组装ID: 时间戳(41位) + 机器ID(10位) + 序列号(12位)
	id := ((now - g.epoch) << g.timeShift) |
		(g.machineID << g.machineShift) |
		g.sequence

	return id
}

// GenerateOrderNo 生成订单号
func (g *SnowflakeGenerator) GenerateOrderNo() string {
	id := g.GenerateID()
	// 订单号格式: ORD + 雪花ID
	return fmt.Sprintf("ORD%d", id)
}

// SimpleIDGenerator 简单ID生成器（用于测试）
type SimpleIDGenerator struct {
	mutex   sync.Mutex
	counter int64
}

// NewSimpleIDGenerator 创建简单ID生成器
func NewSimpleIDGenerator() *SimpleIDGenerator {
	return &SimpleIDGenerator{
		counter: time.Now().UnixNano() / 1000000, // 毫秒时间戳作为起始值
	}
}

// GenerateID 生成递增ID
func (g *SimpleIDGenerator) GenerateID() int64 {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.counter++
	return g.counter
}

// GenerateOrderNo 生成订单号
func (g *SimpleIDGenerator) GenerateOrderNo() string {
	id := g.GenerateID()
	return fmt.Sprintf("ORD%d", id)
}

// RandomIDGenerator 随机ID生成器
type RandomIDGenerator struct{}

// NewRandomIDGenerator 创建随机ID生成器
func NewRandomIDGenerator() *RandomIDGenerator {
	return &RandomIDGenerator{}
}

// GenerateID 生成随机ID
func (g *RandomIDGenerator) GenerateID() int64 {
	// 生成8字节随机数
	bytes := make([]byte, 8)
	rand.Read(bytes)

	var id int64
	for i := 0; i < 8; i++ {
		id = (id << 8) | int64(bytes[i])
	}

	// 确保为正数
	if id < 0 {
		id = -id
	}

	return id
}

// GenerateOrderNo 生成订单号
func (g *RandomIDGenerator) GenerateOrderNo() string {
	id := g.GenerateID()
	return fmt.Sprintf("ORD%d", id)
}

// GlobalIDGenerator 全局ID生成器
var GlobalIDGenerator IDGenerator

// InitGlobalIDGenerator 初始化全局ID生成器
func InitGlobalIDGenerator(generatorType string, machineID int64) error {
	switch generatorType {
	case "uuid":
		GlobalIDGenerator = NewUUIDGenerator()
	case "snowflake":
		var err error
		GlobalIDGenerator, err = NewSnowflakeGenerator(machineID)
		if err != nil {
			return err
		}
	case "simple":
		GlobalIDGenerator = NewSimpleIDGenerator()
	case "random":
		GlobalIDGenerator = NewRandomIDGenerator()
	default:
		return fmt.Errorf("不支持的ID生成器类型: %s", generatorType)
	}
	return nil
}

// GenerateOrderNo 生成订单号（全局方法）
func GenerateOrderNo() string {
	if GlobalIDGenerator == nil {
		// 默认使用UUID生成器
		GlobalIDGenerator = NewUUIDGenerator()
	}
	return GlobalIDGenerator.GenerateOrderNo()
}

// GenerateID 生成ID（全局方法）
func GenerateID() int64 {
	if GlobalIDGenerator == nil {
		// 默认使用UUID生成器
		GlobalIDGenerator = NewUUIDGenerator()
	}
	return GlobalIDGenerator.GenerateID()
}
