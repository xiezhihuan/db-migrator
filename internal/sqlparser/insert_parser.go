package sqlparser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/xiezhihuan/db-migrator/internal/types"
)

// InsertParser INSERT语句解析器实现
type InsertParser struct {
	delimiter string
	variables map[string]interface{} // 存储MySQL变量
}

// NewInsertParser 创建新的INSERT解析器
func NewInsertParser() *InsertParser {
	return &InsertParser{
		delimiter: ";",
		variables: make(map[string]interface{}),
	}
}

// ParseInsertFile 解析包含INSERT语句的SQL文件
func (p *InsertParser) ParseInsertFile(filePath string) ([]types.InsertStatement, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法打开文件 %s: %v", filePath, err)
	}
	defer file.Close()

	var statements []types.InsertStatement
	var currentStatement strings.Builder
	var inMultiLineComment bool
	var inString bool
	var stringChar rune

	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		// 处理分隔符设置
		if strings.HasPrefix(strings.TrimSpace(strings.ToUpper(line)), "DELIMITER") {
			p.delimiter = strings.TrimSpace(line[9:])
			continue
		}

		processedLine := p.processLine(line, &inMultiLineComment, &inString, &stringChar)
		if processedLine == "" {
			continue
		}

		currentStatement.WriteString(processedLine)
		currentStatement.WriteString(" ")

		// 检查是否遇到分隔符
		if p.isStatementEnd(processedLine, inString) {
			statementText := strings.TrimSpace(currentStatement.String())
			statementText = strings.TrimSuffix(statementText, p.delimiter)
			statementText = strings.TrimSpace(statementText)

			if statementText != "" {
				// 尝试解析SET变量语句
				if p.parseSetStatement(statementText) {
					currentStatement.Reset()
					continue
				}

				// 解析INSERT语句
				stmt, err := p.parseInsertStatement(statementText, lineNumber)
				if err != nil {
					return nil, fmt.Errorf("第%d行解析错误: %v", lineNumber, err)
				}
				if stmt != nil {
					statements = append(statements, *stmt)
				}
			}

			currentStatement.Reset()
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取文件错误: %v", err)
	}

	return statements, nil
}

// processLine 处理单行，移除注释（与DDL解析器相同的逻辑）
func (p *InsertParser) processLine(line string, inMultiLineComment *bool, inString *bool, stringChar *rune) string {
	var result strings.Builder
	runes := []rune(line)

	for i, r := range runes {
		// 处理多行注释
		if *inMultiLineComment {
			if r == '*' && i+1 < len(runes) && runes[i+1] == '/' {
				*inMultiLineComment = false
				i++ // 跳过下一个字符
			}
			continue
		}

		// 处理字符串
		if *inString {
			result.WriteRune(r)
			if r == *stringChar && (i == 0 || runes[i-1] != '\\') {
				*inString = false
			}
			continue
		}

		// 检查字符串开始
		if r == '\'' || r == '"' || r == '`' {
			*inString = true
			*stringChar = r
			result.WriteRune(r)
			continue
		}

		// 检查单行注释
		if r == '-' && i+1 < len(runes) && runes[i+1] == '-' {
			break
		}

		// 检查多行注释开始
		if r == '/' && i+1 < len(runes) && runes[i+1] == '*' {
			*inMultiLineComment = true
			i++ // 跳过下一个字符
			continue
		}

		result.WriteRune(r)
	}

	return strings.TrimSpace(result.String())
}

// isStatementEnd 检查是否是语句结束
func (p *InsertParser) isStatementEnd(line string, inString bool) bool {
	if inString {
		return false
	}
	return strings.HasSuffix(strings.TrimSpace(line), p.delimiter)
}

// parseInsertStatement 解析单个INSERT语句
func (p *InsertParser) parseInsertStatement(statement string, lineNumber int) (*types.InsertStatement, error) {
	statement = strings.TrimSpace(statement)
	if statement == "" {
		return nil, nil
	}

	upperStatement := strings.ToUpper(statement)

	// 只处理INSERT语句
	if !strings.HasPrefix(upperStatement, "INSERT") {
		return nil, nil // 忽略非INSERT语句
	}

	return p.parseInsert(statement, lineNumber)
}

