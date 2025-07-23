package logging

import (
	"context"
	"db-migrator/internal/builder"
	"db-migrator/internal/checker"
	"db-migrator/internal/types"
)

// CreateLoggingSystemMigration 创建日志系统迁移
type CreateLoggingSystemMigration struct{}

func (m *CreateLoggingSystemMigration) Version() string {
	return "001"
}

func (m *CreateLoggingSystemMigration) Description() string {
	return "创建日志系统 - 应用日志、错误日志、访问日志、性能监控"
}

func (m *CreateLoggingSystemMigration) Up(ctx context.Context, db types.DB) error {
	checker := checker.NewMySQLChecker(db, "logging_db")
	advancedBuilder := builder.NewAdvancedBuilder(checker, db)

	// 1. 创建应用日志表
	err := advancedBuilder.Table("application_logs").
		ID().
		String("trace_id", 64).Nullable().Comment("链路追踪ID").End().
		String("span_id", 32).Nullable().Comment("跨度ID").End().
		Enum("level", []string{"debug", "info", "warning", "error", "critical"}).NotNull().Comment("日志级别").End().
		String("logger", 100).NotNull().Comment("日志记录器").End().
		String("message", 1000).NotNull().Comment("日志消息").End().
		Json("context").Nullable().Comment("上下文数据").End().
		String("user_id", 36).Nullable().Comment("用户ID").End().
		String("session_id", 40).Nullable().Comment("会话ID").End().
		String("ip_address", 45).Nullable().Comment("IP地址").End().
		String("user_agent", 500).Nullable().Comment("用户代理").End().
		String("url", 2000).Nullable().Comment("请求URL").End().
		String("method", 10).Nullable().Comment("HTTP方法").End().
		String("service", 100).Nullable().Comment("服务名称").End().
		String("environment", 50).Default("production").Comment("环境").End().
		String("version", 50).Nullable().Comment("应用版本").End().
		String("hostname", 255).Nullable().Comment("主机名").End().
		Integer("process_id").Nullable().Comment("进程ID").End().
		String("thread_id", 50).Nullable().Comment("线程ID").End().
		Json("tags").Nullable().Comment("标签").End().
		Timestamp("logged_at").Default("CURRENT_TIMESTAMP").Comment("记录时间").End().
		Index("trace_id").End().
		Index("level").End().
		Index("logger").End().
		Index("user_id").End().
		Index("service").End().
		Index("environment").End().
		Index("logged_at").End().
		Engine("InnoDB").
		Comment("应用日志表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 2. 创建错误日志表
	err = advancedBuilder.Table("error_logs").
		ID().
		String("trace_id", 64).Nullable().Comment("链路追踪ID").End().
		String("error_id", 64).NotNull().Comment("错误ID").End().
		String("error_code", 50).Nullable().Comment("错误代码").End().
		String("error_type", 100).NotNull().Comment("错误类型").End().
		String("message", 1000).NotNull().Comment("错误消息").End().
		Text("stack_trace").Nullable().Comment("堆栈跟踪").End().
		String("file", 500).Nullable().Comment("文件路径").End().
		Integer("line").Nullable().Comment("行号").End().
		String("function", 200).Nullable().Comment("函数名").End().
		Json("context").Nullable().Comment("错误上下文").End().
		String("user_id", 36).Nullable().Comment("用户ID").End().
		String("session_id", 40).Nullable().Comment("会话ID").End().
		String("ip_address", 45).Nullable().Comment("IP地址").End().
		String("url", 2000).Nullable().Comment("请求URL").End().
		String("method", 10).Nullable().Comment("HTTP方法").End().
		Integer("status_code").Nullable().Comment("HTTP状态码").End().
		String("service", 100).Nullable().Comment("服务名称").End().
		String("environment", 50).Default("production").Comment("环境").End().
		String("version", 50).Nullable().Comment("应用版本").End().
		Integer("count").Default(1).Comment("出现次数").End().
		Timestamp("first_occurred").Default("CURRENT_TIMESTAMP").Comment("首次出现时间").End().
		Timestamp("last_occurred").Default("CURRENT_TIMESTAMP").Comment("最后出现时间").End().
		Boolean("is_resolved").Default(false).Comment("是否已解决").End().
		String("resolved_by", 36).Nullable().Comment("解决人").End().
		Timestamp("resolved_at").Nullable().Comment("解决时间").End().
		Text("resolution_notes").Nullable().Comment("解决备注").End().
		Index("trace_id").End().
		Index("error_id").End().
		Index("error_type").End().
		Index("user_id").End().
		Index("service").End().
		Index("is_resolved").End().
		Index("first_occurred").End().
		Index("last_occurred").End().
		Engine("InnoDB").
		Comment("错误日志表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 3. 创建访问日志表
	err = advancedBuilder.Table("access_logs").
		ID().
		String("request_id", 64).NotNull().Comment("请求ID").End().
		String("user_id", 36).Nullable().Comment("用户ID").End().
		String("session_id", 40).Nullable().Comment("会话ID").End().
		String("ip_address", 45).NotNull().Comment("IP地址").End().
		String("user_agent", 1000).Nullable().Comment("用户代理").End().
		String("method", 10).NotNull().Comment("HTTP方法").End().
		String("url", 2000).NotNull().Comment("请求URL").End().
		String("path", 1000).NotNull().Comment("路径").End().
		String("query_string", 2000).Nullable().Comment("查询字符串").End().
		Json("headers").Nullable().Comment("请求头").End().
		Text("request_body").Nullable().Comment("请求体").End().
		Integer("status_code").NotNull().Comment("响应状态码").End().
		Integer("response_size").Default(0).Comment("响应大小(字节)").End().
		Integer("response_time").NotNull().Comment("响应时间(毫秒)").End().
		String("referer", 2000).Nullable().Comment("来源页面").End().
		String("controller", 100).Nullable().Comment("控制器").End().
		String("action", 100).Nullable().Comment("动作").End().
		Json("route_params").Nullable().Comment("路由参数").End().
		String("service", 100).Nullable().Comment("服务名称").End().
		String("environment", 50).Default("production").Comment("环境").End().
		String("load_balancer", 50).Nullable().Comment("负载均衡器").End().
		String("server", 100).Nullable().Comment("服务器").End().
		Boolean("is_bot").Default(false).Comment("是否机器人").End().
		Boolean("is_mobile").Default(false).Comment("是否移动设备").End().
		String("country", 50).Nullable().Comment("国家").End().
		String("city", 100).Nullable().Comment("城市").End().
		Timestamp("accessed_at").Default("CURRENT_TIMESTAMP").Comment("访问时间").End().
		Index("request_id").End().
		Index("user_id").End().
		Index("ip_address").End().
		Index("method").End().
		Index("status_code").End().
		Index("response_time").End().
		Index("service").End().
		Index("accessed_at").End().
		Engine("InnoDB").
		Comment("访问日志表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 4. 创建性能监控表
	err = advancedBuilder.Table("performance_metrics").
		ID().
		String("metric_name", 100).NotNull().Comment("指标名称").End().
		String("metric_type", 50).NotNull().Comment("指标类型").End().
		Decimal("value", 15, 6).NotNull().Comment("指标值").End().
		String("unit", 20).Nullable().Comment("单位").End().
		Json("dimensions").Nullable().Comment("维度标签").End().
		String("service", 100).Nullable().Comment("服务名称").End().
		String("environment", 50).Default("production").Comment("环境").End().
		String("hostname", 255).Nullable().Comment("主机名").End().
		String("instance", 100).Nullable().Comment("实例").End().
		Timestamp("measured_at").Default("CURRENT_TIMESTAMP").Comment("测量时间").End().
		Index("metric_name").End().
		Index("metric_type").End().
		Index("service").End().
		Index("environment").End().
		Index("measured_at").End().
		Engine("InnoDB").
		Comment("性能监控表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 5. 创建系统事件表
	err = advancedBuilder.Table("system_events").
		ID().
		String("event_id", 64).NotNull().Comment("事件ID").End().
		String("event_type", 100).NotNull().Comment("事件类型").End().
		String("event_name", 200).NotNull().Comment("事件名称").End().
		String("source", 100).NotNull().Comment("事件源").End().
		Enum("severity", []string{"low", "medium", "high", "critical"}).Default("medium").Comment("严重程度").End().
		String("title", 500).NotNull().Comment("事件标题").End().
		Text("description").Nullable().Comment("事件描述").End().
		Json("payload").Nullable().Comment("事件数据").End().
		String("user_id", 36).Nullable().Comment("相关用户ID").End().
		String("resource_type", 100).Nullable().Comment("资源类型").End().
		String("resource_id", 100).Nullable().Comment("资源ID").End().
		String("service", 100).Nullable().Comment("服务名称").End().
		String("environment", 50).Default("production").Comment("环境").End().
		String("correlation_id", 64).Nullable().Comment("关联ID").End().
		Json("tags").Nullable().Comment("标签").End().
		Enum("status", []string{"open", "acknowledged", "resolved", "closed"}).Default("open").Comment("状态").End().
		String("assigned_to", 36).Nullable().Comment("分配给").End().
		Timestamp("occurred_at").Default("CURRENT_TIMESTAMP").Comment("发生时间").End().
		Timestamp("acknowledged_at").Nullable().Comment("确认时间").End().
		Timestamp("resolved_at").Nullable().Comment("解决时间").End().
		Index("event_id").End().
		Index("event_type").End().
		Index("source").End().
		Index("severity").End().
		Index("user_id").End().
		Index("resource_type").End().
		Index("resource_id").End().
		Index("service").End().
		Index("status").End().
		Index("occurred_at").End().
		Engine("InnoDB").
		Comment("系统事件表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 6. 创建日志配置表
	err = advancedBuilder.Table("log_configurations").
		ID().
		String("service", 100).NotNull().Comment("服务名称").End().
		String("environment", 50).NotNull().Comment("环境").End().
		String("logger", 100).NotNull().Comment("日志记录器").End().
		Enum("level", []string{"debug", "info", "warning", "error", "critical"}).NotNull().Comment("日志级别").End().
		Boolean("enabled").Default(true).Comment("是否启用").End().
		Integer("retention_days").Default(30).Comment("保留天数").End().
		Boolean("sampling_enabled").Default(false).Comment("是否启用采样").End().
		Decimal("sampling_rate", 5, 4).Default(1.0000).Comment("采样率").End().
		Json("filters").Nullable().Comment("过滤器").End().
		Json("metadata").Nullable().Comment("扩展配置").End().
		String("created_by", 36).Nullable().Comment("创建人").End().
		String("updated_by", 36).Nullable().Comment("更新人").End().
		Timestamps().
		Index("service").End().
		Index("environment").End().
		Index("logger").End().
		Index("enabled").End().
		Unique("service", "environment", "logger").End().
		Engine("InnoDB").
		Comment("日志配置表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 7. 创建日志归档表
	err = advancedBuilder.Table("log_archives").
		ID().
		String("archive_name", 200).NotNull().Comment("归档名称").End().
		String("table_name", 100).NotNull().Comment("源表名").End().
		Date("start_date").NotNull().Comment("开始日期").End().
		Date("end_date").NotNull().Comment("结束日期").End().
		BigInteger("record_count").Default(0).Comment("记录数量").End().
		BigInteger("file_size").Default(0).Comment("文件大小(字节)").End().
		String("storage_path", 500).Nullable().Comment("存储路径").End().
		String("compression", 50).Default("gzip").Comment("压缩方式").End().
		String("checksum", 64).Nullable().Comment("校验和").End().
		Enum("status", []string{"pending", "processing", "completed", "failed"}).Default("pending").Comment("状态").End().
		Text("error_message").Nullable().Comment("错误信息").End().
		String("created_by", 36).Nullable().Comment("创建人").End().
		Timestamp("started_at").Nullable().Comment("开始时间").End().
		Timestamp("completed_at").Nullable().Comment("完成时间").End().
		Timestamps().
		Index("table_name").End().
		Index("start_date").End().
		Index("end_date").End().
		Index("status").End().
		Engine("InnoDB").
		Comment("日志归档表").
		Create(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (m *CreateLoggingSystemMigration) Down(ctx context.Context, db types.DB) error {
	tables := []string{
		"log_archives",
		"log_configurations",
		"system_events",
		"performance_metrics",
		"access_logs",
		"error_logs",
		"application_logs",
	}

	for _, table := range tables {
		_, err := db.Exec("DROP TABLE IF EXISTS " + table)
		if err != nil {
			return err
		}
	}

	return nil
}
