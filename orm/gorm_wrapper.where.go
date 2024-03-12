package orm

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
