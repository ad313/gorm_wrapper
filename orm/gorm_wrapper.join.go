package orm

import (
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// LeftJoin 左连表
func (o *OrmWrapper[T]) LeftJoin(table schema.Tabler, alias string, leftColumn any, rightColumn any) *OrmWrapper[T] {
	return o.Join(table, alias, leftColumn, rightColumn, LeftJoin)
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
	return o.Join(table, alias, leftColumn, rightColumn, InnerJoin)
}

// InnerJoinIf 内连表
func (o *OrmWrapper[T]) InnerJoinIf(do bool, table schema.Tabler, alias string, leftColumn any, rightColumn any) *OrmWrapper[T] {
	if do {
		return o.InnerJoin(table, alias, leftColumn, rightColumn)
	}

	return o
}

// Join 连表
func (o *OrmWrapper[T]) Join(table schema.Tabler, alias string, leftColumn any, rightColumn any, joinKey string) *OrmWrapper[T] {
	if o.builder.joinModels == nil {
		o.builder.joinModels = make([]*joinModel, 0)
	}

	if joinKey == "" {
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
		joinKey:   joinKey,
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

// JoinIf 连表
func (o *OrmWrapper[T]) JoinIf(do bool, table schema.Tabler, alias string, leftColumn any, rightColumn any, joinKey string) *OrmWrapper[T] {
	if do {
		return o.Join(table, alias, leftColumn, rightColumn, joinKey)
	}

	return o
}

// JoinChildTable 连衍生表
func (o *OrmWrapper[T]) JoinChildTable(db *gorm.DB, alias string, leftColumn any, rightColumn any, joinKey string) *OrmWrapper[T] {
	if o.builder.joinModels == nil {
		o.builder.joinModels = make([]*joinModel, 0)
	}

	if joinKey == "" {
		o.Error = errors.New("join 连表参数不正确")
		return o
	}

	if db == nil || leftColumn == nil || rightColumn == nil || alias == "" {
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

	var leftTableName string
	var lastJoin *joinModel
	if len(o.builder.joinModels) == 0 {
		leftTableName = chooseTrueValue(o.builder.TableAlias != "", o.builder.TableAlias, o.builder.TableName)
	} else {
		lastJoin = o.builder.joinModels[len(o.builder.joinModels)-1]
		leftTableName = lastJoin.Alias
	}

	var join = &joinModel{
		Db:        db,
		tableName: "",
		Alias:     alias,
		Left:      formatSqlName(leftTableName, _dbType) + "." + left,
		Right:     formatSqlName(alias, _dbType) + "." + right,
		joinKey:   joinKey,
	}

	//如果上一个表不是衍生表，判断软删除字段
	if !o.builder.isUnscoped && (lastJoin != nil && lastJoin.Db == nil) {
		leftSoftDel, err := getTableSoftDeleteColumnSql(lastJoin.Table, leftTableName, _dbType)
		if err != nil {
			o.Error = err
			return o
		}

		if leftSoftDel != "" {
			join.ext = " AND " + leftSoftDel
		}

	}

	o.builder.joinModels = append(o.builder.joinModels, join)

	return o
}

// JoinChildTableIf 连衍生表
func (o *OrmWrapper[T]) JoinChildTableIf(do bool, table *gorm.DB, alias string, leftColumn any, rightColumn any, joinKey string) *OrmWrapper[T] {
	if do {
		return o.JoinChildTable(table, alias, leftColumn, rightColumn, joinKey)
	}

	return o
}
