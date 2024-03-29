package orm

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"strings"
)

// PagerList 分页数据结果模型
type PagerList[T interface{}] struct {
	Page       int32  `json:"page" form:"page"`               //页码
	PageSize   int32  `json:"page_size" form:"page_size"`     //分页条数
	TotalCount int32  `json:"total_count" form:"total_count"` //总条数
	Order      string `json:"order" form:"order"`             //排序字段
	Data       []*T   `json:"data" form:"data"`               //数据项
}

// Pager 分页数据请求模型
type Pager struct {
	Page     int32  `json:"page" form:"page"`           //页码
	PageSize int32  `json:"page_size" form:"page_size"` //分页条数
	Order    string `json:"order" form:"order"`         //排序字段
	Keyword  string `json:"keyword" form:"keyword"`     //关键词
}

type ormWrapperBuilder[T interface{}] struct {
	wrapper *OrmWrapper[T]

	TableName  string   //表名
	TableAlias string   //表别名
	childTable *gorm.DB //查询衍生表，为空则查询主表

	ctx        context.Context //context
	DbContext  *gorm.DB        //操作后的db
	_DbContext *gorm.DB        //最初始的db
	isOuterDb  bool            //是否外部传入db

	where           [][]any        //普通条件
	WhereCondition  []OrmCondition //condition条件
	HavingCondition []OrmCondition //having条件
	joinModels      []*joinModel   //leftJoin 集合
	selectColumns   []string       //select 字段集合
	selectModes     []*selectMode  //select 字段模型集合
	groupByColumns  []string       //group by 字段集合
	orderByColumns  []string       //order by 字段集合

	isUnscoped bool //和gorm一样，忽略软删除字段
}

// joinModel 连表条件，会自动加上软删除字段
type joinModel struct {
	Table     schema.Tabler //连接表，右表
	Db        *gorm.DB      //连接衍生表
	tableName string        //右表表名
	Alias     string        //右表表别名
	Left      string        //左表字段
	Right     string        //右表字段
	ext       string        //扩展字段，比如有软删除字段，这里加上软删除sql
	joinKey   string        //left join;inner join
}

func (o *ormWrapperBuilder[T]) addWhere(query interface{}, args []interface{}) {
	if o.where == nil {
		o.where = make([][]interface{}, 0)
	}

	if query == nil {
		o.wrapper.Error = errors.New("query 条件不能为空")
		return
	}

	o.where = append(o.where, append([]interface{}{query}, args...))
}

func (o *ormWrapperBuilder[T]) addWhereWithWhereCondition(condition OrmCondition) {
	if o.WhereCondition == nil {
		o.WhereCondition = make([]OrmCondition, 0)
	}

	if condition == nil {
		o.wrapper.Error = errors.New("where condition 不能为空")
	}

	o.WhereCondition = append(o.WhereCondition, condition)
}

func (o *ormWrapperBuilder[T]) addHavingWithWhereCondition(condition OrmCondition) {
	if o.HavingCondition == nil {
		o.HavingCondition = make([]OrmCondition, 0)
	}

	if condition == nil {
		o.wrapper.Error = errors.New("having condition 不能为空")
	}

	o.HavingCondition = append(o.HavingCondition, condition)
}

func (o *ormWrapperBuilder[T]) mergeColumnName(column string, columnAlias string, tableAlias string) string {
	if tableAlias != "" {
		column = formatSqlName(tableAlias, _dbType) + "." + column
	}

	if columnAlias != "" {
		column += " as " + getSqlSm(_dbType) + columnAlias + getSqlSm(_dbType)
	}

	return column
}

func (o *ormWrapperBuilder[T]) mergeColumnNameWithFunc(column string, columnAlias string, tableAlias string, f string) string {
	var sql, _ = mergeTableColumnWithFunc(column, tableAlias, f, _dbType)
	if columnAlias != "" {
		sql += " as " + getSqlSm(_dbType) + columnAlias + getSqlSm(_dbType)
	}

	return sql
}

func (o *ormWrapperBuilder[T]) getPrimaryKey() (string, error) {
	if o.wrapper.tableInfo.PrimaryKeyName == "" {
		o.wrapper.Error = errors.New(fmt.Sprintf("表 %v 没有主键，操作失败", o.wrapper.tableInfo.Name))
		return "", o.wrapper.Error
	}

	return o.wrapper.tableInfo.PrimaryKeyName, nil
}

// 设置主表，针对没有主动设置表别名，这里自动加上表名称做表别名
func (o *ormWrapperBuilder[T]) buildModel() {
	//没有手动设置表别名，这里判断是否需要加：left join、exists
	if o.TableAlias == "" {
		//leftJoin
		if len(o.joinModels) > 0 {
			o.TableAlias = o.TableName
		} else {
			//exists
			if len(o.WhereCondition) > 0 {
				for _, condition := range o.WhereCondition {
					_, ok := condition.(*ExistsCondition)
					if ok {
						o.TableAlias = o.TableName
					}
				}
			}
		}
	}

	var addSoftDelCondition = func(builder *ormWrapperBuilder[T], table schema.Tabler, tableAlias string) {
		if builder.isUnscoped {
			return
		}

		softDel, err := getTableSoftDeleteColumnSql(table, tableAlias, _dbType)
		if err != nil {
			builder.wrapper.Error = err
			return
		}

		builder.addWhereWithWhereCondition(&OriginalCondition{
			Sql: softDel,
			Arg: nil,
		})
	}

	if o.TableAlias == "" {
		o.DbContext = o.DbContext.Model(new(T)).Unscoped()
		addSoftDelCondition(o, o.wrapper.table, "")
		return
	}

	//衍生表
	if o.childTable != nil {
		o.DbContext = o.DbContext.Unscoped().Table("(?) as "+formatTableAlias(o.TableAlias), o.childTable)
	} else {
		//这里的 TableAlias 不能包装前后缀，否则无法识别-- 突然又好了
		o.DbContext = o.DbContext.Unscoped().Table(formatSqlName(o.TableName, _dbType) + " as " + formatTableAlias(o.TableAlias))
		addSoftDelCondition(o, o.wrapper.table, o.TableAlias)
	}
}