// parseInsert 解析INSERT语句
func (p *InsertParser) parseInsert(statement string, lineNumber int) (*types.InsertStatement, error) {
	// 匹配 INSERT INTO table_name (columns) VALUES (values)
	// 支持多种格式：
	// 1. INSERT INTO table (col1, col2) VALUES (val1, val2)
	// 2. INSERT INTO table VALUES (val1, val2)
	// 3. INSERT IGNORE INTO table ...
	// 4. INSERT INTO table (col1, col2) VALUES (val1, val2), (val3, val4)

	// 提取表名
	tableName, err := p.extractTableName(statement)
	if err != nil {
		return nil, err
	}

	// 提取列名
	columns, err := p.extractColumns(statement)
	if err != nil {
		return nil, err
	}

	// 提取值
	values, err := p.extractValues(statement)
	if err != nil {
		return nil, err
	}

	return &types.InsertStatement{
		TableName:  tableName,
		Columns:    columns,
		Values:     values,
		Statement:  statement,
		LineNumber: lineNumber,
	}, nil
}

// extractTableName 提取表名
func (p *InsertParser) extractTableName(statement string) (string, error) {
	// 匹配 INSERT [IGNORE] INTO table_name
	re := regexp.MustCompile(`(?i)INSERT\s+(?:IGNORE\s+)?INTO\s+(?:\x60?(\w+)\x60?\.)?(?:\x60?(\w+)\x60?)`)
	matches := re.FindStringSubmatch(statement)

	if len(matches) < 3 {
		return "", fmt.Errorf("无法解析表名")
	}

	tableName := matches[2]
	if matches[1] != "" {
		tableName = matches[1] + "." + tableName
	}

	return tableName, nil
}

// extractColumns 提取列名
func (p *InsertParser) extractColumns(statement string) ([]string, error) {
	// 查找 (column1, column2, ...) 部分
	re := regexp.MustCompile(`(?i)INTO\s+(?:\x60?\w+\x60?\.)?(?:\x60?\w+\x60?)\s*\(\s*([^)]+)\)\s+VALUES`)
	matches := re.FindStringSubmatch(statement)

	if len(matches) < 2 {
		// 没有指定列名，返回空列表
		return []string{}, nil
	}

	columnsPart := strings.TrimSpace(matches[1])
	if columnsPart == "" {
		return []string{}, nil
	}

	// 分割列名
	var columns []string
	for _, col := range strings.Split(columnsPart, ",") {
		col = strings.TrimSpace(col)
		col = strings.Trim(col, "`") // 移除反引号
		if col != "" {
			columns = append(columns, col)
		}
	}

	return columns, nil
}

// extractValues 提取值
func (p *InsertParser) extractValues(statement string) ([][]interface{}, error) {
	// 查找 VALUES 后的部分
	re := regexp.MustCompile(`(?i)VALUES\s+(.+)$`)
	matches := re.FindStringSubmatch(statement)

	if len(matches) < 2 {
		return nil, fmt.Errorf("无法找到VALUES子句")
	}

	valuesPart := strings.TrimSpace(matches[1])
	return p.parseValueGroups(valuesPart)
}

// parseValueGroups 解析值组
func (p *InsertParser) parseValueGroups(valuesPart string) ([][]interface{}, error) {
	var result [][]interface{}
	var current []interface{}
	var currentValue strings.Builder
	var inString bool
	var stringChar rune
	var parenDepth int
	var inValueGroup bool

	runes := []rune(valuesPart)

	for i, r := range runes {
		if inString {
			currentValue.WriteRune(r)
			if r == stringChar && (i == 0 || runes[i-1] != '\\') {
				inString = false
			}
			continue
		}

		switch r {
		case '\'', '"':
			inString = true
			stringChar = r
			currentValue.WriteRune(r)

		case '(':
			if !inValueGroup {
				inValueGroup = true
				parenDepth = 1
			} else {
				parenDepth++
				currentValue.WriteRune(r)
			}

		case ')':
			if inValueGroup {
				parenDepth--
				if parenDepth == 0 {
					// 处理最后一个值
					if currentValue.Len() > 0 {
						value, err := p.parseValue(strings.TrimSpace(currentValue.String()))
						if err != nil {
							return nil, err
						}
						current = append(current, value)
						currentValue.Reset()
					}

					result = append(result, current)
					current = []interface{}{}
					inValueGroup = false
				} else {
					currentValue.WriteRune(r)
				}
			}

		case ',':
			if inValueGroup && parenDepth == 1 {
				// 值分隔符
				value, err := p.parseValue(strings.TrimSpace(currentValue.String()))
				if err != nil {
					return nil, err
				}
				current = append(current, value)
				currentValue.Reset()
			} else if !inValueGroup {
				// 值组分隔符，忽略
			} else {
				currentValue.WriteRune(r)
			}

		default:
			if inValueGroup {
				currentValue.WriteRune(r)
			}
		}
	}

	return result, nil
}

