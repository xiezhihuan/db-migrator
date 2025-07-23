package builder

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// TableBuilder 高级表构建器
type TableBuilder struct {
	sqlBuilder  *SQLBuilder
	tableName   string
	columns     []AdvancedColumn
	indexes     []IndexDef
	foreignKeys []ForeignKeyDef
	options     TableOptions
}

// AdvancedColumn 高级列定义
type AdvancedColumn struct {
	Name       string
	Type       ColumnType
	Size       int
	NotNull    bool
	Default    interface{}
	AutoIncr   bool
	PrimaryKey bool
	Unique     bool
	Comment    string
	After      string // 在指定列之后添加
}

// ColumnType 列类型枚举
type ColumnType string

const (
	TypeInt       ColumnType = "INT"
	TypeBigInt    ColumnType = "BIGINT"
	TypeSmallInt  ColumnType = "SMALLINT"
	TypeTinyInt   ColumnType = "TINYINT"
	TypeDecimal   ColumnType = "DECIMAL"
	TypeFloat     ColumnType = "FLOAT"
	TypeDouble    ColumnType = "DOUBLE"
	TypeVarchar   ColumnType = "VARCHAR"
	TypeChar      ColumnType = "CHAR"
	TypeText      ColumnType = "TEXT"
	TypeLongText  ColumnType = "LONGTEXT"
	TypeJson      ColumnType = "JSON"
	TypeDate      ColumnType = "DATE"
	TypeDateTime  ColumnType = "DATETIME"
	TypeTimestamp ColumnType = "TIMESTAMP"
	TypeTime      ColumnType = "TIME"
	TypeBoolean   ColumnType = "BOOLEAN"
	TypeEnum      ColumnType = "ENUM"
	TypeSet       ColumnType = "SET"
	TypeBlob      ColumnType = "BLOB"
	TypeLongBlob  ColumnType = "LONGBLOB"
)

// IndexDef 索引定义
type IndexDef struct {
	Name    string
	Columns []string
	Type    IndexType
	Unique  bool
}

// IndexType 索引类型
type IndexType string

const (
	IndexNormal   IndexType = "INDEX"
	IndexUnique   IndexType = "UNIQUE"
	IndexFullText IndexType = "FULLTEXT"
	IndexSpatial  IndexType = "SPATIAL"
)

// ForeignKeyDef 外键定义
type ForeignKeyDef struct {
	Name      string
	Column    string
	RefTable  string
	RefColumn string
	OnDelete  ReferenceAction
	OnUpdate  ReferenceAction
}

// ReferenceAction 引用动作
type ReferenceAction string

const (
	ActionCascade    ReferenceAction = "CASCADE"
	ActionSetNull    ReferenceAction = "SET NULL"
	ActionRestrict   ReferenceAction = "RESTRICT"
	ActionNoAction   ReferenceAction = "NO ACTION"
	ActionSetDefault ReferenceAction = "SET DEFAULT"
)

// TableOptions 表选项
type TableOptions struct {
	Engine    string
	Charset   string
	Collation string
	Comment   string
	AutoIncr  int64
	RowFormat string
}

// NewTableBuilder 创建表构建器
func NewTableBuilder(sqlBuilder *SQLBuilder, tableName string) *TableBuilder {
	return &TableBuilder{
		sqlBuilder:  sqlBuilder,
		tableName:   tableName,
		columns:     make([]AdvancedColumn, 0),
		indexes:     make([]IndexDef, 0),
		foreignKeys: make([]ForeignKeyDef, 0),
		options: TableOptions{
			Engine:  "InnoDB",
			Charset: "utf8mb4",
		},
	}
}

// ID 添加主键ID列
func (tb *TableBuilder) ID(name ...string) *TableBuilder {
	columnName := "id"
	if len(name) > 0 {
		columnName = name[0]
	}

	tb.columns = append(tb.columns, AdvancedColumn{
		Name:       columnName,
		Type:       TypeInt,
		NotNull:    true,
		AutoIncr:   true,
		PrimaryKey: true,
		Comment:    "主键ID",
	})
	return tb
}

