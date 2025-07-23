package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"

	"db-migrator/internal/types"

	"gopkg.in/yaml.v3"
)

// DataBuilder 数据构建器
type DataBuilder struct {
	db      types.DB
	checker types.Checker
}

// NewDataBuilder 创建数据构建器
func NewDataBuilder(checker types.Checker, db types.DB) *DataBuilder {
	return &DataBuilder{
		db:      db,
		checker: checker,
	}
}

// DataInsertStrategy 数据插入策略
type DataInsertStrategy string

const (
	StrategyInsertOnly        DataInsertStrategy = "insert"   // 仅插入，遇到重复跳过
	StrategyInsertOrUpdate    DataInsertStrategy = "upsert"   // 插入或更新
	StrategyTruncateAndInsert DataInsertStrategy = "truncate" // 清空后插入
	StrategyReplace           DataInsertStrategy = "replace"  // 替换
	StrategyIgnore            DataInsertStrategy = "ignore"   // 忽略重复
)

// TableDataBuilder 表数据构建器
type TableDataBuilder struct {
	dataBuilder *DataBuilder
	tableName   string
	strategy    DataInsertStrategy
	condition   string
	batchSize   int
}

// Table 指定表名
func (db *DataBuilder) Table(tableName string) *TableDataBuilder {
	return &TableDataBuilder{
		dataBuilder: db,
		tableName:   tableName,
		strategy:    StrategyInsertOnly,
		batchSize:   1000, // 默认批量大小
	}
}

// Strategy 设置插入策略
func (tdb *TableDataBuilder) Strategy(strategy DataInsertStrategy) *TableDataBuilder {
	tdb.strategy = strategy
	return tdb
}

// Where 设置条件（用于更新策略）
func (tdb *TableDataBuilder) Where(condition string) *TableDataBuilder {
	tdb.condition = condition
	return tdb
}

// BatchSize 设置批量插入大小
func (tdb *TableDataBuilder) BatchSize(size int) *TableDataBuilder {
	tdb.batchSize = size
	return tdb
}

// InsertData 插入数据记录
func (tdb *TableDataBuilder) InsertData(ctx context.Context, data []map[string]interface{}) error {
	if len(data) == 0 {
		return nil
	}

	// 检查表是否存在
	exists, err := tdb.dataBuilder.checker.TableExists(ctx, tdb.tableName)
	if err != nil {
		return fmt.Errorf("检查表 %s 是否存在失败: %v", tdb.tableName, err)
	}
	if !exists {
		return fmt.Errorf("表 %s 不存在", tdb.tableName)
	}

	switch tdb.strategy {
	case StrategyTruncateAndInsert:
		return tdb.truncateAndInsert(ctx, data)
	case StrategyInsertOrUpdate:
		return tdb.insertOrUpdate(ctx, data)
	case StrategyReplace:
		return tdb.replace(ctx, data)
	case StrategyIgnore:
		return tdb.insertIgnore(ctx, data)
	default:
		return tdb.insertOnly(ctx, data)
	}
}

// InsertFromStruct 从结构体插入数据
func (tdb *TableDataBuilder) InsertFromStruct(ctx context.Context, structs interface{}) error {
	data, err := tdb.structsToMaps(structs)
	if err != nil {
		return err
	}
	return tdb.InsertData(ctx, data)
}

// InsertFromJSON 从JSON文件插入数据
func (tdb *TableDataBuilder) InsertFromJSON(ctx context.Context, filePath string) error {
	jsonData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取JSON文件失败: %v", err)
	}

	var data []map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("解析JSON数据失败: %v", err)
	}

	return tdb.InsertData(ctx, data)
}

// InsertFromYAML 从YAML文件插入数据
func (tdb *TableDataBuilder) InsertFromYAML(ctx context.Context, filePath string) error {
	yamlData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取YAML文件失败: %v", err)
	}

	var data []map[string]interface{}
	if err := yaml.Unmarshal(yamlData, &data); err != nil {
		return fmt.Errorf("解析YAML数据失败: %v", err)
	}

	return tdb.InsertData(ctx, data)
}

// InsertSQL 直接执行SQL插入
func (tdb *TableDataBuilder) InsertSQL(ctx context.Context, sql string, args ...interface{}) error {
	_, err := tdb.dataBuilder.db.Exec(sql, args...)
	if err != nil {
		return fmt.Errorf("执行SQL插入失败: %v", err)
	}
	return nil
}

// 私有方法实现

func (tdb *TableDataBuilder) insertOnly(ctx context.Context, data []map[string]interface{}) error {
	return tdb.batchInsert(ctx, data, "INSERT")
}

func (tdb *TableDataBuilder) insertIgnore(ctx context.Context, data []map[string]interface{}) error {
	return tdb.batchInsert(ctx, data, "INSERT IGNORE")
}

func (tdb *TableDataBuilder) replace(ctx context.Context, data []map[string]interface{}) error {
	return tdb.batchInsert(ctx, data, "REPLACE")
}

