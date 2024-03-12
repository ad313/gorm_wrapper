package orm

import "errors"

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
