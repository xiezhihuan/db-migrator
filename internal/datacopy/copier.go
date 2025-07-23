package datacopy

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"db-migrator/internal/types"
)

// CopyStrategy 数据复制策略
type CopyStrategy string

const (
	CopyStrategyOverwrite CopyStrategy = "overwrite" // 完全覆盖
	CopyStrategyMerge     CopyStrategy = "merge"     // 智能合并
	CopyStrategyInsertNew CopyStrategy = "insert"    // 仅插入新数据
	CopyStrategyIgnore    CopyStrategy = "ignore"    // 忽略重复
)

// CopyScope 数据复制范围
type CopyScope string

const (
	CopyScopeFullTable   CopyScope = "full"      // 整表复制
	CopyScopeConditional CopyScope = "condition" // 条件复制
	CopyScopeMapping     CopyScope = "mapping"   // 字段映射
	CopyScopeTransform   CopyScope = "transform" // 数据转换
)

// FieldMapping 字段映射配置
type FieldMapping struct {
	SourceField string `json:"source_field"`
	TargetField string `json:"target_field"`
	Transform   string `json:"transform,omitempty"` // SQL转换表达式
}

// CopyConfig 复制配置
type CopyConfig struct {
	Strategy      CopyStrategy              `json:"strategy"`
	Scope         CopyScope                 `json:"scope"`
	Tables        []string                  `json:"tables"`
	Conditions    map[string]string         `json:"conditions,omitempty"`     // 表名 -> WHERE条件
	FieldMappings map[string][]FieldMapping `json:"field_mappings,omitempty"` // 表名 -> 字段映射
	BatchSize     int                       `json:"batch_size"`
	Timeout       time.Duration             `json:"timeout"`
	OnError       string                    `json:"on_error"` // "stop", "continue", "rollback"
}

// ProgressCallback 进度回调函数
type ProgressCallback func(table string, current, total int64, err error)

// DataCopier 数据复制器
type DataCopier struct {
	sourceDB   types.DB
	targetDB   types.DB
	config     CopyConfig
	onProgress ProgressCallback
}

// NewDataCopier 创建数据复制器
func NewDataCopier(sourceDB, targetDB types.DB, config CopyConfig) *DataCopier {
	if config.BatchSize <= 0 {
		config.BatchSize = 1000
	}
	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Minute
	}
	if config.OnError == "" {
		config.OnError = "stop"
	}

	return &DataCopier{
		sourceDB: sourceDB,
		targetDB: targetDB,
		config:   config,
	}
}

// SetProgressCallback 设置进度回调
func (dc *DataCopier) SetProgressCallback(callback ProgressCallback) {
	dc.onProgress = callback
}

