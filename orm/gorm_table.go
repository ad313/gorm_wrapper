package orm

import (
	"errors"
	"fmt"
	"github.com/ad313/gorm_wrapper/orm/ref"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/soft_delete"
	"reflect"
	"sync"
)

// 表缓存
var tableCache sync.Map
var tableSchemaCache sync.Map

// 字段缓存
var columnCache = make(map[uintptr]string)

// TableInfo 表信息
type TableInfo[T interface{}] struct {
	T *T
	*TableSchema
}

type TableSchema struct {
	Name              string //表名
	DeletedColumnName string //软删除字段名
	DeleteCondition   string //软删除 片段sql：IS NULL，=0
	PrimaryKeyName    string //主键名称
}

// OrmTableResult 创建表缓存结果
type OrmTableResult[T interface{}] struct {
	Table *TableInfo[T]
	Error error
}

// BuildOrmTable 获取
func BuildOrmTable[T interface{}]() *OrmTableResult[T] {
	modelType := ref.GetPath[T]()
	if model, ok := tableCache.Load(modelType); ok {
		m, isReal := model.(*TableInfo[T])
		if isReal {
			return &OrmTableResult[T]{Table: m}
		}
	}

	//验证 schema.Tabler
	t, table, ok := ref.IsType[T, schema.Tabler]()
	if !ok {
		return &OrmTableResult[T]{Error: errors.New("T 必须实现 schema.Tabler")}
	}

	//处理字段
	m, tableSchema := getColumnNameMap(t)
	for key, v := range m {
		columnCache[key] = v
	}

	//缓存
	tableSchema.Name = (*table).TableName()
	var cache = &TableInfo[T]{
		T:           t,
		TableSchema: tableSchema,
	}
	tableCache.Store(modelType, cache)
	tableSchemaCache.Store(modelType, cache.TableSchema)

	return &OrmTableResult[T]{Table: cache}
}

// GetTableColumn 通过模型字段获取数据库字段
func GetTableColumn(column any) string {
	var v = reflect.ValueOf(column)
	var addr uintptr
	if v.Kind() == reflect.Pointer {
		addr = v.Pointer()
		n, ok := columnCache[addr]
		if ok {
			return n
		}
	} else {
		log("column must be of type Pointer")
		return ""
	}

	return ""
}

func GetString(column any) string {
	if str, ok := column.(string); ok {
		return str
	}

	return ""
}

// GetTableSchema 获取表元数据
func GetTableSchema(table schema.Tabler) *TableSchema {
	modelType := ref.GetPathByValue(table)
	if model, ok := tableSchemaCache.Load(modelType); ok {
		m, isReal := model.(*TableSchema)
		if isReal {
			return m
		}
	}

	return nil
}

func getColumnNameMap(model any) (map[uintptr]string, *TableSchema) {
	var columnNameMap = make(map[uintptr]string)
	valueOf := reflect.ValueOf(model).Elem()
	typeOf := reflect.TypeOf(model).Elem()

	var tableSchema = &TableSchema{}
	var childrenTableSchema = &TableSchema{}

	for i := 0; i < valueOf.NumField(); i++ {
		field := typeOf.Field(i)
		// 如果当前实体嵌入了其他实体，同样需要缓存它的字段名
		if field.Anonymous {
			// 如果存在多重嵌套，通过递归方式获取他们的字段名
			subFieldMap, _child := getSubFieldColumnNameMap(valueOf, field)
			for pointer, columnName := range subFieldMap {
				columnNameMap[pointer] = columnName
			}

			if childrenTableSchema.DeletedColumnName == "" && _child.DeletedColumnName != "" {
				childrenTableSchema.DeletedColumnName = _child.DeletedColumnName
			}
		} else {
			// 获取对象字段指针值
			pointer := valueOf.Field(i).Addr().Pointer()
			columnName := parseColumnName(field)
			if columnName != "" {
				columnNameMap[pointer] = columnName
			}

			//判断软删除字段
			if sql, ok := isGormDeletedAt(field, valueOf); ok {
				tableSchema.DeletedColumnName = columnName
				tableSchema.DeleteCondition = sql
			}

			//主键
			var key = parsePrimaryKeyName(field)
			if key != "" && columnName != "" {
				tableSchema.PrimaryKeyName = columnName
			}
		}
	}

	//优先用本级的软删除字段
	if tableSchema.DeletedColumnName == "" {
		tableSchema.DeletedColumnName = childrenTableSchema.DeletedColumnName
		tableSchema.DeleteCondition = childrenTableSchema.DeleteCondition
	}

	return columnNameMap, tableSchema
}

