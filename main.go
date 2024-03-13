package main

//
//import (
//	"context"
//	"fmt"
//	"github.com/ad313/gorm_wrapper/orm"
//	"gorm.io/driver/mysql"
//	"gorm.io/gorm"
//	"gorm.io/plugin/soft_delete"
//	"strconv"
//)
//
//var db *gorm.DB
//var mysqlConn = "root:123456@tcp(192.168.1.80:30680)/test?charset=utf8mb4&parseTime=True&loc=Local"
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
//	return db
//}
//
//func init() {
//	db = NewDb(mysqlConn)
//
//	orm.Init(db, orm.MySql)
//}
//
//type Table3 struct {
//	Id        string                `gorm:"column:id;type:varchar(36);primaryKey;not null"` //标识
//	Name      string                `gorm:"column:name;type:varchar(200)" json:"name"`
//	Age       int32                 `gorm:"column:age;type:int" json:"age"`
//	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;softDelete:flag"`
//}
//
//func (t *Table3) TableName() string {
//	return "Table3"
//}
//
//var table3 = orm.BuildOrmTable[Table3]().Table.T
//
//// GetDbContext 获取DbContext。当外部开启事务时，传入开启事务后的db
//func (a *Table3) GetDbContext(ctx context.Context, db ...*gorm.DB) *orm.OrmWrapper[Table3] {
//	return orm.BuildOrmWrapper[Table3](ctx, db...)
//}
//
//func main() {
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
//	//字段子查询
//	var db1 = table3.GetDbContext(context.Background()).Select(&table3.Id).Limit(1).BuildForQuery()
//	var db2 = table3.GetDbContext(context.Background()).WhereCondition(&orm.Condition{
//		Column:         &table3.Id,
//		CompareSymbols: orm.Eq,
//		Arg:            db1,
//	})
//	list, err := db2.ToList()
//	if err != nil {
//		fmt.Println(err)
//	}
//	fmt.Println(list)
//
//	list, err = table3.GetDbContext(context.Background()).ToList()
//	if len(list) == 0 {
//		return
//	}
//
//	//原始更新
//	var m = list[0]
//	m.Name = "111"
//	m.Age = 22
//	//原始更新
//	var s = db.WithContext(context.Background()).Select("name", "age").Model(m).Updates(m)
//	err = s.Error
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	//包装更新
//	m.Name = ""
//	m.Age = 33
//	c, err := table3.GetDbContext(context.Background(), db).UpdateOne(m, &table3.Name, &table3.Age)
//	if err != nil {
//		fmt.Println(err)
//	}
//	fmt.Println(c)
//
//	//包装更新 列表
//	for i, t := range list {
//		t.Name = "name1" + strconv.Itoa(i)
//		t.Age = 1001 + int32(i)
//	}
//
//	c, err = table3.GetDbContext(context.Background()).UpdateList(list, &table3.Name, &table3.Age)
//	if err != nil {
//		fmt.Println(err)
//	}
//	fmt.Println(c)
//
//	//字典更新
//	var columnMap = map[interface{}]interface{}{&table3.Name: "bbb", &table3.Age: gorm.Expr("age+10")}
//	c, err = table3.GetDbContext(context.Background()).Debug().WhereByColumn(&table3.Id, orm.Eq, m.Id).Update(columnMap)
//	if err != nil {
//		fmt.Println(err)
//	}
//}