// CopyData 执行数据复制
func (dc *DataCopier) CopyData(ctx context.Context) error {
	// 设置超时上下文
	ctx, cancel := context.WithTimeout(ctx, dc.config.Timeout)
	defer cancel()

	var errors []string

	for _, tableName := range dc.config.Tables {
		if err := dc.copyTable(ctx, tableName); err != nil {
			errorMsg := fmt.Sprintf("复制表 %s 失败: %v", tableName, err)
			errors = append(errors, errorMsg)

			// 通知进度回调
			if dc.onProgress != nil {
				dc.onProgress(tableName, 0, 0, err)
			}

			// 根据错误处理策略决定是否继续
			switch dc.config.OnError {
			case "stop":
				return fmt.Errorf(errorMsg)
			case "rollback":
				// TODO: 实现回滚逻辑
				return fmt.Errorf("复制失败，需要回滚: %s", errorMsg)
			case "continue":
				continue
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("部分表复制失败:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}

// copyTable 复制单个表
func (dc *DataCopier) copyTable(ctx context.Context, tableName string) error {
	// 检查源表是否存在
	sourceExists, err := dc.tableExists(dc.sourceDB, tableName)
	if err != nil {
		return fmt.Errorf("检查源表存在性失败: %v", err)
	}
	if !sourceExists {
		return fmt.Errorf("源表 %s 不存在", tableName)
	}

	// 检查目标表是否存在
	targetExists, err := dc.tableExists(dc.targetDB, tableName)
	if err != nil {
		return fmt.Errorf("检查目标表存在性失败: %v", err)
	}
	if !targetExists {
		return fmt.Errorf("目标表 %s 不存在", tableName)
	}

	// 获取表总行数（用于进度显示）
	totalRows, err := dc.getTableRowCount(dc.sourceDB, tableName)
	if err != nil {
		return fmt.Errorf("获取表行数失败: %v", err)
	}

	// 根据策略清理目标表
	if dc.config.Strategy == CopyStrategyOverwrite {
		if err := dc.truncateTable(dc.targetDB, tableName); err != nil {
			return fmt.Errorf("清空目标表失败: %v", err)
		}
	}

	// 执行数据复制
	return dc.copyTableData(ctx, tableName, totalRows)
}

// copyTableData 复制表数据
func (dc *DataCopier) copyTableData(ctx context.Context, tableName string, totalRows int64) error {
	// 构建查询SQL
	selectSQL, err := dc.buildSelectSQL(tableName)
	if err != nil {
		return err
	}

	// 查询源数据
	rows, err := dc.sourceDB.Query(selectSQL)
	if err != nil {
		return fmt.Errorf("查询源数据失败: %v", err)
	}
	defer rows.Close()

	// 获取列信息
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("获取列信息失败: %v", err)
	}

	// 映射字段名
	targetColumns := dc.mapColumns(tableName, columns)

	// 批量处理数据
	var processedRows int64
	batch := make([][]interface{}, 0, dc.config.BatchSize)

	for rows.Next() {
		// 创建值容器
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// 扫描数据
		if err := rows.Scan(valuePtrs...); err != nil {
			return fmt.Errorf("扫描数据失败: %v", err)
		}

		// 转换数据
		transformedValues, err := dc.transformValues(tableName, columns, values)
		if err != nil {
			return fmt.Errorf("转换数据失败: %v", err)
		}

		batch = append(batch, transformedValues)

		// 达到批次大小时执行插入
		if len(batch) >= dc.config.BatchSize {
			if err := dc.insertBatch(tableName, targetColumns, batch); err != nil {
				return fmt.Errorf("批量插入失败: %v", err)
			}

			processedRows += int64(len(batch))
			batch = batch[:0] // 清空批次

			// 通知进度
			if dc.onProgress != nil {
				dc.onProgress(tableName, processedRows, totalRows, nil)
			}

			// 检查上下文是否被取消
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
		}
	}

	// 处理剩余数据
	if len(batch) > 0 {
		if err := dc.insertBatch(tableName, targetColumns, batch); err != nil {
			return fmt.Errorf("插入剩余数据失败: %v", err)
		}
		processedRows += int64(len(batch))
	}

	// 通知完成
	if dc.onProgress != nil {
		dc.onProgress(tableName, processedRows, totalRows, nil)
	}

	return nil
}

// buildSelectSQL 构建查询SQL
func (dc *DataCopier) buildSelectSQL(tableName string) (string, error) {
	sql := fmt.Sprintf("SELECT * FROM %s", tableName)

	// 添加条件
	if condition, exists := dc.config.Conditions[tableName]; exists && condition != "" {
		sql += " WHERE " + condition
	}

	return sql, nil
}

// mapColumns 映射列名
func (dc *DataCopier) mapColumns(tableName string, sourceColumns []string) []string {
	mappings, exists := dc.config.FieldMappings[tableName]
	if !exists {
		return sourceColumns
	}

	// 创建映射关系
	fieldMap := make(map[string]string)
	for _, mapping := range mappings {
		fieldMap[mapping.SourceField] = mapping.TargetField
	}

	// 应用映射
	targetColumns := make([]string, len(sourceColumns))
	for i, col := range sourceColumns {
		if mapped, exists := fieldMap[col]; exists {
			targetColumns[i] = mapped
		} else {
			targetColumns[i] = col
		}
	}

	return targetColumns
}

// transformValues 转换数据值
func (dc *DataCopier) transformValues(tableName string, columns []string, values []interface{}) ([]interface{}, error) {
	mappings, exists := dc.config.FieldMappings[tableName]
	if !exists {
		return values, nil
	}

	// 应用转换
	for _, mapping := range mappings {
		if mapping.Transform != "" {
			// 找到对应的列索引
			for i, col := range columns {
				if col == mapping.SourceField {
					// 这里简化处理，实际应该执行SQL转换
					// 可以扩展为支持更复杂的转换逻辑
					transformed, err := dc.applyTransform(mapping.Transform, values[i])
					if err != nil {
						return nil, err
					}
					values[i] = transformed
					break
				}
			}
		}
	}

	return values, nil
}

// applyTransform 应用数据转换
func (dc *DataCopier) applyTransform(transform string, value interface{}) (interface{}, error) {
	// 简单的转换支持，可以扩展
	switch transform {
	case "UPPER":
		if str, ok := value.(string); ok {
			return strings.ToUpper(str), nil
		}
	case "LOWER":
		if str, ok := value.(string); ok {
			return strings.ToLower(str), nil
		}
	case "NOW()":
		return time.Now(), nil
	}

	// 正则替换示例: REPLACE(field, 'old', 'new')
	replacePattern := regexp.MustCompile(`REPLACE\(field,\s*'([^']+)',\s*'([^']*)'\)`)
	if matches := replacePattern.FindStringSubmatch(transform); len(matches) == 3 {
		if str, ok := value.(string); ok {
			return strings.ReplaceAll(str, matches[1], matches[2]), nil
		}
	}

	return value, nil
}

// insertBatch 批量插入数据
func (dc *DataCopier) insertBatch(tableName string, columns []string, batch [][]interface{}) error {
	if len(batch) == 0 {
		return nil
	}

	// 构建插入SQL
	var insertType string
	switch dc.config.Strategy {
	case CopyStrategyOverwrite:
		insertType = "INSERT"
	case CopyStrategyMerge:
		insertType = "INSERT ... ON DUPLICATE KEY UPDATE"
	case CopyStrategyInsertNew:
		insertType = "INSERT IGNORE"
	case CopyStrategyIgnore:
		insertType = "INSERT IGNORE"
	default:
		insertType = "INSERT"
	}

	// 构建占位符
	placeholders := "(" + strings.Repeat("?,", len(columns)-1) + "?)"
	allPlaceholders := strings.Repeat(placeholders+",", len(batch)-1) + placeholders

	var sql string
	if dc.config.Strategy == CopyStrategyMerge {
		// 构建UPDATE部分
		updateParts := make([]string, len(columns))
		for i, col := range columns {
			updateParts[i] = fmt.Sprintf("%s = VALUES(%s)", col, col)
		}

		sql = fmt.Sprintf("INSERT INTO %s (%s) VALUES %s ON DUPLICATE KEY UPDATE %s",
			tableName,
			strings.Join(columns, ", "),
			allPlaceholders,
			strings.Join(updateParts, ", "))
	} else {
		sql = fmt.Sprintf("%s INTO %s (%s) VALUES %s",
			insertType,
			tableName,
			strings.Join(columns, ", "),
			allPlaceholders)
	}

	// 展平参数
	args := make([]interface{}, 0, len(batch)*len(columns))
	for _, row := range batch {
		args = append(args, row...)
	}

	// 执行插入
	_, err := dc.targetDB.Exec(sql, args...)
	return err
}

// 辅助方法

func (dc *DataCopier) tableExists(db types.DB, tableName string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", tableName).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (dc *DataCopier) getTableRowCount(db types.DB, tableName string) (int64, error) {
	var count int64
	err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count)
	return count, err
}

func (dc *DataCopier) truncateTable(db types.DB, tableName string) error {
	_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s", tableName))
	return err
}

// CrossDatabaseCopier 跨数据库复制器
type CrossDatabaseCopier struct {
	dbManager types.DBManager // Changed from *types.DBManager to types.DBManager
}

// NewCrossDatabaseCopier 创建跨数据库复制器
func NewCrossDatabaseCopier(dbManager types.DBManager) *CrossDatabaseCopier { // Changed from *types.DBManager to types.DBManager
	return &CrossDatabaseCopier{
		dbManager: dbManager,
	}
}

// CopyBetweenDatabases 在数据库之间复制数据
func (cdc *CrossDatabaseCopier) CopyBetweenDatabases(ctx context.Context, sourceDB, targetDB string, config CopyConfig, callback ProgressCallback) error {
	// 获取源数据库连接
	sourceConn, err := cdc.dbManager.GetDatabase(sourceDB)
	if err != nil {
		return fmt.Errorf("连接源数据库失败: %v", err)
	}

	// 获取目标数据库连接
	targetConn, err := cdc.dbManager.GetDatabase(targetDB)
	if err != nil {
		return fmt.Errorf("连接目标数据库失败: %v", err)
	}

	// 创建复制器
	copier := NewDataCopier(sourceConn, targetConn, config)
	copier.SetProgressCallback(callback)

	// 执行复制
	return copier.CopyData(ctx)
}

// CopyToMultipleDatabases 复制到多个数据库
func (cdc *CrossDatabaseCopier) CopyToMultipleDatabases(ctx context.Context, sourceDB string, targetDBs []string, config CopyConfig, callback ProgressCallback) error {
	var errors []string

	for _, targetDB := range targetDBs {
		if err := cdc.CopyBetweenDatabases(ctx, sourceDB, targetDB, config, callback); err != nil {
			errors = append(errors, fmt.Sprintf("复制到 %s 失败: %v", targetDB, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("部分数据库复制失败:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}