// String 添加字符串列
func (tb *TableBuilder) String(name string, size int) *ColumnBuilder {
	column := AdvancedColumn{
		Name: name,
		Type: TypeVarchar,
		Size: size,
	}
	return &ColumnBuilder{
		tableBuilder: tb,
		column:       &column,
	}
}

// Text 添加文本列
func (tb *TableBuilder) Text(name string) *ColumnBuilder {
	column := AdvancedColumn{
		Name: name,
		Type: TypeText,
	}
	return &ColumnBuilder{
		tableBuilder: tb,
		column:       &column,
	}
}

// Integer 添加整数列
func (tb *TableBuilder) Integer(name string) *ColumnBuilder {
	column := AdvancedColumn{
		Name: name,
		Type: TypeInt,
	}
	return &ColumnBuilder{
		tableBuilder: tb,
		column:       &column,
	}
}

// BigInteger 添加大整数列
func (tb *TableBuilder) BigInteger(name string) *ColumnBuilder {
	column := AdvancedColumn{
		Name: name,
		Type: TypeBigInt,
	}
	return &ColumnBuilder{
		tableBuilder: tb,
		column:       &column,
	}
}

// Decimal 添加小数列
func (tb *TableBuilder) Decimal(name string, precision, scale int) *ColumnBuilder {
	column := AdvancedColumn{
		Name: name,
		Type: TypeDecimal,
		Size: precision*100 + scale, // 编码精度和小数位
	}
	return &ColumnBuilder{
		tableBuilder: tb,
		column:       &column,
	}
}

// Boolean 添加布尔列
func (tb *TableBuilder) Boolean(name string) *ColumnBuilder {
	column := AdvancedColumn{
		Name: name,
		Type: TypeBoolean,
	}
	return &ColumnBuilder{
		tableBuilder: tb,
		column:       &column,
	}
}

// DateTime 添加日期时间列
func (tb *TableBuilder) DateTime(name string) *ColumnBuilder {
	column := AdvancedColumn{
		Name: name,
		Type: TypeDateTime,
	}
	return &ColumnBuilder{
		tableBuilder: tb,
		column:       &column,
	}
}

// Timestamp 添加时间戳列
func (tb *TableBuilder) Timestamp(name string) *ColumnBuilder {
	column := AdvancedColumn{
		Name: name,
		Type: TypeTimestamp,
	}
	return &ColumnBuilder{
		tableBuilder: tb,
		column:       &column,
	}
}

// Date 添加日期列
func (tb *TableBuilder) Date(name string) *ColumnBuilder {
	column := AdvancedColumn{
		Name: name,
		Type: TypeDate,
	}
	return &ColumnBuilder{
		tableBuilder: tb,
		column:       &column,
	}
}

// Json 添加JSON列
func (tb *TableBuilder) Json(name string) *ColumnBuilder {
	column := AdvancedColumn{
		Name: name,
		Type: TypeJson,
	}
	return &ColumnBuilder{
		tableBuilder: tb,
		column:       &column,
	}
}

// Enum 添加枚举列
func (tb *TableBuilder) Enum(name string, values []string) *ColumnBuilder {
	column := AdvancedColumn{
		Name:    name,
		Type:    TypeEnum,
		Default: values, // 临时存储枚举值
	}
	return &ColumnBuilder{
		tableBuilder: tb,
		column:       &column,
	}
}

// Timestamps 添加created_at和updated_at时间戳列
func (tb *TableBuilder) Timestamps() *TableBuilder {
	tb.Timestamp("created_at").Default("CURRENT_TIMESTAMP").Comment("创建时间").End()
	tb.Timestamp("updated_at").Default("CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP").Comment("更新时间").End()
	return tb
}

// SoftDeletes 添加软删除字段
func (tb *TableBuilder) SoftDeletes() *TableBuilder {
	tb.Timestamp("deleted_at").Nullable().Comment("删除时间").End()
	return tb
}