func (o *ormWrapperBuilder[T]) buildWhere() {
	if o.where == nil {
		o.where = make([][]interface{}, 0)
	}

	if len(o.WhereCondition) > 0 {
		for _, condition := range o.WhereCondition {
			sql, param, err := condition.BuildSql(_dbType, o.isUnscoped)
			if err != nil {
				o.wrapper.Error = errors.New("query 条件不能为空")
				return
			}

			if strings.HasPrefix(sql, "(") && strings.HasSuffix(sql, ")") {
				sql = strings.TrimPrefix(sql, "(")
				sql = strings.TrimSuffix(sql, ")")
			}

			o.addWhere(sql, param)
		}
	}

	for _, items := range o.where {
		if len(items) == 0 {
			continue
		}

		if len(items) == 1 {
			o.DbContext = o.DbContext.Where(items[0])
		} else {
			o.DbContext = o.DbContext.Where(items[0], items[1:]...)
		}
	}
}

func (o *ormWrapperBuilder[T]) buildJoin() {
	if len(o.joinModels) > 0 {
		for _, join := range o.joinModels {
			//衍生表
			if join.Db != nil {
				var sql = fmt.Sprintf("%v (?) as %v on %v = %v%v",
					join.joinKey,
					formatTableAlias(join.Alias),
					join.Left,
					join.Right,
					chooseTrueValue(o.isUnscoped, "", join.ext))
				o.DbContext = o.DbContext.
					Joins(sql, join.Db)
			} else {
				o.DbContext = o.DbContext.
					Joins(fmt.Sprintf("%v %v as %v on %v = %v%v",
						join.joinKey,
						formatSqlName(join.tableName, _dbType),
						formatTableAlias(join.Alias),
						join.Left,
						join.Right,
						chooseTrueValue(o.isUnscoped, "", join.ext)))
			}
		}
	}
}

func (o *ormWrapperBuilder[T]) buildSelect() {

	for _, mode := range o.selectModes {

		//todo 子查询
		if db, ok := isDbValue(mode.Column); ok {
			if mode.ColumnAlias == "" {
				o.wrapper.Error = errors.New("当子查询作为选择字段时，必须给字段取别名")
				return
			}

			var sql = db.ToSQL(func(tx *gorm.DB) *gorm.DB {
				return tx.First(new(T))
			})
			if sql == "" {
				o.wrapper.Error = errors.New("select 子查询不正确")
				return
			}

			o.selectColumns = append(o.selectColumns, fmt.Sprintf("(%v) as %v", sql, formatSqlName(mode.ColumnAlias, _dbType)))
			//o.DbContext = o.DbContext.Select("(?) as ?", mode.Column, formatSqlName(mode.ColumnAlias, _dbType))
			continue
		}

		//字段
		var name = GetTableColumn(mode.Column)
		if name != "" {
			name = formatSqlName(name, _dbType)
		} else {
			name = GetString(mode.Column)
		}

		if name == "" {
			o.wrapper.Error = errors.New("选择字段不能为空")
			return
		}

		sql, err := mergeTableColumnWithFunc2(name, mode.TableAlias, mode.Func, _dbType)
		if err != nil {
			o.wrapper.Error = err
			return
		}

		if mode.ColumnAlias != "" {
			sql += " as " + formatSqlName(mode.ColumnAlias, _dbType)
		}

		o.selectColumns = append(o.selectColumns, sql)
	}
	o.DbContext = o.DbContext.Select(o.selectColumns)
}

func (o *ormWrapperBuilder[T]) buildOrderBy() {
	if o.orderByColumns != nil && len(o.orderByColumns) > 0 {
		o.DbContext = o.DbContext.Order(strings.Join(o.orderByColumns, ","))
	}
}

func (o *ormWrapperBuilder[T]) buildGroupBy() {
	if o.groupByColumns != nil && len(o.groupByColumns) > 0 {
		//特殊处理一个参数的情况，否则报错
		if len(o.groupByColumns) == 1 {
			o.DbContext = o.DbContext.Group(strings.ReplaceAll(o.groupByColumns[0], getSqlSm(_dbType), ""))
		} else {
			o.DbContext = o.DbContext.Group(strings.Join(o.groupByColumns, ","))
		}
	}
}

func (o *ormWrapperBuilder[T]) buildHaving() {
	if len(o.HavingCondition) > 0 {
		for _, condition := range o.HavingCondition {
			sql, param, err := condition.BuildSql(_dbType)
			if err != nil {
				o.wrapper.Error = err
				return
			}

			o.DbContext = o.DbContext.Having(sql, param...)
		}
	}
}

// Build 创建 gorm sql
func (o *ormWrapperBuilder[T]) Build() *gorm.DB {
	o.buildWhere()
	o.buildSelect()
	o.buildJoin()
	o.buildOrderBy()
	o.buildGroupBy()
	o.buildHaving()
	return o.DbContext
}

// BuildForQuery 创建 gorm sql
func (o *ormWrapperBuilder[T]) BuildForQuery() *gorm.DB {
	o.buildModel()
	o.Build()
	return o.DbContext
}

func formatTableAlias(alias string) string {
	//return formatSqlName(alias, _dbType)
	return alias
}
