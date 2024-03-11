package gormWapper

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"strings"
)

//type OrmWrapperInterface[T interface{}] interface {
//	//Init(db *gorm.DB, dbType string)
//	Where(query interface{}, args ...interface{}) OrmWrapperInterface[T]
//}

// OrmWrapper gorm包装器
type OrmWrapper[T any] struct {
	Error   error
	Model   *T
	table   schema.Tabler
	builder *ormWrapperBuilder[T]
}

// 数据库实例
var _db *gorm.DB = nil

// 数据库类型
var _dbType string = ""

// Init 初始化塞入db
func Init(db *gorm.DB, dbType string) {
	_db = db
	_dbType = dbType
}

// BuildOrmWrapper 创建gorm包装器
func BuildOrmWrapper[T any](ctx context.Context, db ...*gorm.DB) *OrmWrapper[T] {
	var wrapper = &OrmWrapper[T]{}

	//创建模型
	var buildResult = BuildGormTable[T]()
	wrapper.Model = buildResult.Table.T
	wrapper.Error = buildResult.Error
	wrapper.builder = &ormWrapperBuilder[T]{
		wrapper:        wrapper,
		where:          make([][]any, 0),
		leftJoin:       make([]*leftJoinModel, 0),
		selectColumns:  make([]string, 0),
		groupByColumns: make([]string, 0),
		orderByColumns: make([]string, 0)}

	wrapper.SetDb(ctx, db...)

	if wrapper.Error == nil {
		model, ok := IsTypeByValue[schema.Tabler](*(wrapper.Model))
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
func (a *OrmWrapper[T]) SetDb(ctx context.Context, db ...*gorm.DB) *OrmWrapper[T] {
	if len(db) > 0 {
		a.builder.DbContext = db[0].WithContext(ctx)
		a.builder.isOuterDb = true
	} else {
		a.builder.DbContext = _db.WithContext(ctx)
		a.builder.isOuterDb = false
	}

	a.builder.ctx = ctx

	if a.builder.DbContext == nil {
		a.Error = errors.New("请先初始化db，调用 Init 方法")
	}

	return a
}

// SetTableAlias 指定主表表别名，如果不指定，当有 left join 或者 exists时，默认是表名称
func (a *OrmWrapper[T]) SetTableAlias(alias string) *OrmWrapper[T] {
	if alias != "" {
		a.builder.TableAlias = alias
	}
	return a
}

// Where gorm 原生查询
func (a *OrmWrapper[T]) Where(query interface{}, args ...interface{}) *OrmWrapper[T] {
	a.builder.addWhere(query, args)
	return a
}

// WhereIf gorm 原生查询，加入 bool 条件控制
func (a *OrmWrapper[T]) WhereIf(do bool, query interface{}, args ...interface{}) *OrmWrapper[T] {
	if do {
		return a.Where(query, args...)
	}
	return a
}

// WhereIfNotNil gorm 原生查询，值为 nil 时跳过
func (a *OrmWrapper[T]) WhereIfNotNil(query interface{}, arg interface{}) *OrmWrapper[T] {
	if arg != nil {
		return a.Where(query, arg)
	}
	return a
}

// WhereCondition 通过条件查询
func (a *OrmWrapper[T]) WhereCondition(query WhereCondition) *OrmWrapper[T] {
	a.builder.addWhereWithWhereCondition(query)
	return a
}

// WhereConditionIf 通过条件查询，加入 bool 条件控制
func (a *OrmWrapper[T]) WhereConditionIf(do bool, query WhereCondition) *OrmWrapper[T] {
	if do {
		return a.WhereCondition(query)
	}
	return a
}

// WhereByColumn 通过字段查询，连表时支持传入表别名
func (a *OrmWrapper[T]) WhereByColumn(column any, compareSymbols string, arg interface{}, tableAlias ...string) *OrmWrapper[T] {
	var cond = &Condition{
		TableAlias:     "",
		Column:         column,
		CompareSymbols: compareSymbols,
		Arg:            arg,
		Func:           "",
	}

	if len(tableAlias) > 0 {
		cond.TableAlias = tableAlias[0]
	}

	return a.WhereCondition(cond)
}

// WhereByColumnIf 通过字段查询，连表时支持传入表别名
func (a *OrmWrapper[T]) WhereByColumnIf(do bool, column any, compareSymbols string, arg interface{}, tableAlias ...string) *OrmWrapper[T] {
	if do {
		a.WhereByColumn(column, compareSymbols, arg, tableAlias...)
	}
	return a
}

// LeftJoin 左连表
func (a *OrmWrapper[T]) LeftJoin(table schema.Tabler, alias string, leftColumn any, rightColumn any) *OrmWrapper[T] {
	if a.builder.leftJoin == nil {
		a.builder.leftJoin = make([]*leftJoinModel, 0)
	}

	if table == nil || leftColumn == nil || rightColumn == nil {
		return a
	}

	left, err := resolveColumnName(leftColumn, _dbType)
	if err != nil || left == "" {
		a.Error = errors.New("LeftJoin 未获取到左边字段")
		return a
	}

	right, err := resolveColumnName(rightColumn, _dbType)
	if err != nil || right == "" {
		a.Error = errors.New("LeftJoin 未获取到右边字段")
		return a
	}

	if alias == "" {
		alias = table.TableName()
	}

	var leftTable schema.Tabler
	var leftTableName string
	if len(a.builder.leftJoin) == 0 {
		leftTable = a.table
		leftTableName = chooseTrueValue(a.builder.TableAlias != "", a.builder.TableAlias, a.builder.TableName)
	} else {
		var lastLeftJoin = a.builder.leftJoin[len(a.builder.leftJoin)-1]
		leftTableName = lastLeftJoin.Alias
		leftTable = lastLeftJoin.Table
	}

	var joinModel = &leftJoinModel{
		Table:     table,
		tableName: table.TableName(),
		Alias:     alias,
		Left:      formatSqlName(leftTableName, _dbType) + "." + left,
		Right:     formatSqlName(alias, _dbType) + "." + right,
	}

	//软删除
	if !a.builder.isUnscoped {
		leftSoftDel, err := getTableSoftDeleteColumnSql(leftTable, leftTableName, _dbType)
		if err != nil {
			a.Error = err
			return a
		}

		rightSoftDel, err := getTableSoftDeleteColumnSql(table, alias, _dbType)
		if err != nil {
			a.Error = err
			return a
		}

		if leftSoftDel != "" {
			joinModel.ext = " AND " + leftSoftDel
		}

		if rightSoftDel != "" {
			joinModel.ext += " AND " + rightSoftDel
		}
	}

	a.builder.leftJoin = append(a.builder.leftJoin, joinModel)

	return a
}

// LeftJoinIf 左连表
func (a *OrmWrapper[T]) LeftJoinIf(do bool, table schema.Tabler, alias string, leftColumn any, rightColumn any) *OrmWrapper[T] {
	if do {
		return a.LeftJoin(table, alias, leftColumn, rightColumn)
	}

	return a
}

// Select 查询主表字段
func (a *OrmWrapper[T]) Select(selectColumns ...interface{}) *OrmWrapper[T] {
	if selectColumns == nil || len(selectColumns) == 0 {
		return a
	}

	//判断是否有 leftJoin，如果有则给字段名加上主表别名
	var table = chooseTrueValue(len(a.builder.leftJoin) == 0, "", a.builder.TableAlias)

	return a.SelectWithTableAlias(table, selectColumns...)
}

// SelectWithTableAlias 传入表别名，查询此表下的多个字段
func (a *OrmWrapper[T]) SelectWithTableAlias(tableAlias string, selectColumns ...interface{}) *OrmWrapper[T] {
	if selectColumns == nil || len(selectColumns) == 0 {
		return a
	}

	for _, column := range selectColumns {
		name, err := resolveColumnName(column, _dbType)
		if err != nil || name == "" {
			a.Error = errors.New("未获取到字段名称")
			continue
		}

		a.builder.selectColumns = append(a.builder.selectColumns, a.builder.mergeColumnName(name, "", tableAlias))
	}

	return a
}

// SelectColumn 单次查询一个字段，可传入 字段别名，表名；可多次调用
func (a *OrmWrapper[T]) SelectColumn(selectColumn interface{}, columnAlias string, tableAlias ...string) *OrmWrapper[T] {
	name, err := resolveColumnName(selectColumn, _dbType)
	if err != nil || name == "" {
		a.Error = errors.New("未获取到字段名称")
		return a
	}

	a.builder.selectColumns = append(a.builder.selectColumns, a.builder.mergeColumnName(name, columnAlias, FirstOrDefault(tableAlias)))

	return a
}

// SelectColumnOriginal 单次查询一个字段，可传入 字段别名，表名；可多次调用；不处理字段名
func (a *OrmWrapper[T]) SelectColumnOriginal(selectColumn string, columnAlias string, tableAlias ...string) *OrmWrapper[T] {
	if selectColumn == "" {
		a.Error = errors.New("未获取到字段名称")
		return a
	}

	a.builder.selectColumns = append(a.builder.selectColumns, a.builder.mergeColumnName(selectColumn, columnAlias, FirstOrDefault(tableAlias)))

	return a
}

// SelectWithFunc 传入表别名，查询此表下的字段
func (a *OrmWrapper[T]) SelectWithFunc(selectColumn string, columnAlias string, f string, tableAlias ...string) *OrmWrapper[T] {
	name, err := resolveColumnName(selectColumn, _dbType)
	if err != nil || name == "" {
		a.Error = errors.New("未获取到字段名称")
		return a
	}

	var table = FirstOrDefault(tableAlias)
	a.builder.selectColumns = append(a.builder.selectColumns, a.builder.mergeColumnNameWithFunc(name, columnAlias, table, f))

	return a
}

// GroupBy 可多次调用，按照调用顺序排列字段
func (a *OrmWrapper[T]) GroupBy(column any, tableAlias ...string) *OrmWrapper[T] {
	if a.builder.groupByColumns == nil {
		a.builder.groupByColumns = make([]string, 0)
	}

	name, err := mergeTableColumnWithFunc(column, FirstOrDefault(tableAlias), "", _dbType)
	if err != nil || name == "" {
		a.Error = errors.New("未获取到 GroupBy 字段名称")
		return a
	}

	a.builder.groupByColumns = append(a.builder.groupByColumns, name)

	return a
}

// OrderBy 可多次调用，按照调用顺序排列字段
func (a *OrmWrapper[T]) OrderBy(column any, tableAlias ...string) *OrmWrapper[T] {
	if a.builder.orderByColumns == nil {
		a.builder.orderByColumns = make([]string, 0)
	}

	name, err := mergeTableColumnWithFunc(column, FirstOrDefault(tableAlias), "", _dbType)
	if err != nil || name == "" {
		a.Error = errors.New("未获取到 OrderBy 字段名称")
		return a
	}

	a.builder.orderByColumns = append(a.builder.orderByColumns, name)

	return a
}

// OrderByDesc 可多次调用，按照调用顺序排列字段
func (a *OrmWrapper[T]) OrderByDesc(column any, tableAlias ...string) *OrmWrapper[T] {
	if a.builder.orderByColumns == nil {
		a.builder.orderByColumns = make([]string, 0)
	}

	name, err := mergeTableColumnWithFunc(column, FirstOrDefault(tableAlias), "", _dbType)
	if err != nil || name == "" {
		a.Error = errors.New("未获取到 OrderByDesc 字段名称")
		return a
	}

	a.builder.orderByColumns = append(a.builder.orderByColumns, name+" desc")

	return a
}

// Unscoped 和gorm一样，忽略软删除字段
func (a *OrmWrapper[T]) Unscoped() *OrmWrapper[T] {
	a.builder.isUnscoped = true
	a.builder.DbContext = a.builder.DbContext.Unscoped()
	return a
}

// ToSql 转换成 Sql
func (a *OrmWrapper[T]) ToSql() (string, error) {
	var db = a.BuildForQuery()
	if a.Error != nil {
		return "", a.Error
	}

	return db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Find(&[]*T{})
	}), nil
}

