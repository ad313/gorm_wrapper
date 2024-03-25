package orm

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"reflect"

	"gorm.io/gorm/schema"
)

// ResolveTableColumnName 从缓存获取数据库字段名称：如果不能匹配，则返回 string 值
func ResolveTableColumnName(column any, dbType string) string {
	var kind = reflect.ValueOf(column).Kind()
	if kind == reflect.Pointer {
		var name = GetTableColumn(column)
		if name == "" {
			return ""
		}

		return getSqlSm(dbType) + name + getSqlSm(dbType)
	} else {
		if str, ok := column.(string); ok && str != "" {
			if str == "*" {
				return "*"
			}
			if str == "1" {
				return "1"
			}
			return getSqlSm(dbType) + str + getSqlSm(dbType)
		} else {
			return ""
		}
	}
}

// mergeWhereString 组合 where 条件
func mergeWhereString(column any, compareSymbols string, tableAlias string, f string, dbType string, arg interface{}) (string, error) {
	name, err := resolveColumnName(column, dbType)
	if err != nil {
		return "", err
	}

	var valueExpress = "?"
	switch compareSymbols {
	case "IN":
		valueExpress = "(?)"
		break
	case "NOT IN":
		valueExpress = "(?)"
		break
	case "IS NULL":
		valueExpress = ""
		break
	case "IS NOT NULL":
		valueExpress = ""
		break
	}

	//判断是否是子查询
	if arg != nil {
		if isDb(arg) {
			valueExpress = "(?)"
		}
	}

	var table = ""
	if tableAlias != "" {
		table = formatSqlName(tableAlias, dbType) + "."
	}

	name = table + name
	return fmt.Sprintf("%v %v %v", mergeNameAndFunc(name, f), getCompareSymbols(compareSymbols), valueExpress), nil
}

// mergeWhereValue 处理查询值，当 IN 条件时，拼接 %
func mergeWhereValue(compareSymbols string, value interface{}) interface{} {
	if value == nil {
		return value
	}
	v, ok := value.(string)
	if !ok {
		return value
	}

	switch compareSymbols {
	case "Like":
		return "%" + v + "%"
	case "NOT Like":
		return "%" + v + "%"
	case "STARTWITH":
		return v + "%"
	case "ENDWITH":
		return "%" + v
	}

	return value
}

// 检查各种条件下参数是否为空
func checkParam(compareSymbols string, value interface{}) (interface{}, error) {
	if compareSymbols == "" {
		return nil, errors.New("compareSymbols 不能为空")
	}

	var check = true
	switch compareSymbols {
	case "Like":
		break
	case "NOT Like":
		break
	case "STARTWITH":
		break
	case "ENDWITH":
		break
	case "IN":
		break
	case "NOT IN":
		break
	case "IS NULL":
		check = false
		break
	case "IS NOT NULL":
		check = false
		break
	}

	//不需要参数
	if !check {
		value = nil
	}

	if check && value == nil {
		return nil, errors.New("参数不能为空")
	}

	return value, nil
}

func getCompareSymbols(compareSymbols string) string {
	switch compareSymbols {
	case "Like":
		return compareSymbols
	case "NOT Like":
		return compareSymbols
	case "STARTWITH":
		return "Like"
	case "ENDWITH":
		return "Like"
	}

	return compareSymbols
}

// resolveColumnName 从缓存获取数据库字段名称：如果不能匹配，则返回 string 值
func resolveColumnName(column any, dbType string) (string, error) {
	var name = ResolveTableColumnName(column, dbType)
	if name == "" {
		return "", errors.New("未获取到字段名称")
	}
	return name, nil
}

// 处理数据库表名
func formatSqlName(alias string, dbType string) string {
	if alias == "" {
		return alias
	}

	return getSqlSm(dbType) + alias + getSqlSm(dbType)
}

//// 处理数据库表名 加上别名
//func mergeTableWithAlias(table string, alias string, dbType string) string {
//	if table == "" {
//		return table
//	}
//
//	table = getSqlSm(dbType) + table + getSqlSm(dbType)
//
//	if alias != "" {
//		table += " as " + alias
//	}
//
//	return table
//}

//// 处理数据库表名 加上别名
//func mergeTableWithAliasByValue(table schema.Tabler, alias string, dbType string) string {
//	return mergeTableWithAlias(table.TableName(), alias, dbType)
//}

// 获取软删除字段
func getTableSoftDeleteColumnSql(table schema.Tabler, tableAlias string, dbType string) (string, error) {
	var tableSchema = GetTableSchema(table)
	if tableSchema != nil && tableSchema.DeletedColumnName != "" && tableSchema.DeleteCondition != "" {
		n, err := resolveColumnName(tableSchema.DeletedColumnName, dbType)
		if err != nil {
			return "", err
		}

		if tableAlias != "" {
			n = formatSqlName(tableAlias, _dbType) + "." + n
		}

		return n + tableSchema.DeleteCondition, nil
	}

	return "", nil
}

func mergeTableColumnWithFunc(column interface{}, table string, f string, dbType string) (string, error) {
	name, err := resolveColumnName(column, dbType)
	if err != nil {
		return "", err
	}

	return chooseTrueValue(table != "", mergeNameAndFunc(formatSqlName(table, dbType)+"."+name, f), mergeNameAndFunc(name, f)), nil
}

func mergeTableColumnWithFunc2(columnName string, table string, f string, dbType string) (string, error) {
	if columnName == "" {
		return "", errors.New("column 不能为空")
	}

	return chooseTrueValue(table != "", mergeNameAndFunc(formatSqlName(table, dbType)+"."+columnName, f), mergeNameAndFunc(columnName, f)), nil
}

// 合并字段和数据库函数
func mergeNameAndFunc(name, f string) string {
	return chooseTrueValue(f == "", name, f+"("+name+")")
}

// chooseTrueValue 模拟三元表达式，获取值
func chooseTrueValue[T interface{}](condition bool, trueValue, falseValue T) T {
	if condition {
		return trueValue
	}
	return falseValue
}

func FirstOrDefault[T interface{}](slice []T) T {
	if len(slice) > 0 {
		return slice[0]
	}
	return *new(T)
}

// 判断是否是 *gorm.DB
func isDb(arg interface{}) bool {
	_, ok := isDbValue(arg)
	return ok
}

// 判断是否是 *gorm.DB
func isDbValue(arg interface{}) (*gorm.DB, bool) {
	db, ok := arg.(*gorm.DB)
	return db, ok
}