// Index 添加普通索引
func (tb *TableBuilder) Index(columns ...string) *IndexBuilder {
	name := fmt.Sprintf("idx_%s_%s", tb.tableName, strings.Join(columns, "_"))
	return &IndexBuilder{
		tableBuilder: tb,
		indexDef: &IndexDef{
			Name:    name,
			Columns: columns,
			Type:    IndexNormal,
		},
	}
}

// Unique 添加唯一索引
func (tb *TableBuilder) Unique(columns ...string) *IndexBuilder {
	name := fmt.Sprintf("uk_%s_%s", tb.tableName, strings.Join(columns, "_"))
	return &IndexBuilder{
		tableBuilder: tb,
		indexDef: &IndexDef{
			Name:    name,
			Columns: columns,
			Type:    IndexUnique,
			Unique:  true,
		},
	}
}

// ForeignKey 添加外键
func (tb *TableBuilder) ForeignKey(column string) *ForeignKeyBuilder {
	name := fmt.Sprintf("fk_%s_%s", tb.tableName, column)
	return &ForeignKeyBuilder{
		tableBuilder: tb,
		foreignKey: &ForeignKeyDef{
			Name:   name,
			Column: column,
		},
	}
}

// Engine 设置存储引擎
func (tb *TableBuilder) Engine(engine string) *TableBuilder {
	tb.options.Engine = engine
	return tb
}

// Charset 设置字符集
func (tb *TableBuilder) Charset(charset string) *TableBuilder {
	tb.options.Charset = charset
	return tb
}

// Comment 设置表注释
func (tb *TableBuilder) Comment(comment string) *TableBuilder {
	tb.options.Comment = comment
	return tb
}

// Create 创建表
func (tb *TableBuilder) Create(ctx context.Context) error {
	sql := tb.buildCreateTableSQL()
	return tb.sqlBuilder.CreateTableIfNotExists(ctx, tb.tableName, sql)
}

// buildCreateTableSQL 构建创建表的SQL
func (tb *TableBuilder) buildCreateTableSQL() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("CREATE TABLE %s (", tb.tableName))

	// 添加列定义
	var columnDefs []string
	for _, col := range tb.columns {
		columnDefs = append(columnDefs, "  "+tb.buildColumnDef(col))
	}

	// 添加索引定义
	for _, idx := range tb.indexes {
		if idx.Unique {
			columnDefs = append(columnDefs, fmt.Sprintf("  UNIQUE KEY %s (%s)",
				idx.Name, strings.Join(idx.Columns, ", ")))
		} else {
			columnDefs = append(columnDefs, fmt.Sprintf("  KEY %s (%s)",
				idx.Name, strings.Join(idx.Columns, ", ")))
		}
	}

	// 添加外键定义
	for _, fk := range tb.foreignKeys {
		fkDef := fmt.Sprintf("  CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)",
			fk.Name, fk.Column, fk.RefTable, fk.RefColumn)

		if fk.OnDelete != "" {
			fkDef += fmt.Sprintf(" ON DELETE %s", fk.OnDelete)
		}
		if fk.OnUpdate != "" {
			fkDef += fmt.Sprintf(" ON UPDATE %s", fk.OnUpdate)
		}

		columnDefs = append(columnDefs, fkDef)
	}

	parts = append(parts, strings.Join(columnDefs, ",\n"))
	parts = append(parts, ")")

	// 添加表选项
	if tb.options.Engine != "" {
		parts = append(parts, fmt.Sprintf("ENGINE=%s", tb.options.Engine))
	}
	if tb.options.Charset != "" {
		parts = append(parts, fmt.Sprintf("DEFAULT CHARSET=%s", tb.options.Charset))
	}
	if tb.options.Comment != "" {
		parts = append(parts, fmt.Sprintf("COMMENT='%s'", tb.options.Comment))
	}

	return strings.Join(parts, "\n")
}