// parseValue 解析单个值
func (p *InsertParser) parseValue(valueStr string) (interface{}, error) {
	valueStr = strings.TrimSpace(valueStr)

	if valueStr == "" {
		return nil, fmt.Errorf("空值")
	}

	// NULL值
	if strings.ToUpper(valueStr) == "NULL" {
		return nil, nil
	}

	// 处理MySQL变量 @variableName
	if strings.HasPrefix(valueStr, "@") {
		varName := strings.ToLower(valueStr[1:])
		if value, exists := p.variables[varName]; exists {
			return value, nil
		}
		return nil, fmt.Errorf("未定义的变量: %s", valueStr)
	}

	// 字符串值（带引号）
	if (strings.HasPrefix(valueStr, "'") && strings.HasSuffix(valueStr, "'")) ||
		(strings.HasPrefix(valueStr, "\"") && strings.HasSuffix(valueStr, "\"")) {
		return valueStr[1 : len(valueStr)-1], nil
	}

	// 数字值
	if intVal, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		return intVal, nil
	}

	if floatVal, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return floatVal, nil
	}

	// 布尔值
	if strings.ToUpper(valueStr) == "TRUE" {
		return true, nil
	}
	if strings.ToUpper(valueStr) == "FALSE" {
		return false, nil
	}

	// 其他情况作为字符串处理
	return valueStr, nil
}

// parseSetStatement 解析SET变量语句
func (p *InsertParser) parseSetStatement(statement string) bool {
	statement = strings.TrimSpace(statement)

	// 匹配 SET @variable = value 格式
	setRegex := regexp.MustCompile(`(?i)^SET\s+@(\w+)\s*=\s*(.+)$`)
	matches := setRegex.FindStringSubmatch(statement)

	if len(matches) == 3 {
		varName := strings.ToLower(matches[1])
		valueExpr := strings.TrimSpace(matches[2])

		// 计算表达式值
		value := p.evaluateExpression(valueExpr)
		p.variables[varName] = value

		return true
	}

	return false
}

// evaluateExpression 计算表达式值
func (p *InsertParser) evaluateExpression(expr string) interface{} {
	expr = strings.TrimSpace(expr)
	upperExpr := strings.ToUpper(expr)

	// 处理 UNIX_TIMESTAMP(NOW()) 函数
	if upperExpr == "UNIX_TIMESTAMP(NOW())" {
		return time.Now().Unix()
	}

	// 处理 NOW() 函数
	if upperExpr == "NOW()" {
		return time.Now().Format("2006-01-02 15:04:05")
	}

	// 处理数字
	if num, err := strconv.ParseInt(expr, 10, 64); err == nil {
		return num
	}

	if num, err := strconv.ParseFloat(expr, 64); err == nil {
		return num
	}

	// 处理字符串（去掉引号）
	if (strings.HasPrefix(expr, "'") && strings.HasSuffix(expr, "'")) ||
		(strings.HasPrefix(expr, "\"") && strings.HasSuffix(expr, "\"")) {
		return expr[1 : len(expr)-1]
	}

	// 默认返回字符串
	return expr
}

// GetVariables 获取解析到的变量
func (p *InsertParser) GetVariables() map[string]interface{} {
	return p.variables
}

// ValidateInsertStatements 验证INSERT语句
func (p *InsertParser) ValidateInsertStatements(statements []types.InsertStatement) error {
	if len(statements) == 0 {
		return fmt.Errorf("没有找到有效的INSERT语句")
	}

	for i, stmt := range statements {
		if stmt.TableName == "" {
			return fmt.Errorf("第%d个语句缺少表名", i+1)
		}

		if len(stmt.Values) == 0 {
			return fmt.Errorf("第%d个语句缺少插入值", i+1)
		}

		// 如果指定了列名，检查列数和值数是否匹配
		if len(stmt.Columns) > 0 {
			for j, values := range stmt.Values {
				if len(values) != len(stmt.Columns) {
					return fmt.Errorf("第%d个语句第%d组值的列数不匹配：期望%d列，实际%d列",
						i+1, j+1, len(stmt.Columns), len(values))
				}
			}
		}
	}

	return nil
}

// ExtractTableNames 提取所有表名
func (p *InsertParser) ExtractTableNames(statements []types.InsertStatement) []string {
	tableSet := make(map[string]bool)
	var tables []string

	for _, stmt := range statements {
		if !tableSet[stmt.TableName] {
			tableSet[stmt.TableName] = true
			tables = append(tables, stmt.TableName)
		}
	}

	return tables
}
