package main

//
//import (
//	"context"
//	"fmt"
//	"github.com/ad313/gorm_wrapper/orm"
//	"github.com/google/uuid"
//	"gorm.io/driver/mysql"
//	"gorm.io/gorm"
//	"gorm.io/plugin/soft_delete"
//	"strconv"
//)
//
//var db *gorm.DB
//
//var mysqlConn = "root:123456@tcp(192.168.1.80:30680)/test?charset=utf8mb4&parseTime=True&loc=Local"
//
////var mysqlConn = "root:Zxcv1234@#@tcp(192.168.0.120:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
//
//func NewDb(conn string) *gorm.DB {
//	if len(conn) == 0 {
//		panic("数据库连接字符未配置")
//	}
//
//	db, err := gorm.Open(mysql.Open(conn), &gorm.Config{})
//	if err != nil {
//		panic("创建mysql 数据库失败")
//	}
//
//	return db.Debug()
//}
//
//func init() {
//	db = NewDb(mysqlConn)
//
//	db.AutoMigrate(&Table1{}, &Table2{}, &Table3{})
//
//	orm.Init(db.Debug())
//}
//
//type Table1 struct {
//	Id        string                `gorm:"column:id;type:varchar(36);primaryKey;not null"` //标识
//	Name      string                `gorm:"column:name;type:varchar(200)" json:"name"`
//	Age       int32                 `gorm:"column:age;type:int" json:"age"`
//	IsDeleted soft_delete.DeletedAt `gorm:"column:is_deleted;softDelete:flag"`
//}
//
//type Table2 struct {
//	Id        string                `gorm:"column:id;type:varchar(36);primaryKey;not null"` //标识
//	Name      string                `gorm:"column:name;type:varchar(200)" json:"name"`
//	Age       int32                 `gorm:"column:age;type:int" json:"age"`
//	IsDeleted soft_delete.DeletedAt `gorm:"column:is_deleted;softDelete:flag"`
//}
//
//type Table3 struct {
//	Id        string                `gorm:"column:id;type:varchar(36);primaryKey;not null"` //标识
//	Name      string                `gorm:"column:name;type:varchar(200)" json:"name"`
//	Age       int32                 `gorm:"column:age;type:int" json:"age"`
//	IsDeleted soft_delete.DeletedAt `gorm:"column:is_deleted;softDelete:flag"`
//}
//
//func (t *Table1) TableName() string {
//	return "Table1"
//}
//
//func (t *Table2) TableName() string {
//	return "Table2"
//}
//
//func (t *Table3) TableName() string {
//	return "Table3"
//}
//
//var table1 = orm.BuildOrmTable[Table1]().Table.T
//var table2 = orm.BuildOrmTable[Table2]().Table.T
//var table3 = orm.BuildOrmTable[Table3]().Table.T
//
//// GetDbContext 获取DbContext。当外部开启事务时，传入开启事务后的db
//func (a *Table1) GetDbContext(ctx context.Context, db ...*gorm.DB) *orm.OrmWrapper[Table1] {
//	return orm.BuildOrmWrapper[Table1](ctx, db...)
//}
//
//func (a *Table2) GetDbContext(ctx context.Context, db ...*gorm.DB) *orm.OrmWrapper[Table2] {
//	return orm.BuildOrmWrapper[Table2](ctx, db...)
//}
//
//func (a *Table3) GetDbContext(ctx context.Context, db ...*gorm.DB) *orm.OrmWrapper[Table3] {
//	return orm.BuildOrmWrapper[Table3](ctx, db...)
//}
//
//func main() {
//
//	table1.GetDbContext(context.Background()).WhereOriginal("1=1").Unscoped().Delete()
//
//	list, err := Create(10)
//	if err != nil {
//		panic(err)
//	}
//
//	if len(list) != 10 {
//		panic("err")
//	}
//
//	Update(list)
//
//	//Select()
//
//	//Where()
//
//	//Join()
//
//}
//
//func Where() {
//
//	//1、通过字段查询
//	model, err := table1.GetDbContext(context.Background()).Where(&table1.Name, orm.Eq, "a").FirstOrDefault()
//	if err != nil {
//		panic(err)
//	}
//	fmt.Println(model)
//	//Sql：SELECT * FROM `Table1` WHERE `name` = 'a' AND `is_deleted` = 0 LIMIT 1
//
//	//2、表别名
//	model, err = table1.GetDbContext(context.Background()).
//		SetTable("t").
//		Where(&table1.Name, orm.Eq, "a", "t").
//		FirstOrDefault()
//	if err != nil {
//		panic(err)
//	}
//	fmt.Println(model)
//	//Sql：SELECT * FROM `Table1` as t WHERE `t`.`name` = 'a' AND `t`.`is_deleted` = 0 LIMIT 1
//
//	//3、字符串字段
//	model, err = table1.GetDbContext(context.Background()).
//		SetTable("t").
//		Where("name", orm.Eq, "a", "t").
//		FirstOrDefault()
//	if err != nil {
//		panic(err)
//	}
//	fmt.Println(model)
//	//Sql：SELECT * FROM `Table1` as t WHERE `t`.`name` = 'a' AND `t`.`is_deleted` = 0 LIMIT 1
//
//	//4、子查询 todo
//
//	//Condition
//	model, err = table1.GetDbContext(context.Background()).WhereCondition(&orm.Condition{
//		TableAlias:     "",           //指定字段的表别名
//		Column:         &table1.Name, //强类型或字符串
//		CompareSymbols: orm.Eq,       //操作比较符
//		Arg:            "a",          //值
//		Func:           "",           //可以对字段包装一个数据库函数
//	}).FirstOrDefault()
//	//Sql：SELECT * FROM `Table1` WHERE `name` = 'a' AND `is_deleted` = 0 LIMIT 1
//
//	//ColumnCondition 常用语exists、子查询等
//
//	//ExistsCondition
//	//第一个条件
//	var cond1 = &orm.Condition{
//		Column:         &table2.Name, //强类型或字符串
//		CompareSymbols: orm.Eq,       //操作比较符
//		Arg:            "name2",      //值
//	}
//
//	//第二个条件
//	var cond2 = &orm.Condition{
//		Column:         &table2.Age, //强类型或字符串
//		CompareSymbols: orm.Gt,      //操作比较符
//		Arg:            18,          //值
//	}
//
//	//第三个条件
//	var cond3 = &orm.ColumnCondition{
//		InnerAlias:     "", //当是join中，则是左边；exists中，则是内表
//		InnerColumn:    &table2.Id,
//		OuterAlias:     "outer", //外部表别名，exists或者join时必须取别名
//		OuterColumn:    &table1.Id,
//		CompareSymbols: orm.Eq,
//	}
//
//	//组合条件，这里三个条件之间是 And
//	var existsConditionBuilder = orm.NewAnd(cond1, cond2, cond3)
//
//	//组装exists条件
//	var existsCondition = &orm.ExistsCondition{
//		Table:            table2,
//		ConditionBuilder: existsConditionBuilder,
//		IsNotExists:      false, //默认 exists，true 就是 not exists
//		Func:             "",
//	}
//
//	//ColumnCondition 常用语exists、子查询等
//	model, err = table1.GetDbContext(context.Background()).
//
//		//设置主表别名，对应 cond3 中的 OuterAlias
//		SetTable("outer").
//		WhereCondition(existsCondition).
//		FirstOrDefault()
//	//Sql：SELECT * FROM	`Table1` AS `outer`
//	//WHERE	(
//	//EXISTS (SELECT 1 FROM `Table2`
//	//		WHERE `is_deleted` = 0 AND ( `name` = 'name2' AND `age` > 18 AND `id` = `outer`.`id` )))
//	//AND `outer`.`is_deleted` = 0 	LIMIT 1
//
//	//ConditionBuilder
//
//	//OriginalCondition gorm 原生条件
//	var cond4 = &orm.OriginalCondition{
//		Sql: "age > ?",
//		Arg: 1,
//	}
//	var cond5 = &orm.OriginalCondition{
//		Sql: "name IN (?)",
//		Arg: []string{"aaa", "bbb"},
//	}
//
//	//组合条件，随意嵌套，无限层级
//	var builder = orm.NewAnd(
//		cond1,
//		cond2,
//		orm.NewOr(cond4, cond5, existsCondition),
//	)
//
//	list, err := table1.GetDbContext(context.Background()).
//
//		//由于使用了 exists，这里必须设置主表表别名
//		SetTable("outer").
//		WhereCondition(builder).
//		ToList()
//	//Sql： SELECT * FROM `Table1` as `outer`
//	// WHERE (
//	//     `name` = 'name2'
//	// AND `age` > 18
//	// AND (
//	//     age > 18
//	//		 OR name IN ('aaa','bbb')
//	//		 OR Exists (SELECT 1 FROM `Table2` WHERE
//	//		                                  `is_deleted` = 0
//	//																			AND (`name` = 'name2' AND `age` > 18 AND `id` = `outer`.`id`))))
//	//AND `outer`.`is_deleted` = 0
//	fmt.Println(list)
//	return
//
//}
//
//func Create(count int) ([]*Table1, error) {
//	if count == 0 {
//		return make([]*Table1, 0), nil
//	}
//
//	var list = make([]*Table1, 0)
//	for i := 0; i < count; i++ {
//		list = append(list, &Table1{
//			Id:   uuid.NewString(),
//			Name: uuid.NewString(),
//			Age:  18,
//		})
//	}
//
//	var err = table1.GetDbContext(context.Background()).Inserts(list)
//	if err != nil {
//		panic(err)
//	}
//
//	return list, nil
//}
//
//func Update(list []*Table1) {
//	//原始更新
//	var m = list[0]
//	m.Name = "111"
//	m.Age = 22
//	//原始更新
//	var s = db.WithContext(context.Background()).Select("name", "age").Model(m).Updates(m)
//	var err = s.Error
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	//包装更新
//	m.Name = ""
//	m.Age = 33
//	c, err := table1.GetDbContext(context.Background()).UpdateOne(m, &table3.Name, &table3.Age)
//	if err != nil {
//		fmt.Println(err)
//	}
//	fmt.Println(c)
//	//sql：UPDATE `Table1` SET `name`='',`age`=33 WHERE `Table1`.`is_deleted` = 0 AND `id` = 'e20c46fe-edc9-43e7-b633-834951809b0c'
//
//	//包装更新
//	m.Name = ""
//	m.Age = 33
//	c, err = table1.GetDbContext(context.Background()).SetTable("a").UpdateOne(m, &table3.Name, &table3.Age)
//	if err != nil {
//		fmt.Println(err)
//	}
//	fmt.Println(c)
//	//sql：UPDATE `Table1` SET `name`='',`age`=33 WHERE `Table1`.`is_deleted` = 0 AND `id` = 'e20c46fe-edc9-43e7-b633-834951809b0c'
//
//	//包装更新 列表
//	for i, t := range list {
//		t.Name = "name1" + strconv.Itoa(i)
//		t.Age = 1001 + int32(i)
//	}
//
//	c, err = table1.GetDbContext(context.Background()).UpdateList(list, &table3.Name, &table3.Age)
//	if err != nil {
//		fmt.Println(err)
//	}
//	fmt.Println(c)
//
//	//外部开启事务更新
//	db.Transaction(func(tx *gorm.DB) error {
//		table1.GetDbContext(context.Background(), tx).UpdateList(list, &table3.Name, &table3.Age)
//		return nil
//	})
//
//	//
//	////字典更新
//	//var columnMap = map[interface{}]interface{}{&table3.Name: "bbb", &table3.Age: gorm.Expr("age+10")}
//	//c, err = table1.GetDbContext(context.Background()).Debug().WhereByColumn(&table3.Id, orm.Eq, m.Id).Update(columnMap)
//	//if err != nil {
//	//	fmt.Println(err)
//	//}
//}
//
//func Select() {
//
//	//select
//	model, err := table1.GetDbContext(context.Background()).Select(&table1.Id, &table1.Name).FirstOrDefault()
//	if err != nil {
//		panic(err)
//	}
//	fmt.Println(model)
//	//sql：SELECT `id`,`name` FROM `Table1` WHERE `is_deleted` = 0 LIMIT 1
//
//	//SelectTable
//	model, err = table1.GetDbContext(context.Background()).
//		SetTable("a").
//		//指定查询 a 表下的字段
//		SelectTable("a", &table1.Id, &table1.Name).
//		FirstOrDefault()
//	if err != nil {
//		panic(err)
//	}
//	fmt.Println(model)
//	//sql：SELECT `a`.`id`,`a`.`name` FROM `Table1` as `a` WHERE `a`.`is_deleted` = 0 LIMIT 1
//
//	//SelectOne
//	model, err = table1.GetDbContext(context.Background()).
//		SelectOne(&table1.Id, "Id_column").
//		SelectOne(&table1.Name, "Name_column").
//		FirstOrDefault()
//	if err != nil {
//		panic(err)
//	}
//	fmt.Println(model)
//	//sql：SELECT `id` as `Id_column`,`name` as `Name_column` FROM `Table1` WHERE `is_deleted` = 0 LIMIT 1
//
//	//SelectOneWithFunc
//	model, err = table1.GetDbContext(context.Background()).
//
//		//给 id 加上函数 Upper
//		SelectOneWithFunc(&table1.Id, "Id_column", orm.Upper).
//		SelectOneWithFunc(&table1.Name, "Name_column", "").
//		FirstOrDefault()
//	if err != nil {
//		panic(err)
//	}
//	fmt.Println(model)
//	//sql：SELECT Upper(`id`) as `Id_column`,`name` as `Name_column` FROM `Table1` WHERE `is_deleted` = 0 LIMIT 1
//
//}
//
//func Join() {
//	model, err := table1.GetDbContext(context.Background()).
//		SetTable("t1").
//		SelectTable("t1", "*").
//		SelectOne(&table2.Age, "Age_t2", "t2").
//		LeftJoin(table2, "t2", &table1.Id, &table2.Id).
//		WhereCondition(&orm.ColumnCondition{
//			InnerAlias:     "t1",
//			InnerColumn:    &table1.Age,
//			OuterAlias:     "t2",
//			OuterColumn:    &table2.Age,
//			CompareSymbols: orm.Gt,
//		}).
//		Distinct().
//		OrderBy(&table1.Name, "t1").
//		OrderByDesc(&table2.Age, "t2").
//		FirstOrDefault()
//	if err != nil {
//		panic(err)
//	}
//	fmt.Println(model)
//	//sql：SELECT DISTINCT
//	//	`t1`.*,
//	//	`t2`.`age` AS `Age_t2`
//	//FROM
//	//	`Table1` AS t1
//	//	LEFT JOIN `Table2` AS t2 ON `t1`.`id` = `t2`.`id`
//	//	AND `t1`.`is_deleted` = 0
//	//	AND `t2`.`is_deleted` = 0
//	//WHERE
//	//	`t1`.`age` > `t2`.`age`
//	//	AND `t1`.`is_deleted` = 0
//	//ORDER BY
//	//	`t1`.`name`,
//	//	`t2`.`age` DESC
//	//	LIMIT 1
//}
