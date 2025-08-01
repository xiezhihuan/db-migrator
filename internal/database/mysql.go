package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/xiezhihuan/db-migrator/internal/types"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLDB MySQL数据库实现
type MySQLDB struct {
	db *sql.DB
}

// NewMySQLDB 创建MySQL数据库连接
func NewMySQLDB(config types.DatabaseConfig) (*MySQLDB, error) {
	// 构建连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=Local",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		config.Charset,
	)

	db, err := sql.Open(config.Driver, dsn)
	if err != nil {
		return nil, &types.Error{
			Code:    types.ErrCodeDatabaseConnection,
			Message: "failed to open database connection",
			Cause:   err,
		}
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, &types.Error{
			Code:    types.ErrCodeDatabaseConnection,
			Message: "failed to ping database",
			Cause:   err,
		}
	}

	// 设置连接池参数
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	return &MySQLDB{db: db}, nil
}

// Exec 执行SQL语句
func (m *MySQLDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return m.db.Exec(query, args...)
}

// Query 查询数据
func (m *MySQLDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return m.db.Query(query, args...)
}

// QueryRow 查询单行数据
func (m *MySQLDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return m.db.QueryRow(query, args...)
}

// Begin 开始事务
func (m *MySQLDB) Begin() (*sql.Tx, error) {
	return m.db.Begin()
}

// Close 关闭连接
func (m *MySQLDB) Close() error {
	return m.db.Close()
}

// GetRawDB 获取原始数据库连接（用于特殊操作）
func (m *MySQLDB) GetRawDB() *sql.DB {
	return m.db
}
