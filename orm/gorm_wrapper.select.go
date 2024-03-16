package orm

// 查询字段模型
type selectMode struct {
	Column      interface{} //字段、字符串、子查询*gorm.DB
	ColumnAlias string      //字段别名
	TableAlias  string      //表别名
	Func        string      //数据库函数
}

// Select 选择多个字段；传入字段、字符串；如果是字符串，则对字符串不做任何处理，原样查询
func (o *OrmWrapper[T]) Select(columns ...interface{}) *OrmWrapper[T] {

	//todo 这里不能是子查询，因为没有取别名

	return o.SelectTable("", columns...)
}

// SelectTable 在 Select 基础上传入表别名
func (o *OrmWrapper[T]) SelectTable(tableAlias string, columns ...interface{}) *OrmWrapper[T] {
	if len(columns) == 0 {
		return o
	}

	//if tableAlias == "" {
	//	o.Error = errors.New("[SelectTable] 参数 tableAlias 不能为空")
	//}

	for _, column := range columns {
		o.builder.selectModes = append(o.builder.selectModes, &selectMode{Column: column, TableAlias: tableAlias})
	}

	return o
}

// SelectOne 选择一个字段（字段、字符串、子查询*gorm.DB），可传入 字段别名，表名；可多次调用
func (o *OrmWrapper[T]) SelectOne(column interface{}, columnAlias string, tableAlias ...string) *OrmWrapper[T] {
	return o.SelectOneWithFunc(column, columnAlias, "", tableAlias...)
}

// SelectOneWithFunc 在 SelectOne 的基础上传入函数包装，比如 Max、Min、Count 等
func (o *OrmWrapper[T]) SelectOneWithFunc(column interface{}, columnAlias string, f string, tableAlias ...string) *OrmWrapper[T] {
	o.builder.selectModes = append(o.builder.selectModes, &selectMode{Column: column, ColumnAlias: columnAlias, Func: f, TableAlias: FirstOrDefault(tableAlias)})
	return o
}

//// Select 选择多个字段；传入字段、字符串；如果是字符串，则对字符串不做任何处理，原样查询
//func (o *OrmWrapper[T]) Select(columns ...interface{}) *OrmWrapper[T] {
//	if columns == nil || len(columns) == 0 {
//		return o
//	}
//
//	//判断是否有 Join，如果有则给字段名加上主表别名
//	var table = chooseTrueValue(len(o.builder.joinModels) == 0, "", o.builder.TableAlias)
//
//	return o.SelectTable(table, columns...)
//}
//
//// SelectTable 在 Select 基础上传入表别名
//func (o *OrmWrapper[T]) SelectTable(tableAlias string, columns ...interface{}) *OrmWrapper[T] {
//	if columns == nil || len(columns) == 0 {
//		return o
//	}
//
//	for _, column := range columns {
//		name, err := resolveColumnName(column, _dbType)
//		if err != nil || name == "" {
//			o.Error = errors.New("未获取到字段名称")
//			continue
//		}
//
//		o.builder.selectColumns = append(o.builder.selectColumns, o.builder.mergeColumnName(name, "", tableAlias))
//	}
//
//	return o
//}
//
//// SelectOne 选择一个字段（字段、字符串、子查询），可传入 字段别名，表名；可多次调用
//func (o *OrmWrapper[T]) SelectOne(column interface{}, columnAlias string, tableAlias ...string) *OrmWrapper[T] {
//	name, err := resolveColumnName(column, _dbType)
//	if err != nil || name == "" {
//		o.Error = errors.New("未获取到字段名称")
//		return o
//	}
//
//	o.builder.selectColumns = append(o.builder.selectColumns, o.builder.mergeColumnName(name, columnAlias, FirstOrDefault(tableAlias)))
//
//	return o
//}
//
//// SelectOneWithFunc 在 SelectOne 的基础上传入函数包装，比如 Max、Min、Count 等
//func (o *OrmWrapper[T]) SelectOneWithFunc(column interface{}, columnAlias string, f string, tableAlias ...string) *OrmWrapper[T] {
//	name, err := resolveColumnName(column, "")
//	if err != nil || name == "" {
//		o.Error = errors.New("未获取到字段名称")
//		return o
//	}
//
//	var table = FirstOrDefault(tableAlias)
//	o.builder.selectColumns = append(o.builder.selectColumns, o.builder.mergeColumnNameWithFunc(name, columnAlias, table, f))
//
//	return o
//}