func (tdb *TableDataBuilder) truncateAndInsert(ctx context.Context, data []map[string]interface{}) error {
	// 先清空表
	_, err := tdb.dataBuilder.db.Exec(fmt.Sprintf("TRUNCATE TABLE %s", tdb.tableName))
	if err != nil {
		return fmt.Errorf("清空表 %s 失败: %v", tdb.tableName, err)
	}

	// 再插入数据
	return tdb.insertOnly(ctx, data)
}

func (tdb *TableDataBuilder) insertOrUpdate(ctx context.Context, data []map[string]interface{}) error {
	if len(data) == 0 {
		return nil
	}

	// 构建字段列表
	columns := make([]string, 0, len(data[0]))
	for col := range data[0] {
		columns = append(columns, col)
	}

	// 构建更新部分
	updateParts := make([]string, len(columns))
	for i, col := range columns {
		updateParts[i] = fmt.Sprintf("%s = VALUES(%s)", col, col)
	}

	// 批量处理
	for i := 0; i < len(data); i += tdb.batchSize {
		end := i + tdb.batchSize
		if end > len(data) {
			end = len(data)
		}

		batch := data[i:end]
		sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s ON DUPLICATE KEY UPDATE %s",
			tdb.tableName,
			strings.Join(columns, ", "),
			tdb.buildValuePlaceholders(len(columns), len(batch)),
			strings.Join(updateParts, ", "))

		args := tdb.flattenBatchData(batch, columns)
		_, err := tdb.dataBuilder.db.Exec(sql, args...)
		if err != nil {
			return fmt.Errorf("批量插入或更新失败: %v", err)
		}
	}

	return nil
}

func (tdb *TableDataBuilder) batchInsert(ctx context.Context, data []map[string]interface{}, insertType string) error {
	if len(data) == 0 {
		return nil
	}

	// 构建字段列表
	columns := make([]string, 0, len(data[0]))
	for col := range data[0] {
		columns = append(columns, col)
	}

	// 批量处理
	for i := 0; i < len(data); i += tdb.batchSize {
		end := i + tdb.batchSize
		if end > len(data) {
			end = len(data)
		}

		batch := data[i:end]
		sql := fmt.Sprintf("%s INTO %s (%s) VALUES %s",
			insertType,
			tdb.tableName,
			strings.Join(columns, ", "),
			tdb.buildValuePlaceholders(len(columns), len(batch)))

		args := tdb.flattenBatchData(batch, columns)
		_, err := tdb.dataBuilder.db.Exec(sql, args...)
		if err != nil {
			return fmt.Errorf("批量%s失败: %v", insertType, err)
		}
	}

	return nil
}

func (tdb *TableDataBuilder) buildValuePlaceholders(columnCount, rowCount int) string {
	rowPlaceholder := "(" + strings.Repeat("?,", columnCount-1) + "?)"
	return strings.Repeat(rowPlaceholder+",", rowCount-1) + rowPlaceholder
}

func (tdb *TableDataBuilder) flattenBatchData(data []map[string]interface{}, columns []string) []interface{} {
	args := make([]interface{}, 0, len(data)*len(columns))
	for _, row := range data {
		for _, col := range columns {
			args = append(args, row[col])
		}
	}
	return args
}

func (tdb *TableDataBuilder) structsToMaps(structs interface{}) ([]map[string]interface{}, error) {
	v := reflect.ValueOf(structs)
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("期望切片类型，得到 %v", v.Kind())
	}

	var result []map[string]interface{}
	for i := 0; i < v.Len(); i++ {
		item := v.Index(i)
		if item.Kind() == reflect.Ptr {
			item = item.Elem()
		}

		if item.Kind() != reflect.Struct {
			return nil, fmt.Errorf("期望结构体类型，得到 %v", item.Kind())
		}

		m := make(map[string]interface{})
		t := item.Type()
		for j := 0; j < item.NumField(); j++ {
			field := t.Field(j)
			tag := field.Tag.Get("db")
			if tag == "-" {
				continue
			}
			if tag == "" {
				tag = strings.ToLower(field.Name)
			}

			m[tag] = item.Field(j).Interface()
		}
		result = append(result, m)
	}

	return result, nil
}

// 便捷方法

// QuickInsert 快速插入数据
func (db *DataBuilder) QuickInsert(ctx context.Context, tableName string, data []map[string]interface{}) error {
	return db.Table(tableName).InsertData(ctx, data)
}

// QuickInsertFromJSON 快速从JSON插入
func (db *DataBuilder) QuickInsertFromJSON(ctx context.Context, tableName, jsonFile string) error {
	return db.Table(tableName).InsertFromJSON(ctx, jsonFile)
}

// QuickInsertFromYAML 快速从YAML插入
func (db *DataBuilder) QuickInsertFromYAML(ctx context.Context, tableName, yamlFile string) error {
	return db.Table(tableName).InsertFromYAML(ctx, yamlFile)
}

// UpsertData 插入或更新数据
func (db *DataBuilder) UpsertData(ctx context.Context, tableName string, data []map[string]interface{}) error {
	return db.Table(tableName).Strategy(StrategyInsertOrUpdate).InsertData(ctx, data)
}