// buildColumnDef 构建列定义
func (tb *TableBuilder) buildColumnDef(col AdvancedColumn) string {
	var parts []string
	parts = append(parts, col.Name)

	// 构建类型定义
	typeStr := string(col.Type)
	if col.Type == TypeEnum && col.Default != nil {
		// 处理枚举类型
		if values, ok := col.Default.([]string); ok {
			quotedValues := make([]string, len(values))
			for i, v := range values {
				quotedValues[i] = fmt.Sprintf("'%s'", v)
			}
			typeStr = fmt.Sprintf("ENUM(%s)", strings.Join(quotedValues, ", "))
		}
	} else if col.Size > 0 {
		if col.Type == TypeDecimal {
			// 解码精度和小数位
			precision := col.Size / 100
			scale := col.Size % 100
			typeStr = fmt.Sprintf("%s(%d,%d)", typeStr, precision, scale)
		} else {
			typeStr = fmt.Sprintf("%s(%d)", typeStr, col.Size)
		}
	}
	parts = append(parts, typeStr)

	// 添加约束
	if col.NotNull {
		parts = append(parts, "NOT NULL")
	}

	// 添加默认值
	if col.Default != nil && col.Type != TypeEnum {
		switch v := col.Default.(type) {
		case string:
			if v == "CURRENT_TIMESTAMP" || strings.Contains(v, "CURRENT_TIMESTAMP") {
				parts = append(parts, fmt.Sprintf("DEFAULT %s", v))
			} else {
				parts = append(parts, fmt.Sprintf("DEFAULT '%s'", v))
			}
		case int, int64, float64:
			parts = append(parts, fmt.Sprintf("DEFAULT %v", v))
		case bool:
			if v {
				parts = append(parts, "DEFAULT TRUE")
			} else {
				parts = append(parts, "DEFAULT FALSE")
			}
		}
	}

	if col.AutoIncr {
		parts = append(parts, "AUTO_INCREMENT")
	}

	if col.PrimaryKey {
		parts = append(parts, "PRIMARY KEY")
	}

	if col.Unique && !col.PrimaryKey {
		parts = append(parts, "UNIQUE")
	}

	if col.Comment != "" {
		parts = append(parts, fmt.Sprintf("COMMENT '%s'", col.Comment))
	}

	return strings.Join(parts, " ")
}

// ColumnBuilder 列构建器（链式调用）
type ColumnBuilder struct {
	tableBuilder *TableBuilder
	column       *AdvancedColumn
}

// NotNull 设置为非空
func (cb *ColumnBuilder) NotNull() *ColumnBuilder {
	cb.column.NotNull = true
	return cb
}

// Nullable 设置为可空
func (cb *ColumnBuilder) Nullable() *ColumnBuilder {
	cb.column.NotNull = false
	return cb
}

// Default 设置默认值
func (cb *ColumnBuilder) Default(value interface{}) *ColumnBuilder {
	cb.column.Default = value
	return cb
}

// Unique 设置为唯一
func (cb *ColumnBuilder) Unique() *ColumnBuilder {
	cb.column.Unique = true
	return cb
}

// Comment 设置注释
func (cb *ColumnBuilder) Comment(comment string) *ColumnBuilder {
	cb.column.Comment = comment
	return cb
}

// After 在指定列之后添加
func (cb *ColumnBuilder) After(columnName string) *ColumnBuilder {
	cb.column.After = columnName
	return cb
}

// End 结束列定义，添加到表中
func (cb *ColumnBuilder) End() *TableBuilder {
	cb.tableBuilder.columns = append(cb.tableBuilder.columns, *cb.column)
	return cb.tableBuilder
}

// IndexBuilder 索引构建器
type IndexBuilder struct {
	tableBuilder *TableBuilder
	indexDef     *IndexDef
}

// Name 设置索引名称
func (ib *IndexBuilder) Name(name string) *IndexBuilder {
	ib.indexDef.Name = name
	return ib
}

// End 结束索引定义
func (ib *IndexBuilder) End() *TableBuilder {
	ib.tableBuilder.indexes = append(ib.tableBuilder.indexes, *ib.indexDef)
	return ib.tableBuilder
}

