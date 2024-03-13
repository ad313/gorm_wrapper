package orm

import (
	"errors"
	"strings"
)

// OrmCondition 定义查询条件
type OrmCondition interface {
	BuildSql(dbType string, ext ...interface{}) (string, []interface{}, error) //生成 sql
}

// ConditionBuilder 条件构建器
type ConditionBuilder struct {
	Or    bool           //and、or
	Items []OrmCondition //条件集合
	error error
}

// NewAnd 创建 Builder，当条件个数为1，则加在builder本身，大于1，则加在Items；内部关系是 And
func NewAnd(conditions ...OrmCondition) *ConditionBuilder {
	return newConditionBuilder(false, conditions...)
}

// NewOr 创建 Builder，当条件个数为1，则加在builder本身，大于1，则加在Items；内部关系是 Or
func NewOr(conditions ...OrmCondition) *ConditionBuilder {
	return newConditionBuilder(true, conditions...)
}

// Add builder 添加子条件
func (c *ConditionBuilder) Add(conditions ...OrmCondition) *ConditionBuilder {
	if c.Items == nil {
		c.Items = make([]OrmCondition, 0)
	}

	if len(conditions) > 0 {
		c.Items = append(c.Items, conditions...)
	}

	return c
}

// BuildSql 生成sql
func (c *ConditionBuilder) BuildSql(dbType string, extend ...interface{}) (string, []interface{}, error) {
	if c == nil || len(c.Items) == 0 {
		return "", nil, errors.New("没有任何条件")
	}

	var _sql = make([]string, 0)
	var _param = make([]interface{}, 0)
	for _, item := range c.Items {
		sql, param, err := item.BuildSql(dbType, extend)
		if err != nil {
			return "", nil, err
		}

		_sql = append(_sql, sql)
		_param = append(_param, param...)
	}

	//条件符号
	var compareSymbols = chooseTrueValue(c.Or, " OR ", " AND ")

	//一个条件时可省略括号
	if len(_sql) == 1 {
		return _sql[0], _param, nil
	} else {
		return "(" + strings.Join(_sql, compareSymbols) + ")", _param, nil
	}
}

func (c *ConditionBuilder) Error(error string) *ConditionBuilder {
	c.error = errors.New(error)
	return c
}

// 创建 Builder，当条件个数为1，则加在builder本身，大于1，则加在Items
func newConditionBuilder(or bool, conditions ...OrmCondition) *ConditionBuilder {
	var builder = &ConditionBuilder{
		Or:    or,
		Items: make([]OrmCondition, 0),
		error: nil,
	}

	return builder.Add(conditions...)
}

//
//// ConditionBuilder 条件构建器
//type ConditionBuilder struct {
//	Or      bool                //and、or
//	Items   []*ConditionBuilder //条件集合
//	Current OrmCondition      //当前条件
//	error   error
//}
//
//func NewAndEmpty() *ConditionBuilder {
//	return &ConditionBuilder{
//		Or:      false,
//		Items:   make([]*ConditionBuilder, 0),
//		Current: nil,
//		error:   nil,
//	}
//}
//
//func NewOrEmpty() *ConditionBuilder {
//	return &ConditionBuilder{
//		Or:      true,
//		Items:   make([]*ConditionBuilder, 0),
//		Current: nil,
//		error:   nil,
//	}
//}
//
//// NewAnd 创建 Builder，当条件个数为1，则加在builder本身，大于1，则加在Items；内部关系是 And
//func NewAnd(conditions ...OrmCondition) *ConditionBuilder {
//	return newConditionBuilder(false, conditions...)
//}
//
//// NewOr 创建 Builder，当条件个数为1，则加在builder本身，大于1，则加在Items；内部关系是 Or
//func NewOr(conditions ...OrmCondition) *ConditionBuilder {
//	return newConditionBuilder(true, conditions...)
//}
//
//// SetCondition builder 设置本级条件
//func (c *ConditionBuilder) SetCondition(condition OrmCondition) *ConditionBuilder {
//	c.Current = condition
//	return c
//}
//
//// AddChildrenBuilder builder 添加子条件
//func (c *ConditionBuilder) AddChildrenBuilder(builders ...*ConditionBuilder) *ConditionBuilder {
//	if len(builders) == 0 {
//		return c.Error("AddChildrenBuilder builders is empty")
//	}
//
//	c.Items = append(c.Items, builders...)
//	return c
//}
//
//// AddChildrenCondition builder 添加子条件
//func (c *ConditionBuilder) AddChildrenCondition(conditions ...OrmCondition) *ConditionBuilder {
//	if len(conditions) == 0 {
//		return c.Error("AddChildrenBuilder conditions is empty")
//	}
//
//	for _, condition := range conditions {
//		c.AddChildrenBuilder(&ConditionBuilder{Current: condition})
//	}
//
//	return c
//}
//
//// BuildSql 生成sql
//func (c *ConditionBuilder) BuildSql(dbType string, extend ...interface{}) (string, []interface{}, error) {
//	if c == nil {
//		return "", nil, errors.New("没有任何条件")
//	}
//
//	//没有子项，条件就是本身；有子项则用子项
//	if len(c.Items) == 0 {
//		if c.Current == nil {
//			return "", nil, errors.New("没有任何条件")
//		}
//
//		return c.Current.BuildSql(dbType)
//	}
//
//	var _sql = make([]string, 0)
//	var _param = make([]interface{}, 0)
//	for _, item := range c.Items {
//		sql, param, err := item.BuildSql(dbType, extend)
//		if err != nil {
//			return "", nil, err
//		}
//
//		_sql = append(_sql, sql)
//		_param = append(_param, param...)
//	}
//
//	//条件符号
//	var compareSymbols = chooseTrueValue(c.Or, " OR ", " AND ")
//
//	//一个条件时可省略括号
//	if len(_sql) == 1 {
//		return _sql[0], _param, nil
//	} else {
//		return "(" + strings.Join(_sql, compareSymbols) + ")", _param, nil
//	}
//}
//
//func (c *ConditionBuilder) Error(error string) *ConditionBuilder {
//	c.error = errors.New(error)
//	return c
//}
//
//// 创建 Builder，当条件个数为1，则加在builder本身，大于1，则加在Items
//func newConditionBuilder(or bool, conditions ...OrmCondition) *ConditionBuilder {
//	var builder = &ConditionBuilder{
//		Or:    or,
//		Items: nil,
//		error: nil,
//	}
//
//	if len(conditions) == 0 {
//		return builder.Error("newConditionBuilder conditions has no item")
//	}
//
//	if len(conditions) == 1 {
//		return builder.SetCondition(conditions[0])
//	}
//
//	return builder.AddChildrenCondition(conditions...)
//}
