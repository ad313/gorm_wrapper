package orm

import (
	"context"
	"errors"
	"fmt"
	"github.com/ad313/gorm_wrapper/orm/ref"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// OrmWrapper gorm包装器
type OrmWrapper[T any] struct {
	Error error //过程中产生的错误
	//Model       *T            //初始化一个表实例，用这个模型解析字段
	table     schema.Tabler //通过它获取表名称
	tableInfo *TableInfo[T] //表信息
	builder   *ormWrapperBuilder[T]
}

// 数据库实例
var _db *gorm.DB = nil

// 数据库类型
var _dbType string = ""

// Init 初始化塞入db，可以指定数据库类型，不指定时从db解析
func Init(db *gorm.DB, dbType ...string) {
	_db = db
	_dbType = FirstOrDefault(dbType)

	if _dbType == "" {
		dialect := db.Dialector.Name()
		switch dialect {
		case "mysql":
			_dbType = MySql
			break
		case "postgres":
			_dbType = Postgres
			break
		case "sqlite":
			_dbType = Sqlite
			break
		case "sqlserver":
			_dbType = Sqlserver
			break
		case "dm":
			_dbType = Dm
			break
		default:
			fmt.Printf("未知的数据库类型： %s", dialect)
			_dbType = dialect
			break
		}
	}
}

// BuildOrmWrapper 创建gorm包装器
func BuildOrmWrapper[T any](ctx context.Context, db ...*gorm.DB) *OrmWrapper[T] {
	var wrapper = &OrmWrapper[T]{}

	//创建模型
	var ormTableResult = BuildOrmTable[T]()
	wrapper.tableInfo = ormTableResult.Table
	wrapper.Error = ormTableResult.Error
	wrapper.builder = &ormWrapperBuilder[T]{
		wrapper:    wrapper,
		where:      make([][]any, 0),
		joinModels: make([]*joinModel, 0),
		//selectColumns:  make([]string, 0),
		groupByColumns: make([]string, 0),
		orderByColumns: make([]string, 0)}

	wrapper.SetDb(ctx, db...)

	if wrapper.Error == nil {
		model, ok := ref.IsTypeByValue[schema.Tabler](*(wrapper.tableInfo.T))
		if ok {
			wrapper.builder.TableName = (*model).TableName()
			wrapper.table = *model
		} else {
			wrapper.Error = errors.New("传入类型必须是实现了 TableName 的表实体")
		}
	}

	return wrapper
}

// SetDb 外部传入db，适用于外部开事务的场景
func (o *OrmWrapper[T]) SetDb(ctx context.Context, db ...*gorm.DB) *OrmWrapper[T] {
	if len(db) > 0 && db[0] != nil {
		o.builder.DbContext = db[0].WithContext(ctx)
		o.builder.isOuterDb = true
	} else {
		o.builder.DbContext = _db.WithContext(ctx)
		o.builder.isOuterDb = false
	}

	o.builder.ctx = ctx
	o.builder._DbContext = o.builder.DbContext

	if o.builder.DbContext == nil {
		o.Error = errors.New("请先初始化db，调用 Init 方法")
	}

	return o
}

// GetNewDb 获取新的db
func (o *OrmWrapper[T]) GetNewDb() *gorm.DB {
	return o.builder._DbContext
}

// SetTable 设置查询一个衍生表
func (o *OrmWrapper[T]) SetTable(alias string, childTable ...*gorm.DB) *OrmWrapper[T] {
	if alias == "" && len(childTable) > 0 {
		o.Error = errors.New("alias 表别名不能为空")
		return o
	}

	if alias != "" {
		o.builder.TableAlias = alias
	}

	if len(childTable) > 0 {
		o.builder.childTable = childTable[0]
	}

	return o
}

// GroupBy 可多次调用，按照调用顺序排列字段
func (o *OrmWrapper[T]) GroupBy(column any, tableAlias ...string) *OrmWrapper[T] {
	if o.builder.groupByColumns == nil {
		o.builder.groupByColumns = make([]string, 0)
	}

	name, err := mergeTableColumnWithFunc(column, FirstOrDefault(tableAlias), "", _dbType)
	if err != nil || name == "" {
		o.Error = errors.New("未获取到 GroupBy 字段名称")
		return o
	}

	o.builder.groupByColumns = append(o.builder.groupByColumns, name)

	return o
}

func (o *OrmWrapper[T]) Having(having OrmCondition) *OrmWrapper[T] {
	o.builder.addHavingWithWhereCondition(having)
	return o
}

// OrderBy 可多次调用，按照调用顺序排列字段
func (o *OrmWrapper[T]) OrderBy(column any, tableAlias ...string) *OrmWrapper[T] {
	if o.builder.orderByColumns == nil {
		o.builder.orderByColumns = make([]string, 0)
	}

	name, err := mergeTableColumnWithFunc(column, FirstOrDefault(tableAlias), "", _dbType)
	if err != nil || name == "" {
		o.Error = errors.New("未获取到 OrderBy 字段名称")
		return o
	}

	o.builder.orderByColumns = append(o.builder.orderByColumns, name)

	return o
}

// OrderByDesc 可多次调用，按照调用顺序排列字段
func (o *OrmWrapper[T]) OrderByDesc(column any, tableAlias ...string) *OrmWrapper[T] {
	if o.builder.orderByColumns == nil {
		o.builder.orderByColumns = make([]string, 0)
	}

	name, err := mergeTableColumnWithFunc(column, FirstOrDefault(tableAlias), "", _dbType)
	if err != nil || name == "" {
		o.Error = errors.New("未获取到 OrderByDesc 字段名称")
		return o
	}

	o.builder.orderByColumns = append(o.builder.orderByColumns, name+" desc")

	return o
}

// Limit Limit
func (o *OrmWrapper[T]) Limit(limit int) *OrmWrapper[T] {
	if limit > 0 {
		o.builder.DbContext = o.builder.DbContext.Limit(limit)
	}
	return o
}

// Offset Offset
func (o *OrmWrapper[T]) Offset(limit int) *OrmWrapper[T] {
	if limit > 0 {
		o.builder.DbContext = o.builder.DbContext.Offset(limit)
	}
	return o
}

// Debug 打印sql
func (o *OrmWrapper[T]) Debug() *OrmWrapper[T] {
	o.builder.DbContext = o.builder.DbContext.Debug()
	return o
}

// Unscoped 和gorm一样，忽略软删除字段
func (o *OrmWrapper[T]) Unscoped() *OrmWrapper[T] {
	o.builder.isUnscoped = true
	o.builder.DbContext = o.builder.DbContext.Unscoped()
	return o
}

// Build 创建 gorm sql
func (o *OrmWrapper[T]) Build() *gorm.DB {
	if o.Error != nil {
		return nil
	}

	return o.builder.Build()
}

// BuildForQuery 创建成一个子表，用于其他地方子查询
func (o *OrmWrapper[T]) BuildForQuery() *gorm.DB {
	if o.Error != nil {
		return nil
	}

	return o.builder.BuildForQuery()
}

// ToSql 转换成 Sql
func (o *OrmWrapper[T]) ToSql() (string, error) {
	var db = o.BuildForQuery()
	if o.Error != nil {
		return "", o.Error
	}

	return db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Find(&[]*T{})
	}), nil
}

// ToFirstOrDefaultSql 转换成 Sql
func (o *OrmWrapper[T]) ToFirstOrDefaultSql() (string, error) {
	var db = o.BuildForQuery()
	if o.Error != nil {
		return "", o.Error
	}

	return db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.First(new(T))
	}), nil
}