// Count 查询总条数
func (a *OrmWrapper[T]) Count() (int64, error) {

	//Build sql
	a.BuildForQuery()

	//创建语句过程中的错误
	if a.Error != nil {
		return 0, a.Error
	}

	var err error
	var total int64

	//left join 加上 distinct
	if len(a.builder.leftJoin) > 0 {
		err = _db.Table("(?) as leftJoinTableWrapper", a.builder.DbContext).Count(&total).Error
	} else {
		err = a.builder.DbContext.Count(&total).Error
	}
	if err != nil {
		return 0, err
	}

	return total, nil
}

// FirstOrDefault 返回第一条，没命中返回nil，可以传入自定义scan，自定义接收数据
func (a *OrmWrapper[T]) FirstOrDefault(scan ...func(db *gorm.DB) error) (*T, error) {

	//Build sql
	a.BuildForQuery()

	//创建语句过程中的错误
	if a.Error != nil {
		return nil, a.Error
	}

	var err error
	var result = new(T)
	if len(scan) > 0 {
		if scan[0] == nil {
			return nil, errors.New("scan 函数不能为空")
		}
		err = scan[0](a.builder.DbContext)
	} else {
		//First 会自动添加主键排序
		err = a.builder.DbContext.Take(result).Error
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return result, nil
}

// ToList 返回列表，可以传入自定义scan，自定义接收数据
func (a *OrmWrapper[T]) ToList(scan ...func(db *gorm.DB) error) ([]*T, error) {

	//创建语句过程中的错误
	if a.Error != nil {
		return nil, a.Error
	}

	//Build sql
	a.BuildForQuery()

	if len(scan) > 0 {
		if scan[0] == nil {
			return nil, errors.New("scan 函数不能为空")
		}
		return nil, scan[0](a.builder.DbContext)
	}

	var list = make([]*T, 0)
	err := a.builder.DbContext.Scan(&list).Error
	if err != nil {
		return nil, err
	}

	return list, nil
}

// ToPagerList 分页查询，可以自定义scan，否则返回当前实体分页结果
func (a *OrmWrapper[T]) ToPagerList(pager *Pager, scan ...func(db *gorm.DB) error) (*PagerList[T], error) {

	//Build sql
	a.BuildForQuery()

	//创建语句过程中的错误
	if a.Error != nil {
		return nil, a.Error
	}

	if pager == nil {
		return nil, errors.New("传入分页数据不能为空")
	}

	//包含空格 asc desc
	if strings.Contains(pager.Order, " ") {
		var arr = strings.Split(pager.Order, " ")
		if strings.ToUpper(arr[1]) == "DESC" {
			a.OrderByDesc(arr[0])
		} else {
			a.OrderBy(arr[0])
		}
	} else if pager.Order != "" {
		a.OrderBy(pager.Order)
	}

	if pager.Page <= 0 {
		pager.Page = 1
	}

	if pager.PageSize <= 0 {
		pager.PageSize = 20
	}

	//总条数
	var total int64
	var err error

	//left join 加上 distinct
	if len(a.builder.leftJoin) > 0 {
		err = _db.Table("(?) as leftJoinTableWrapper", a.builder.DbContext).Count(&total).Error
	} else {
		err = a.builder.DbContext.Count(&total).Error
	}
	if err != nil {
		return nil, err
	}

	var result = &PagerList[T]{
		Page:       pager.Page,
		PageSize:   pager.PageSize,
		TotalCount: int32(total),
		Order:      pager.Order,
	}

	if result.TotalCount == 0 {
		return result, nil
	}

	if len(scan) == 0 {
		err = a.builder.DbContext.Offset(int((pager.Page - 1) * pager.PageSize)).Limit(int(pager.PageSize)).Scan(&result.Data).Error
	} else {
		if scan[0] == nil {
			return nil, errors.New("scan 函数不能为空")
		}

		err = scan[0](a.builder.DbContext.Offset(int((pager.Page - 1) * pager.PageSize)).Limit(int(pager.PageSize)))
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

// Update 更新，传了字段只更新出入字段，否则更新全部
func (a *OrmWrapper[T]) Update(item *T, updateColumns ...interface{}) (int64, error) {
	if item == nil {
		return 0, nil
	}

	var isUpdateAll = false
	if len(updateColumns) > 0 {
		a.Select(updateColumns...)
	} else {
		isUpdateAll = true
	}

	a.BuildForQuery()

	//创建语句过程中的错误
	if a.Error != nil {
		return 0, a.Error
	}

	var result *gorm.DB
	if isUpdateAll {
		result = a.builder.DbContext.Save(item)
		return result.RowsAffected, result.Error
	} else {
		result = a.builder.DbContext.UpdateColumns(item)
		return result.RowsAffected, result.Error
	}
}

// UpdateList 更新，传了字段只更新出入字段，否则更新全部
func (a *OrmWrapper[T]) UpdateList(items []*T, updateColumns ...interface{}) (int64, error) {
	if len(items) == 0 {
		return 0, nil
	}

	var total int64 = 0

	//外部开启了事务
	if a.builder.isOuterDb {
		for _, item := range items {
			c, err := a.Update(item, updateColumns...)
			if err != nil {
				return 0, err
			}

			total += c
		}

		return total, nil
	}

	//本地开事务
	var db = a.builder.DbContext

	err := db.Transaction(func(tx *gorm.DB) error {
		for i, item := range items {
			//重新设置db
			if i == 0 {
				a.SetDb(a.builder.ctx, tx)
			}

			c, err := a.Update(item, updateColumns...)
			if err != nil {
				return err
			}

			total += c
		}

		a.SetDb(a.builder.ctx, db)

		return nil
	})

	if err != nil {
		return 0, err
	}

	return total, nil
}

// Build 创建 gorm sql
func (a *OrmWrapper[T]) Build() *gorm.DB {
	if a.Error != nil {
		return nil
	}

	return a.builder.Build()
}

// BuildForQuery 创建 gorm sql
func (a *OrmWrapper[T]) BuildForQuery() *gorm.DB {
	if a.Error != nil {
		return nil
	}

	return a.builder.BuildForQuery()
}
