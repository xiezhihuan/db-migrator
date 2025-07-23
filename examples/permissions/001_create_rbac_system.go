package permissions

import (
	"context"
	"db-migrator/internal/builder"
	"db-migrator/internal/checker"
	"db-migrator/internal/types"
)

// CreateRBACSystemMigration 创建RBAC权限系统迁移
type CreateRBACSystemMigration struct{}

func (m *CreateRBACSystemMigration) Version() string {
	return "001"
}

func (m *CreateRBACSystemMigration) Description() string {
	return "创建基于角色的权限控制系统(RBAC) - 用户、角色、权限、组织"
}

func (m *CreateRBACSystemMigration) Up(ctx context.Context, db types.DB) error {
	checker := checker.NewMySQLChecker(db, "rbac_db")
	advancedBuilder := builder.NewAdvancedBuilder(checker, db)

	// 1. 创建组织表
	err := advancedBuilder.Table("organizations").
		ID().
		String("name", 200).NotNull().Comment("组织名称").End().
		String("code", 50).NotNull().Unique().Comment("组织代码").End().
		Text("description").Nullable().Comment("组织描述").End().
		Integer("parent_id").Nullable().Comment("父组织ID").End().
		String("type", 50).Default("department").Comment("组织类型").End().
		Integer("level").Default(1).Comment("组织层级").End().
		String("path", 500).Nullable().Comment("组织路径").End().
		Integer("sort_order").Default(0).Comment("排序").End().
		String("manager_id", 36).Nullable().Comment("管理员ID").End().
		String("contact_email", 255).Nullable().Comment("联系邮箱").End().
		String("contact_phone", 20).Nullable().Comment("联系电话").End().
		Boolean("is_active").Default(true).Comment("是否激活").End().
		Json("metadata").Nullable().Comment("扩展数据").End().
		Timestamps().
		ForeignKey("parent_id").References("organizations", "id").OnDelete(builder.ActionSetNull).End().
		Index("parent_id").End().
		Index("code").End().
		Index("level").End().
		Index("is_active").End().
		Engine("InnoDB").
		Comment("组织表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 2. 创建用户表
	err = advancedBuilder.Table("users").
		String("id", 36).NotNull().Comment("用户ID(UUID)").End().
		String("username", 100).NotNull().Unique().Comment("用户名").End().
		String("email", 255).NotNull().Unique().Comment("邮箱").End().
		String("password_hash", 255).NotNull().Comment("密码哈希").End().
		String("first_name", 100).Nullable().Comment("名").End().
		String("last_name", 100).Nullable().Comment("姓").End().
		String("display_name", 200).Nullable().Comment("显示名称").End().
		String("avatar", 500).Nullable().Comment("头像URL").End().
		String("phone", 20).Nullable().Comment("电话").End().
		String("employee_id", 50).Nullable().Comment("员工号").End().
		String("job_title", 200).Nullable().Comment("职位").End().
		Integer("organization_id").Nullable().Comment("所属组织ID").End().
		String("manager_id", 36).Nullable().Comment("直属经理ID").End().
		Date("hire_date").Nullable().Comment("入职日期").End().
		Enum("status", []string{"active", "inactive", "suspended", "pending"}).Default("pending").Comment("状态").End().
		Boolean("is_system").Default(false).Comment("是否系统用户").End().
		Timestamp("email_verified_at").Nullable().Comment("邮箱验证时间").End().
		Timestamp("last_login_at").Nullable().Comment("最后登录时间").End().
		String("last_login_ip", 45).Nullable().Comment("最后登录IP").End().
		Integer("login_attempts").Default(0).Comment("登录尝试次数").End().
		Timestamp("locked_until").Nullable().Comment("锁定到期时间").End().
		Json("preferences").Nullable().Comment("用户偏好设置").End().
		Json("metadata").Nullable().Comment("扩展数据").End().
		Timestamps().
		ForeignKey("organization_id").References("organizations", "id").OnDelete(builder.ActionSetNull).End().
		ForeignKey("manager_id").References("users", "id").OnDelete(builder.ActionSetNull).End().
		Index("username").End().
		Index("email").End().
		Index("organization_id").End().
		Index("manager_id").End().
		Index("status").End().
		Index("employee_id").End().
		Engine("InnoDB").
		Comment("用户表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 为用户表添加主键
	err = advancedBuilder.ModifyTable("users").
		AddIndex(ctx, "PRIMARY", []string{"id"}, false)
	if err != nil {
		return err
	}

	// 3. 创建权限表
	err = advancedBuilder.Table("permissions").
		ID().
		String("name", 100).NotNull().Unique().Comment("权限名称").End().
		String("guard_name", 50).Default("web").Comment("守卫名称").End().
		String("resource", 100).NotNull().Comment("资源").End().
		String("action", 100).NotNull().Comment("操作").End().
		String("display_name", 200).Nullable().Comment("显示名称").End().
		Text("description").Nullable().Comment("权限描述").End().
		String("category", 100).Nullable().Comment("权限分类").End().
		String("module", 100).Nullable().Comment("所属模块").End().
		Boolean("is_system").Default(false).Comment("是否系统权限").End().
		Json("conditions").Nullable().Comment("权限条件").End().
		Integer("sort_order").Default(0).Comment("排序").End().
		Timestamps().
		Index("name").End().
		Index("resource").End().
		Index("action").End().
		Index("category").End().
		Index("module").End().
		Index("is_system").End().
		Unique("name", "guard_name").End().
		Engine("InnoDB").
		Comment("权限表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 4. 创建角色表
	err = advancedBuilder.Table("roles").
		ID().
		String("name", 100).NotNull().Comment("角色名称").End().
		String("guard_name", 50).Default("web").Comment("守卫名称").End().
		String("display_name", 200).Nullable().Comment("显示名称").End().
		Text("description").Nullable().Comment("角色描述").End().
		String("type", 50).Default("custom").Comment("角色类型").End().
		Integer("level").Default(1).Comment("角色级别").End().
		Boolean("is_system").Default(false).Comment("是否系统角色").End().
		Boolean("is_default").Default(false).Comment("是否默认角色").End().
		Json("permissions_cache").Nullable().Comment("权限缓存").End().
		Json("metadata").Nullable().Comment("扩展数据").End().
		Integer("sort_order").Default(0).Comment("排序").End().
		Timestamps().
		Index("name").End().
		Index("type").End().
		Index("level").End().
		Index("is_system").End().
		Index("is_default").End().
		Unique("name", "guard_name").End().
		Engine("InnoDB").
		Comment("角色表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 5. 创建角色权限关联表
	err = advancedBuilder.Table("role_permissions").
		ID().
		Integer("role_id").NotNull().Comment("角色ID").End().
		Integer("permission_id").NotNull().Comment("权限ID").End().
		Json("conditions").Nullable().Comment("权限条件").End().
		String("granted_by", 36).Nullable().Comment("授权人").End().
		Timestamp("granted_at").Default("CURRENT_TIMESTAMP").Comment("授权时间").End().
		ForeignKey("role_id").References("roles", "id").OnDelete(builder.ActionCascade).End().
		ForeignKey("permission_id").References("permissions", "id").OnDelete(builder.ActionCascade).End().
		ForeignKey("granted_by").References("users", "id").OnDelete(builder.ActionSetNull).End().
		Index("role_id").End().
		Index("permission_id").End().
		Unique("role_id", "permission_id").End().
		Engine("InnoDB").
		Comment("角色权限关联表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 6. 创建用户角色关联表
	err = advancedBuilder.Table("user_roles").
		ID().
		String("user_id", 36).NotNull().Comment("用户ID").End().
		Integer("role_id").NotNull().Comment("角色ID").End().
		Integer("organization_id").Nullable().Comment("组织范围").End().
		String("scope", 100).Default("global").Comment("权限范围").End().
		Json("conditions").Nullable().Comment("附加条件").End().
		String("assigned_by", 36).Nullable().Comment("分配人").End().
		Timestamp("assigned_at").Default("CURRENT_TIMESTAMP").Comment("分配时间").End().
		Timestamp("expires_at").Nullable().Comment("过期时间").End().
		Boolean("is_active").Default(true).Comment("是否激活").End().
		ForeignKey("user_id").References("users", "id").OnDelete(builder.ActionCascade).End().
		ForeignKey("role_id").References("roles", "id").OnDelete(builder.ActionCascade).End().
		ForeignKey("organization_id").References("organizations", "id").OnDelete(builder.ActionCascade).End().
		ForeignKey("assigned_by").References("users", "id").OnDelete(builder.ActionSetNull).End().
		Index("user_id").End().
		Index("role_id").End().
		Index("organization_id").End().
		Index("scope").End().
		Index("is_active").End().
		Index("expires_at").End().
		Unique("user_id", "role_id", "organization_id").End().
		Engine("InnoDB").
		Comment("用户角色关联表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 7. 创建用户直接权限关联表
	err = advancedBuilder.Table("user_permissions").
		ID().
		String("user_id", 36).NotNull().Comment("用户ID").End().
		Integer("permission_id").NotNull().Comment("权限ID").End().
		Integer("organization_id").Nullable().Comment("组织范围").End().
		String("scope", 100).Default("global").Comment("权限范围").End().
		Json("conditions").Nullable().Comment("附加条件").End().
		String("granted_by", 36).Nullable().Comment("授权人").End().
		Timestamp("granted_at").Default("CURRENT_TIMESTAMP").Comment("授权时间").End().
		Timestamp("expires_at").Nullable().Comment("过期时间").End().
		Boolean("is_active").Default(true).Comment("是否激活").End().
		ForeignKey("user_id").References("users", "id").OnDelete(builder.ActionCascade).End().
		ForeignKey("permission_id").References("permissions", "id").OnDelete(builder.ActionCascade).End().
		ForeignKey("organization_id").References("organizations", "id").OnDelete(builder.ActionCascade).End().
		ForeignKey("granted_by").References("users", "id").OnDelete(builder.ActionSetNull).End().
		Index("user_id").End().
		Index("permission_id").End().
		Index("organization_id").End().
		Index("scope").End().
		Index("is_active").End().
		Index("expires_at").End().
		Unique("user_id", "permission_id", "organization_id").End().
		Engine("InnoDB").
		Comment("用户直接权限关联表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 8. 创建会话表
	err = advancedBuilder.Table("user_sessions").
		String("id", 40).NotNull().Comment("会话ID").End().
		String("user_id", 36).Nullable().Comment("用户ID").End().
		String("ip_address", 45).Nullable().Comment("IP地址").End().
		Text("user_agent").Nullable().Comment("用户代理").End().
		Text("payload").NotNull().Comment("会话数据").End().
		Integer("last_activity").NotNull().Comment("最后活动时间").End().
		Timestamps().
		ForeignKey("user_id").References("users", "id").OnDelete(builder.ActionCascade).End().
		Index("user_id").End().
		Index("last_activity").End().
		Engine("InnoDB").
		Comment("用户会话表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 9. 创建操作日志表
	err = advancedBuilder.Table("audit_logs").
		ID().
		String("user_id", 36).Nullable().Comment("操作用户ID").End().
		String("action", 100).NotNull().Comment("操作动作").End().
		String("resource", 100).NotNull().Comment("操作资源").End().
		String("resource_id", 100).Nullable().Comment("资源ID").End().
		Json("old_values").Nullable().Comment("旧值").End().
		Json("new_values").Nullable().Comment("新值").End().
		String("ip_address", 45).Nullable().Comment("IP地址").End().
		Text("user_agent").Nullable().Comment("用户代理").End().
		Json("metadata").Nullable().Comment("扩展数据").End().
		Timestamp("created_at").Default("CURRENT_TIMESTAMP").Comment("创建时间").End().
		ForeignKey("user_id").References("users", "id").OnDelete(builder.ActionSetNull).End().
		Index("user_id").End().
		Index("action").End().
		Index("resource").End().
		Index("resource_id").End().
		Index("created_at").End().
		Engine("InnoDB").
		Comment("操作审计日志表").
		Create(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (m *CreateRBACSystemMigration) Down(ctx context.Context, db types.DB) error {
	tables := []string{
		"audit_logs",
		"user_sessions",
		"user_permissions",
		"user_roles",
		"role_permissions",
		"roles",
		"permissions",
		"users",
		"organizations",
	}

	for _, table := range tables {
		_, err := db.Exec("DROP TABLE IF EXISTS " + table)
		if err != nil {
			return err
		}
	}

	return nil
}
