package orm

import (
	"context"
	"errors"
	"github.com/ad313/gorm_wrapper/orm/ref"
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

// Init 初始化塞入db
func Init(db *gorm.DB, dbType string) {
	_db = db
	_dbType = dbType
}

// BuildOrmWrapper 创建gorm包装器
func BuildOrmWrapper[T any](ctx context.Context, db ...*gorm.DB) *OrmWrapper[T] {
	var wrapper = &OrmWrapper[T]{}

	//创建模型
	var ormTableResult = BuildOrmTable[T]()
	wrapper.tableInfo = ormTableResult.Table
	//wrapper.Model = ormTableResult.Table.T
	wrapper.Error = ormTableResult.Error
	wrapper.builder = &ormWrapperBuilder[T]{
		wrapper:        wrapper,
		where:          make([][]any, 0),
		joinModels:     make([]*joinModel, 0),
		selectColumns:  make([]string, 0),
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
	if len(db) > 0 {
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

// SetTableAlias 指定主表表别名，如果不指定，当有 left join 或者 exists时，默认是表名称
func (o *OrmWrapper[T]) SetTableAlias(alias string) *OrmWrapper[T] {
	if alias != "" {
		o.builder.TableAlias = alias
	}
	return o
}

// Where gorm 原生查询
func (o *OrmWrapper[T]) Where(query interface{}, args ...interface{}) *OrmWrapper[T] {
	o.builder.addWhere(query, args)
	return o
}

// WhereIf gorm 原生查询，加入 bool 条件控制
func (o *OrmWrapper[T]) WhereIf(do bool, query interface{}, args ...interface{}) *OrmWrapper[T] {
	if do {
		return o.Where(query, args...)
	}
	return o
}

// WhereIfNotNil gorm 原生查询，值为 nil 时跳过
func (o *OrmWrapper[T]) WhereIfNotNil(query interface{}, arg interface{}) *OrmWrapper[T] {
	if arg != nil {
		return o.Where(query, arg)
	}
	return o
}

// WhereCondition 通过条件查询
func (o *OrmWrapper[T]) WhereCondition(query WhereCondition) *OrmWrapper[T] {
	o.builder.addWhereWithWhereCondition(query)
	return o
}

// WhereConditionIf 通过条件查询，加入 bool 条件控制
func (o *OrmWrapper[T]) WhereConditionIf(do bool, query WhereCondition) *OrmWrapper[T] {
	if do {
		return o.WhereCondition(query)
	}
	return o
}

// WhereByColumn 通过字段查询，连表时支持传入表别名
func (o *OrmWrapper[T]) WhereByColumn(column any, compareSymbols string, arg interface{}, tableAlias ...string) *OrmWrapper[T] {
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

	return o.WhereCondition(cond)
}

// WhereByColumnIf 通过字段查询，连表时支持传入表别名
func (o *OrmWrapper[T]) WhereByColumnIf(do bool, column any, compareSymbols string, arg interface{}, tableAlias ...string) *OrmWrapper[T] {
	if do {
		o.WhereByColumn(column, compareSymbols, arg, tableAlias...)
	}
	return o
}

// LeftJoin 左连表
func (o *OrmWrapper[T]) LeftJoin(table schema.Tabler, alias string, leftColumn any, rightColumn any) *OrmWrapper[T] {
	return o.Join(table, alias, leftColumn, rightColumn, "Left Join")
}

// LeftJoinIf 左连表
func (o *OrmWrapper[T]) LeftJoinIf(do bool, table schema.Tabler, alias string, leftColumn any, rightColumn any) *OrmWrapper[T] {
	if do {
		return o.LeftJoin(table, alias, leftColumn, rightColumn)
	}

	return o
}

// InnerJoin 内连表
func (o *OrmWrapper[T]) InnerJoin(table schema.Tabler, alias string, leftColumn any, rightColumn any) *OrmWrapper[T] {
	return o.Join(table, alias, leftColumn, rightColumn, "Inner Join")
}

// InnerJoinIf 内连表
func (o *OrmWrapper[T]) InnerJoinIf(do bool, table schema.Tabler, alias string, leftColumn any, rightColumn any) *OrmWrapper[T] {
	if do {
		return o.InnerJoin(table, alias, leftColumn, rightColumn)
	}

	return o
}

// Join 内连表
func (o *OrmWrapper[T]) Join(table schema.Tabler, alias string, leftColumn any, rightColumn any, key string) *OrmWrapper[T] {
	if o.builder.joinModels == nil {
		o.builder.joinModels = make([]*joinModel, 0)
	}

	if key == "" {
		o.Error = errors.New("join 连表参数不正确")
		return o
	}

	if table == nil || leftColumn == nil || rightColumn == nil {
		o.Error = errors.New("join 连表参数不正确")
		return o
	}

	left, err := resolveColumnName(leftColumn, _dbType)
	if err != nil || left == "" {
		o.Error = errors.New("LeftJoin 未获取到左边字段")
		return o
	}

	right, err := resolveColumnName(rightColumn, _dbType)
	if err != nil || right == "" {
		o.Error = errors.New("LeftJoin 未获取到右边字段")
		return o
	}

	if alias == "" {
		alias = table.TableName()
	}

	var leftTable schema.Tabler
	var leftTableName string
	if len(o.builder.joinModels) == 0 {
		leftTable = o.table
		leftTableName = chooseTrueValue(o.builder.TableAlias != "", o.builder.TableAlias, o.builder.TableName)
	} else {
		var lastLeftJoin = o.builder.joinModels[len(o.builder.joinModels)-1]
		leftTableName = lastLeftJoin.Alias
		leftTable = lastLeftJoin.Table
	}

	var join = &joinModel{
		Table:     table,
		tableName: table.TableName(),
		Alias:     alias,
		Left:      formatSqlName(leftTableName, _dbType) + "." + left,
		Right:     formatSqlName(alias, _dbType) + "." + right,
		key:       key,
	}

	//软删除
	if !o.builder.isUnscoped {
		leftSoftDel, err := getTableSoftDeleteColumnSql(leftTable, leftTableName, _dbType)
		if err != nil {
			o.Error = err
			return o
		}

		rightSoftDel, err := getTableSoftDeleteColumnSql(table, alias, _dbType)
		if err != nil {
			o.Error = err
			return o
		}

		if leftSoftDel != "" {
			join.ext = " AND " + leftSoftDel
		}

		if rightSoftDel != "" {
			join.ext += " AND " + rightSoftDel
		}
	}

	o.builder.joinModels = append(o.builder.joinModels, join)

	return o
}

//// Join2 内连表
//func (o *OrmWrapper[T]) Join2(table *gorm.DB, alias string, leftColumn any, rightColumn any, key string) *OrmWrapper[T] {
//	if o.builder.joinModels == nil {
//		o.builder.joinModels = make([]*joinModel, 0)
//	}
//
//	if key == "" {
//		o.Error = errors.New("join 连表参数不正确")
//		return o
//	}
//
//	if table == nil || leftColumn == nil || rightColumn == nil {
//		o.Error = errors.New("join 连表参数不正确")
//		return o
//	}
//
//	left, err := resolveColumnName(leftColumn, _dbType)
//	if err != nil || left == "" {
//		o.Error = errors.New("LeftJoin 未获取到左边字段")
//		return o
//	}
//
//	right, err := resolveColumnName(rightColumn, _dbType)
//	if err != nil || right == "" {
//		o.Error = errors.New("LeftJoin 未获取到右边字段")
//		return o
//	}
//
//	if alias == "" {
//		alias = table.TableName()
//	}
//
//	var leftTable schema.Tabler
//	var leftTableName string
//	if len(o.builder.joinModels) == 0 {
//		leftTable = o.table
//		leftTableName = chooseTrueValue(o.builder.TableAlias != "", o.builder.TableAlias, o.builder.TableName)
//	} else {
//		var lastLeftJoin = o.builder.joinModels[len(o.builder.joinModels)-1]
//		leftTableName = lastLeftJoin.Alias
//		leftTable = lastLeftJoin.Table
//	}
//
//	var join = &joinModel{
//		Table:     table,
//		tableName: table.TableName(),
//		Alias:     alias,
//		Left:      formatSqlName(leftTableName, _dbType) + "." + left,
//		Right:     formatSqlName(alias, _dbType) + "." + right,
//		key:       key,
//	}
//
//	//软删除
//	if !o.builder.isUnscoped {
//		leftSoftDel, err := getTableSoftDeleteColumnSql(leftTable, leftTableName, _dbType)
//		if err != nil {
//			o.Error = err
//			return o
//		}
//
//		rightSoftDel, err := getTableSoftDeleteColumnSql(table, alias, _dbType)
//		if err != nil {
//			o.Error = err
//			return o
//		}
//
//		if leftSoftDel != "" {
//			join.ext = " AND " + leftSoftDel
//		}
//
//		if rightSoftDel != "" {
//			join.ext += " AND " + rightSoftDel
//		}
//	}
//
//	o.builder.joinModels = append(o.builder.joinModels, join)
//
//	return o
//}

// Select 查询主表字段
func (o *OrmWrapper[T]) Select(selectColumns ...interface{}) *OrmWrapper[T] {
	if selectColumns == nil || len(selectColumns) == 0 {
		return o
	}

	//判断是否有 leftJoin，如果有则给字段名加上主表别名
	var table = chooseTrueValue(len(o.builder.joinModels) == 0, "", o.builder.TableAlias)

	return o.SelectWithTableAlias(table, selectColumns...)
}

// SelectWithTableAlias 传入表别名，查询此表下的多个字段
func (o *OrmWrapper[T]) SelectWithTableAlias(tableAlias string, selectColumns ...interface{}) *OrmWrapper[T] {
	if selectColumns == nil || len(selectColumns) == 0 {
		return o
	}

	for _, column := range selectColumns {
		name, err := resolveColumnName(column, _dbType)
		if err != nil || name == "" {
			o.Error = errors.New("未获取到字段名称")
			continue
		}

		o.builder.selectColumns = append(o.builder.selectColumns, o.builder.mergeColumnName(name, "", tableAlias))
	}

	return o
}

// SelectColumn 单次查询一个字段，可传入 字段别名，表名；可多次调用
func (o *OrmWrapper[T]) SelectColumn(selectColumn interface{}, columnAlias string, tableAlias ...string) *OrmWrapper[T] {
	name, err := resolveColumnName(selectColumn, _dbType)
	if err != nil || name == "" {
		o.Error = errors.New("未获取到字段名称")
		return o
	}

	o.builder.selectColumns = append(o.builder.selectColumns, o.builder.mergeColumnName(name, columnAlias, FirstOrDefault(tableAlias)))

	return o
}

// SelectColumnOriginal 单次查询一个字段，可传入 字段别名，表名；可多次调用；不处理字段名
func (o *OrmWrapper[T]) SelectColumnOriginal(selectColumn string, columnAlias string, tableAlias ...string) *OrmWrapper[T] {
	if selectColumn == "" {
		o.Error = errors.New("未获取到字段名称")
		return o
	}

	o.builder.selectColumns = append(o.builder.selectColumns, o.builder.mergeColumnName(selectColumn, columnAlias, FirstOrDefault(tableAlias)))

	return o
}

// SelectWithFunc 传入表别名，查询此表下的字段
func (o *OrmWrapper[T]) SelectWithFunc(selectColumn interface{}, columnAlias string, f string, tableAlias ...string) *OrmWrapper[T] {
	name, err := resolveColumnName(selectColumn, "")
	if err != nil || name == "" {
		o.Error = errors.New("未获取到字段名称")
		return o
	}

	var table = FirstOrDefault(tableAlias)
	o.builder.selectColumns = append(o.builder.selectColumns, o.builder.mergeColumnNameWithFunc(name, columnAlias, table, f))

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

func (o *OrmWrapper[T]) Having(having WhereCondition) *OrmWrapper[T] {
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

// Unscoped 和gorm一样，忽略软删除字段
func (o *OrmWrapper[T]) Unscoped() *OrmWrapper[T] {
	o.builder.isUnscoped = true
	o.builder.DbContext = o.builder.DbContext.Unscoped()
	return o
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

// Count 查询总条数
func (o *OrmWrapper[T]) Count() (int64, error) {

	//Build sql
	o.BuildForQuery()

	//创建语句过程中的错误
	if o.Error != nil {
		return 0, o.Error
	}

	var err error
	var total int64

	//left join 加上 distinct
	if len(o.builder.joinModels) > 0 {
		err = _db.Table("(?) as leftJoinTableWrapper", o.builder.DbContext).Count(&total).Error
	} else {
		err = o.builder.DbContext.Count(&total).Error
	}
	if err != nil {
		return 0, err
	}

	return total, nil
}

// FirstOrDefault 返回第一条，没命中返回nil，可以传入自定义scan，自定义接收数据
func (o *OrmWrapper[T]) FirstOrDefault(scan ...func(db *gorm.DB) error) (*T, error) {

	//Build sql
	o.BuildForQuery()

	//创建语句过程中的错误
	if o.Error != nil {
		return nil, o.Error
	}

	var err error
	var result = new(T)
	if len(scan) > 0 {
		if scan[0] == nil {
			return nil, errors.New("scan 函数不能为空")
		}
		err = scan[0](o.builder.DbContext)
	} else {
		//First 会自动添加主键排序
		err = o.builder.DbContext.Take(result).Error
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
func (o *OrmWrapper[T]) ToList(scan ...func(db *gorm.DB) error) ([]*T, error) {

	//创建语句过程中的错误
	if o.Error != nil {
		return nil, o.Error
	}

	//Build sql
	o.BuildForQuery()

	if len(scan) > 0 {
		if scan[0] == nil {
			return nil, errors.New("scan 函数不能为空")
		}
		return nil, scan[0](o.builder.DbContext)
	}

	var list = make([]*T, 0)
	err := o.builder.DbContext.Scan(&list).Error
	if err != nil {
		return nil, err
	}

	return list, nil
}

// ToPagerList 分页查询，可以自定义scan，否则返回当前实体分页结果
func (o *OrmWrapper[T]) ToPagerList(pager *Pager, scan ...func(db *gorm.DB) error) (*PagerList[T], error) {

	//Build sql
	o.BuildForQuery()

	//创建语句过程中的错误
	if o.Error != nil {
		return nil, o.Error
	}

	if pager == nil {
		return nil, errors.New("传入分页数据不能为空")
	}

	//包含空格 asc desc
	if strings.Contains(pager.Order, " ") {
		var arr = strings.Split(pager.Order, " ")
		if strings.ToUpper(arr[1]) == "DESC" {
			o.OrderByDesc(arr[0])
		} else {
			o.OrderBy(arr[0])
		}
	} else if pager.Order != "" {
		o.OrderBy(pager.Order)
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
	if len(o.builder.joinModels) > 0 {
		err = _db.Table("(?) as leftJoinTableWrapper", o.builder.DbContext).Count(&total).Error
	} else {
		err = o.builder.DbContext.Count(&total).Error
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
		err = o.builder.DbContext.Offset(int((pager.Page - 1) * pager.PageSize)).Limit(int(pager.PageSize)).Scan(&result.Data).Error
	} else {
		if scan[0] == nil {
			return nil, errors.New("scan 函数不能为空")
		}

		err = scan[0](o.builder.DbContext.Offset(int((pager.Page - 1) * pager.PageSize)).Limit(int(pager.PageSize)))
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

// Update 更新，传了字段只更新出入字段，否则更新全部
func (o *OrmWrapper[T]) Update(item *T, updateColumns ...interface{}) (int64, error) {
	if item == nil {
		return 0, nil
	}

	var isUpdateAll = false
	if len(updateColumns) > 0 {
		o.Select(updateColumns...)
	} else {
		isUpdateAll = true
	}

	o.BuildForQuery()

	//创建语句过程中的错误
	if o.Error != nil {
		return 0, o.Error
	}

	var result *gorm.DB
	if isUpdateAll {
		result = o.builder.DbContext.Save(item)
		return result.RowsAffected, result.Error
	} else {
		result = o.builder.DbContext.UpdateColumns(item)
		return result.RowsAffected, result.Error
	}
}

// UpdateList 更新，传了字段只更新出入字段，否则更新全部
func (o *OrmWrapper[T]) UpdateList(items []*T, updateColumns ...interface{}) (int64, error) {
	if len(items) == 0 {
		return 0, nil
	}

	var total int64 = 0

	//外部开启了事务
	if o.builder.isOuterDb {
		for _, item := range items {
			c, err := o.Update(item, updateColumns...)
			if err != nil {
				return 0, err
			}

			total += c
		}

		return total, nil
	}

	//本地开事务
	var db = o.builder.DbContext

	err := db.Transaction(func(tx *gorm.DB) error {
		for i, item := range items {
			//重新设置db
			if i == 0 {
				o.SetDb(o.builder.ctx, tx)
			}

			c, err := o.Update(item, updateColumns...)
			if err != nil {
				return err
			}

			total += c
		}

		o.SetDb(o.builder.ctx, db)

		return nil
	})

	if err != nil {
		return 0, err
	}

	return total, nil
}

// DeleteById 通过id删除数据，可以传入id集合
func (o *OrmWrapper[T]) DeleteById(idList ...interface{}) error {
	if len(idList) == 0 {
		o.Error = errors.New("idList 不能为空")
		return o.Error
	}

	return o.builder.DbContext.Delete(new(T), idList).Error
}

// GetById 通过id获取数据
func (o *OrmWrapper[T]) GetById(id interface{}) (*T, error) {
	if id == nil {
		o.Error = errors.New("id 不能为空")
		return nil, o.Error
	}

	key, err := o.builder.getPrimaryKey()
	if err != nil {
		return nil, err
	}

	sql, _, err := (&Condition{
		Column:         key,
		CompareSymbols: Eq,
		Arg:            id,
	}).BuildSql(_dbType)
	if err != nil {
		return nil, err
	}

	var result = new(T)
	err = o.GetNewDb().Model(new(T)).Where(sql, id).Take(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return result, err
}

// GetByIds 通过id获取数据
func (o *OrmWrapper[T]) GetByIds(idList []interface{}) ([]*T, error) {
	if len(idList) == 0 {
		o.Error = errors.New("idList 不能为空")
		return nil, o.Error
	}

	key, err := o.builder.getPrimaryKey()
	if err != nil {
		return nil, err
	}

	sql, _, err := (&Condition{
		Column:         key,
		CompareSymbols: In,
		Arg:            idList,
	}).BuildSql(_dbType)
	if err != nil {
		return nil, err
	}

	var result = make([]*T, 0)
	return result, o.GetNewDb().Model(new(T)).Where(sql, idList).Scan(&result).Error
}

// Build 创建 gorm sql
func (o *OrmWrapper[T]) Build() *gorm.DB {
	if o.Error != nil {
		return nil
	}

	return o.builder.Build()
}

// BuildForQuery 创建 gorm sql
func (o *OrmWrapper[T]) BuildForQuery() *gorm.DB {
	if o.Error != nil {
		return nil
	}

	return o.builder.BuildForQuery()
}

//todo https://gorm.io/zh_CN/docs/query.html
//todo Joins 一个衍生表
