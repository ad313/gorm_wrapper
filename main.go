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
//)
//
//var db *gorm.DB
//
////var mysqlConn = "root:123456@tcp(192.168.1.80:30680)/test?charset=utf8mb4&parseTime=True&loc=Local"
//
//var mysqlConn = "root:Zxcv1234@#@tcp(192.168.0.120:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
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
//	return db
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
//	//list, err := Create(10)
//	//if err != nil {
//	//	panic(err)
//	//}
//	//
//	//if len(list) != 10 {
//	//	panic("err")
//	//}
//	//
//	//Update(list)
//
//	Select()
//
//	//var model = &Table3{}
//	////var err = db.Model(&Table3{}).FirstOrInit(&model).Error
//	////if err != nil {
//	////	panic(err)
//	////}
//	////
//	////fmt.Println(model)
//	//
//	////子表
//	//var db1 = db.Table("Table31 as a")
//	//
//	//var db2 = db.Table("(?) as b", db1)
//	//
//	//var err = db2.FirstOrInit(&model).Error
//	//if err != nil {
//	//	panic(err)
//	//}
//
//	//var db1 = table3.GetDbContext(context.Background()).SetTable("a").BuildForQuery()
//	//
//	//var db2 = table3.GetDbContext(context.Background()).SetTable("b", db1)
//	//
//	//model, err := db2.FirstOrDefault()
//	//if err != nil {
//	//	fmt.Println(err)
//	//}
//	//
//	//fmt.Println(model)
//
//	////字段子查询
//	//var db1 = table3.GetDbContext(context.Background()).Select(&table3.Id).Limit(1).BuildForQuery()
//	//var db2 = table3.GetDbContext(context.Background()).OrmCondition(&orm.Condition{
//	//	Column:         &table3.Id,
//	//	CompareSymbols: orm.Eq,
//	//	Arg:            db1,
//	//})
//	//list, err := db2.ToList()
//	//if err != nil {
//	//	fmt.Println(err)
//	//}
//	//fmt.Println(list)
//	//
//	//list, err = table3.GetDbContext(context.Background()).ToList()
//	//if len(list) == 0 {
//	//	return
//	//}
//	//
//
//	////1、通过字段查询
//	//model, err := table1.GetDbContext(context.Background()).WhereByColumn(&table1.Name, orm.Eq, "a").FirstOrDefault()
//	//if err != nil {
//	//	fmt.Println(err)
//	//}
//	//fmt.Println(model)
//	////sql：SELECT * FROM `Table1` WHERE `name` = 'a' AND `deleted_at` = 0 LIMIT 1
//	//
//	////2、通过条件模型查询
//	//model, err = table1.GetDbContext(context.Background()).WhereCondition(&orm.Condition{Column: &table1.Name, CompareSymbols: orm.Eq, Arg: "a"}).FirstOrDefault()
//	//if err != nil {
//	//	fmt.Println(err)
//	//}
//	//fmt.Println(model)
//	////sql：SELECT * FROM `Table1` WHERE `name` = 'a' AND `deleted_at` = 0 LIMIT 1
//	//
//	////3、条件组合 and
//	//var cond = orm.NewAnd(
//	//	&orm.Condition{Column: &table3.Name, CompareSymbols: orm.Eq, Arg: "a"},
//	//	&orm.Condition{Column: &table3.Age, CompareSymbols: orm.Gt, Arg: 18})
//	//list, err := table3.GetDbContext(context.Background()).WhereCondition(cond).ToList()
//	//if err != nil {
//	//	fmt.Println(err)
//	//}
//	//fmt.Println(list)
//	////sql：SELECT * FROM `Table1` WHERE ((`name` = 'a' AND `age` > 18)) AND `deleted_at` = 0
//	//
//	////4、条件组合 or
//	//var cond2 = orm.NewOr(
//	//	&orm.Condition{Column: &table3.Name, CompareSymbols: orm.Eq, Arg: "a"},
//	//	&orm.Condition{Column: &table3.Age, CompareSymbols: orm.Gt, Arg: 18})
//	//list, err = table3.GetDbContext(context.Background()).WhereCondition(cond2).ToList()
//	//if err != nil {
//	//	fmt.Println(err)
//	//}
//	//fmt.Println(list)
//	////sql：SELECT * FROM `Table1` WHERE ((`name` = 'a' OR `age` > 18)) AND `deleted_at` = 0
//	//
//	////5、条件与条件组合
//	//var cond3 = orm.NewAnd(cond, cond2)
//	//list, err = table3.GetDbContext(context.Background()).WhereCondition(cond3).ToList()
//	//if err != nil {
//	//	fmt.Println(err)
//	//}
//	//fmt.Println(list)
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
//	////原始更新
//	//var s = db.WithContext(context.Background()).Select("name", "age").Model(m).Updates(m)
//	//var err = s.Error
//	//if err != nil {
//	//	fmt.Println(err)
//	//}
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
//	////包装更新 列表
//	//for i, t := range list {
//	//	t.Name = "name1" + strconv.Itoa(i)
//	//	t.Age = 1001 + int32(i)
//	//}
//	//
//	//c, err = table1.GetDbContext(context.Background()).UpdateList(list, &table3.Name, &table3.Age)
//	//if err != nil {
//	//	fmt.Println(err)
//	//}
//	//fmt.Println(c)
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
//	var errr = db.Debug().Table("Table1 as a").First(new(Table1)).Error
//	fmt.Println(errr)
//	sql, err := table1.GetDbContext(context.Background()).SetTable("a").Select(&table1.Id).ToFirstSql()
//	if err != nil {
//		panic(err)
//	}
//	fmt.Println(sql)
//	//sql：SELECT `id` FROM `Table1` as a WHERE `a`.`is_deleted` = 0 ORDER BY `a`.`id` LIMIT 1
//
//	var tmp = table1.GetDbContext(context.Background()).Select(&table1.Id).WhereCondition(&orm.ColumnCondition{
//		InnerColumn:    &table1.Name,
//		OuterAlias:     "a",
//		OuterColumn:    &table1.Name,
//		CompareSymbols: orm.Eq,
//	})
//	sql, err = table1.GetDbContext(context.Background()).
//		SetTable("a").
//		Select(&table1.Id).
//		SelectOne(tmp.BuildForQuery(), "childrenSelect").
//		ToFirstSql()
//	if err != nil {
//		panic(err)
//	}
//	fmt.Println(sql)
//	//SELECT `id`,
//	//(SELECT `id` FROM `Table1` WHERE `name` = `a`.`name` AND `is_deleted` = 0 ORDER BY `Table1`.`id` LIMIT 1) as `childrenSelect`
//	//FROM `Table1` as a WHERE `a`.`is_deleted` = 0 ORDER BY `a`.`id` LIMIT 1
//
//	table1.GetDbContext(context.Background()).Select(&table1.Id).FirstOrDefault()
//	//sql：SELECT `id` FROM `Table1` WHERE `is_deleted` = 0 LIMIT 1
//
//	table1.GetDbContext(context.Background()).Select(&table1.Id, &table1.Name).FirstOrDefault()
//	//sql：SELECT `id`,`name` FROM `Table1` WHERE `is_deleted` = 0 LIMIT 1
//
//	table1.GetDbContext(context.Background()).
//		Select(&table1.Id, &table1.Name).
//		SelectTable("", &table1.Age).
//		FirstOrDefault()
//	//sql：SELECT `id`,`name`,`age` FROM `Table1` WHERE `is_deleted` = 0 LIMIT 1
//}
