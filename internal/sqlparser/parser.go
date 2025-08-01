package sqlparser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/xiezhihuan/db-migrator/internal/types"
)

// Parser SQL解析器实现
type Parser struct {
	delimiter string
}

// NewParser 创建新的SQL解析器
func NewParser() *Parser {
	return &Parser{
		delimiter: ";",
	}
}

// ParseFile 解析SQL文件
func (p *Parser) ParseFile(filePath string) ([]types.SQLStatement, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法打开文件 %s: %v", filePath, err)
	}
	defer file.Close()

	var statements []types.SQLStatement
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
				stmt, err := p.parseStatement(statementText)
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

// processLine 处理单行，移除注释
func (p *Parser) processLine(line string, inMultiLineComment *bool, inString *bool, stringChar *rune) string {
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
func (p *Parser) isStatementEnd(line string, inString bool) bool {
	if inString {
		return false
	}
	return strings.HasSuffix(strings.TrimSpace(line), p.delimiter)
}

// parseStatement 解析单个SQL语句
func (p *Parser) parseStatement(statement string) (*types.SQLStatement, error) {
	statement = strings.TrimSpace(statement)
	if statement == "" {
		return nil, nil
	}

	upperStatement := strings.ToUpper(statement)

	// CREATE TABLE
	if strings.HasPrefix(upperStatement, "CREATE TABLE") {
		return p.parseCreateTable(statement)
	}

	// CREATE VIEW
	if strings.HasPrefix(upperStatement, "CREATE VIEW") ||
		strings.Contains(upperStatement, "CREATE OR REPLACE VIEW") {
		return p.parseCreateView(statement)
	}

	// CREATE PROCEDURE/FUNCTION
	if strings.HasPrefix(upperStatement, "CREATE PROCEDURE") ||
		strings.HasPrefix(upperStatement, "CREATE FUNCTION") ||
		strings.Contains(upperStatement, "CREATE OR REPLACE PROCEDURE") ||
		strings.Contains(upperStatement, "CREATE OR REPLACE FUNCTION") {
		return p.parseCreateProcedure(statement)
	}

	// CREATE TRIGGER
	if strings.HasPrefix(upperStatement, "CREATE TRIGGER") {
		return p.parseCreateTrigger(statement)
	}

	// CREATE INDEX
	if strings.HasPrefix(upperStatement, "CREATE INDEX") ||
		strings.HasPrefix(upperStatement, "CREATE UNIQUE INDEX") {
		return p.parseCreateIndex(statement)
	}

	// 其他DDL语句
	if strings.HasPrefix(upperStatement, "CREATE") {
		return &types.SQLStatement{
			Type:         "CREATE_OTHER",
			Name:         "unknown",
			Statement:    statement,
			Dependencies: []string{},
		}, nil
	}

	return nil, nil // 忽略非CREATE语句
}

// parseCreateTable 解析CREATE TABLE语句
func (p *Parser) parseCreateTable(statement string) (*types.SQLStatement, error) {
	re := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(?:\x60?(\w+)\x60?\.)?(?:\x60?(\w+)\x60?)`)
	matches := re.FindStringSubmatch(statement)

	if len(matches) < 3 {
		return nil, fmt.Errorf("无法解析表名")
	}

	tableName := matches[2]
	if matches[1] != "" {
		tableName = matches[1] + "." + tableName
	}

	// 解析外键依赖
	dependencies := p.extractForeignKeyDependencies(statement)

	return &types.SQLStatement{
		Type:         "CREATE_TABLE",
		Name:         tableName,
		Statement:    statement,
		Dependencies: dependencies,
	}, nil
}

// parseCreateView 解析CREATE VIEW语句
func (p *Parser) parseCreateView(statement string) (*types.SQLStatement, error) {
	re := regexp.MustCompile(`(?i)CREATE\s+(?:OR\s+REPLACE\s+)?VIEW\s+(?:\x60?(\w+)\x60?\.)?(?:\x60?(\w+)\x60?)`)
	matches := re.FindStringSubmatch(statement)

	if len(matches) < 3 {
		return nil, fmt.Errorf("无法解析视图名")
	}

	viewName := matches[2]
	if matches[1] != "" {
		viewName = matches[1] + "." + viewName
	}

	// 解析视图依赖的表
	dependencies := p.extractTableDependencies(statement)

	return &types.SQLStatement{
		Type:         "CREATE_VIEW",
		Name:         viewName,
		Statement:    statement,
		Dependencies: dependencies,
	}, nil
}

// parseCreateProcedure 解析CREATE PROCEDURE/FUNCTION语句
func (p *Parser) parseCreateProcedure(statement string) (*types.SQLStatement, error) {
	re := regexp.MustCompile(`(?i)CREATE\s+(?:OR\s+REPLACE\s+)?(PROCEDURE|FUNCTION)\s+(?:\x60?(\w+)\x60?\.)?(?:\x60?(\w+)\x60?)`)
	matches := re.FindStringSubmatch(statement)

	if len(matches) < 4 {
		return nil, fmt.Errorf("无法解析存储过程/函数名")
	}

	objType := strings.ToUpper(matches[1])
	objName := matches[3]
	if matches[2] != "" {
		objName = matches[2] + "." + objName
	}

	return &types.SQLStatement{
		Type:         "CREATE_" + objType,
		Name:         objName,
		Statement:    statement,
		Dependencies: []string{}, // 存储过程依赖比较复杂，暂时不解析
	}, nil
}

// parseCreateTrigger 解析CREATE TRIGGER语句
func (p *Parser) parseCreateTrigger(statement string) (*types.SQLStatement, error) {
	re := regexp.MustCompile(`(?i)CREATE\s+TRIGGER\s+(?:\x60?(\w+)\x60?\.)?(?:\x60?(\w+)\x60?)\s+(?:BEFORE|AFTER|INSTEAD\s+OF)\s+(?:INSERT|UPDATE|DELETE)(?:\s+OR\s+(?:INSERT|UPDATE|DELETE))*\s+ON\s+(?:\x60?(\w+)\x60?\.)?(?:\x60?(\w+)\x60?)`)
	matches := re.FindStringSubmatch(statement)

	if len(matches) < 5 {
		return nil, fmt.Errorf("无法解析触发器信息")
	}

	triggerName := matches[2]
	if matches[1] != "" {
		triggerName = matches[1] + "." + triggerName
	}

	tableName := matches[4]
	if matches[3] != "" {
		tableName = matches[3] + "." + tableName
	}

	return &types.SQLStatement{
		Type:         "CREATE_TRIGGER",
		Name:         triggerName,
		Statement:    statement,
		Dependencies: []string{tableName}, // 触发器依赖表
	}, nil
}

// parseCreateIndex 解析CREATE INDEX语句
func (p *Parser) parseCreateIndex(statement string) (*types.SQLStatement, error) {
	re := regexp.MustCompile(`(?i)CREATE\s+(?:UNIQUE\s+)?INDEX\s+(?:\x60?(\w+)\x60?\.)?(?:\x60?(\w+)\x60?)\s+ON\s+(?:\x60?(\w+)\x60?\.)?(?:\x60?(\w+)\x60?)`)
	matches := re.FindStringSubmatch(statement)

	if len(matches) < 5 {
		return nil, fmt.Errorf("无法解析索引信息")
	}

	indexName := matches[2]
	if matches[1] != "" {
		indexName = matches[1] + "." + indexName
	}

	tableName := matches[4]
	if matches[3] != "" {
		tableName = matches[3] + "." + tableName
	}

	return &types.SQLStatement{
		Type:         "CREATE_INDEX",
		Name:         indexName,
		Statement:    statement,
		Dependencies: []string{tableName}, // 索引依赖表
	}, nil
}

// extractForeignKeyDependencies 提取外键依赖
func (p *Parser) extractForeignKeyDependencies(statement string) []string {
	var dependencies []string

	// 匹配 FOREIGN KEY ... REFERENCES table_name
	re := regexp.MustCompile(`(?i)REFERENCES\s+(?:\x60?(\w+)\x60?\.)?(?:\x60?(\w+)\x60?)`)
	matches := re.FindAllStringSubmatch(statement, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			tableName := match[2]
			if match[1] != "" {
				tableName = match[1] + "." + tableName
			}
			dependencies = append(dependencies, tableName)
		}
	}

	return dependencies
}

// extractTableDependencies 提取表依赖（用于视图）
func (p *Parser) extractTableDependencies(statement string) []string {
	var dependencies []string

	// 从FROM和JOIN子句中提取表名
	re := regexp.MustCompile(`(?i)(?:FROM|JOIN)\s+(?:\x60?(\w+)\x60?\.)?(?:\x60?(\w+)\x60?)`)
	matches := re.FindAllStringSubmatch(statement, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			tableName := match[2]
			if match[1] != "" {
				tableName = match[1] + "." + tableName
			}
			dependencies = append(dependencies, tableName)
		}
	}

	return dependencies
}

// ValidateStatements 验证语句
func (p *Parser) ValidateStatements(statements []types.SQLStatement) error {
	// 检查重复定义
	names := make(map[string]bool)
	for _, stmt := range statements {
		key := stmt.Type + ":" + stmt.Name
		if names[key] {
			return fmt.Errorf("重复定义: %s %s", stmt.Type, stmt.Name)
		}
		names[key] = true
	}

	return nil
}

// SortByDependencies 按依赖关系排序
func (p *Parser) SortByDependencies(statements []types.SQLStatement) ([]types.SQLStatement, error) {
	// 创建依赖图
	graph := make(map[string][]string)
	inDegree := make(map[string]int)
	stmtMap := make(map[string]types.SQLStatement)

	// 初始化
	for _, stmt := range statements {
		key := stmt.Name
		graph[key] = []string{}
		inDegree[key] = 0
		stmtMap[key] = stmt
	}

	// 构建依赖关系
	for _, stmt := range statements {
		for _, dep := range stmt.Dependencies {
			if _, exists := stmtMap[dep]; exists {
				graph[dep] = append(graph[dep], stmt.Name)
				inDegree[stmt.Name]++
			}
		}
	}

	// 拓扑排序
	var queue []string
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	var result []types.SQLStatement
	processed := 0

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		result = append(result, stmtMap[current])
		processed++

		for _, neighbor := range graph[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if processed != len(statements) {
		return nil, fmt.Errorf("检测到循环依赖")
	}

	// 按类型优先级重新排序（表->视图->存储过程->触发器->索引）
	typeOrder := map[string]int{
		"CREATE_TABLE":     1,
		"CREATE_VIEW":      2,
		"CREATE_PROCEDURE": 3,
		"CREATE_FUNCTION":  3,
		"CREATE_TRIGGER":   4,
		"CREATE_INDEX":     5,
		"CREATE_OTHER":     6,
	}

	sort.Slice(result, func(i, j int) bool {
		if typeOrder[result[i].Type] != typeOrder[result[j].Type] {
			return typeOrder[result[i].Type] < typeOrder[result[j].Type]
		}
		return result[i].Name < result[j].Name
	})

	return result, nil
}