func isGormDeletedAt(field reflect.StructField, valueOf reflect.Value) (string, bool) {
	//判断软删除字段
	_, ok := ref.IsTypeByValue[gorm.DeletedAt](valueOf.FieldByName(field.Name).Interface())
	if ok {
		return " IS NULL", true
	}

	//数字类型
	_, ok = ref.IsTypeByValue[soft_delete.DeletedAt](valueOf.FieldByName(field.Name).Interface())
	if ok {
		return " = 0", true
	}

	return "", false
}

// 递归获取嵌套字段名
func getSubFieldColumnNameMap(valueOf reflect.Value, field reflect.StructField) (map[uintptr]string, *TableSchema) {
	result := make(map[uintptr]string)
	modelType := field.Type
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	var tableSchema = &TableSchema{}
	var childrenTableSchema = &TableSchema{}

	for j := 0; j < modelType.NumField(); j++ {
		subField := modelType.Field(j)
		if subField.Anonymous {
			nestedFields, _child := getSubFieldColumnNameMap(valueOf, subField)
			for key, value := range nestedFields {
				result[key] = value
			}

			if childrenTableSchema.DeletedColumnName == "" && _child.DeletedColumnName != "" {
				childrenTableSchema.DeletedColumnName = _child.DeletedColumnName
			}
			if childrenTableSchema.PrimaryKeyName == "" && _child.PrimaryKeyName != "" {
				childrenTableSchema.PrimaryKeyName = _child.PrimaryKeyName
			}
		} else {
			pointer := valueOf.FieldByName(modelType.Field(j).Name).Addr().Pointer()
			name := parseColumnName(modelType.Field(j))
			result[pointer] = name

			//判断软删除字段
			if sql, ok := isGormDeletedAt(subField, valueOf); ok {
				tableSchema.DeletedColumnName = name
				tableSchema.DeleteCondition = sql
			}

			//主键
			var key = parsePrimaryKeyName(subField)
			if key != "" {
				tableSchema.PrimaryKeyName = key
			}
		}
	}

	//优先用本级的软删除字段
	if tableSchema.DeletedColumnName == "" {
		tableSchema.DeletedColumnName = childrenTableSchema.DeletedColumnName
		tableSchema.DeleteCondition = childrenTableSchema.DeleteCondition
	}
	if tableSchema.PrimaryKeyName == "" {
		tableSchema.PrimaryKeyName = childrenTableSchema.PrimaryKeyName
	}

	return result, tableSchema
}

// 解析字段名称
func parseColumnName(field reflect.StructField) string {
	tagSetting := schema.ParseTagSetting(field.Tag.Get("gorm"), ";")
	name, ok := tagSetting["COLUMN"]
	if ok {
		return name
	}
	return ""
}

// 解析主键名称
func parsePrimaryKeyName(field reflect.StructField) string {
	tagSetting := schema.ParseTagSetting(field.Tag.Get("gorm"), ";")
	name, ok := tagSetting["PRIMARYKEY"]
	if ok {
		return name
	}
	return ""
}

// getSqlSm 获取sql 中 数据库字段分隔符
func getSqlSm(dbType string) string {
	switch dbType {
	case MySql:
		return "`"
	case Sqlite:
		return "'"
	case Dm:
		return "\""
	case Postgres:
		return "\""
	case Sqlserver:
		return "\""
	default:
		break
	}

	return ""
}

// 记录日志
func log(content string) {
	fmt.Println(content)
}