// ForeignKeyBuilder 外键构建器
type ForeignKeyBuilder struct {
	tableBuilder *TableBuilder
	foreignKey   *ForeignKeyDef
}

// References 设置引用表和列
func (fkb *ForeignKeyBuilder) References(table, column string) *ForeignKeyBuilder {
	fkb.foreignKey.RefTable = table
	fkb.foreignKey.RefColumn = column
	return fkb
}

// OnDelete 设置删除动作
func (fkb *ForeignKeyBuilder) OnDelete(action ReferenceAction) *ForeignKeyBuilder {
	fkb.foreignKey.OnDelete = action
	return fkb
}

// OnUpdate 设置更新动作
func (fkb *ForeignKeyBuilder) OnUpdate(action ReferenceAction) *ForeignKeyBuilder {
	fkb.foreignKey.OnUpdate = action
	return fkb
}

// End 结束外键定义
func (fkb *ForeignKeyBuilder) End() *TableBuilder {
	fkb.tableBuilder.foreignKeys = append(fkb.tableBuilder.foreignKeys, *fkb.foreignKey)
	return fkb.tableBuilder
}

// CreateTableFromStruct 从结构体创建表
func CreateTableFromStruct(sqlBuilder *SQLBuilder, tableName string, structType interface{}) *TableBuilder {
	tb := NewTableBuilder(sqlBuilder, tableName)

	t := reflect.TypeOf(structType)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// 跳过非导出字段
		if !field.IsExported() {
			continue
		}

		columnName := getColumnName(field)
		if columnName == "-" {
			continue
		}

		column := AdvancedColumn{
			Name: columnName,
		}

		// 根据Go类型推断数据库类型
		switch field.Type.Kind() {
		case reflect.Int, reflect.Int32:
			column.Type = TypeInt
		case reflect.Int64:
			column.Type = TypeBigInt
		case reflect.String:
			size := getColumnSize(field)
			if size > 0 && size <= 255 {
				column.Type = TypeVarchar
				column.Size = size
			} else {
				column.Type = TypeText
			}
		case reflect.Bool:
			column.Type = TypeBoolean
		case reflect.Float32, reflect.Float64:
			column.Type = TypeDouble
		default:
			if field.Type == reflect.TypeOf(time.Time{}) {
				column.Type = TypeDateTime
			} else {
				column.Type = TypeText // 默认类型
			}
		}

		// 处理标签
		if tag := field.Tag.Get("db"); tag != "" {
			parseDBTag(&column, tag)
		}

		tb.columns = append(tb.columns, column)
	}

	return tb
}

// 辅助函数
func getColumnName(field reflect.StructField) string {
	if tag := field.Tag.Get("db"); tag != "" {
		parts := strings.Split(tag, ",")
		if len(parts) > 0 && parts[0] != "" {
			return parts[0]
		}
	}

	// 将驼峰命名转换为下划线命名
	name := field.Name
	var result strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

func getColumnSize(field reflect.StructField) int {
	if tag := field.Tag.Get("size"); tag != "" {
		if size := parseIntTag(tag); size > 0 {
			return size
		}
	}
	return 0
}

func parseDBTag(column *AdvancedColumn, tag string) {
	parts := strings.Split(tag, ",")
	for _, part := range parts[1:] { // 跳过第一个部分（列名）
		part = strings.TrimSpace(part)
		switch {
		case part == "primary_key":
			column.PrimaryKey = true
			column.NotNull = true
		case part == "auto_increment":
			column.AutoIncr = true
		case part == "not_null":
			column.NotNull = true
		case part == "unique":
			column.Unique = true
		case strings.HasPrefix(part, "default:"):
			defaultValue := strings.TrimPrefix(part, "default:")
			column.Default = defaultValue
		case strings.HasPrefix(part, "comment:"):
			comment := strings.TrimPrefix(part, "comment:")
			column.Comment = comment
		}
	}
}

func parseIntTag(tag string) int {
	var result int
	fmt.Sscanf(tag, "%d", &result)
	return result
}
