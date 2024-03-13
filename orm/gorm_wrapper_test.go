package orm

//
//import (
//	"context"
//	"fmt"
//	"github.com/google/uuid"
//	"gorm.io/driver/mysql"
//	"gorm.io/gorm"
//	"gorm.io/plugin/soft_delete"
//	"testing"
//)
//
//type Table3 struct {
//	Id        string                `gorm:"column:id;type:varchar(36);primaryKey;not null"` //标识
//	Name      string                `gorm:"column:name;type:varchar(200)" json:"name"`
//	Age       int32                 `gorm:"column:age;type:int" json:"age"`
//	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;softDelete:flag"`
//}
//
//type Table4 struct {
//	Id        string         `gorm:"column:id;type:varchar(36);primaryKey;not null"` //标识
//	Name      string         `gorm:"column:name;type:varchar(200)" json:"name"`
//	Age       int32          `gorm:"column:age;type:int" json:"age"`
//	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;type:datetime" json:"deleted_at"`
//}
//
//type Table5 struct {
//	Id        string         `gorm:"column:id;type:varchar(36);primaryKey;not null"` //标识
//	Name      string         `gorm:"column:name;type:varchar(200)" json:"name"`
//	Age       int32          `gorm:"column:age;type:int" json:"age"`
//	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;type:datetime" json:"deleted_at"`
//}
//
//func (t *Table3) TableName() string {
//	return "Table3"
//}
//func (t *Table4) TableName() string {
//	return "Table4"
//}
//func (t *Table5) TableName() string {
//	return "Table5"
//}
//
//var db *gorm.DB
//var mysqlConn = "root:123456@tcp(192.168.1.80:30680)/test?charset=utf8mb4&parseTime=True&loc=Local"
//
//func init() {
//	db = NewDb(mysqlConn)
//
//	db.AutoMigrate(&Table3{}, &Table4{}, &Table5{})
//
//	Init(db, MySql)
//}
//
//var table3 = BuildOrmTable[Table3]().Table.T
//var table4 = BuildOrmTable[Table4]().Table.T
//var table5 = BuildOrmTable[Table5]().Table.T
//
//// GetDbContext 获取DbContext。当外部开启事务时，传入开启事务后的db
//func (a *Table3) GetDbContext(ctx context.Context, db ...*gorm.DB) *OrmWrapper[Table3] {
//	return BuildOrmWrapper[Table3](ctx, db...)
//}
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
//func Test_InitData(t *testing.T) {
//	var id1 = uuid.NewString()
//	var id2 = uuid.NewString()
//	var idList = []interface{}{id1, id2}
//	var err error
//
//	err = table3.GetDbContext(context.Background()).Where("1=1").Build().Delete(&Table3{}).Error
//	if err != nil {
//		panic(err)
//	}
//
//	err = _db.Create(&Table3{
//		Id:   id1,
//		Name: id1 + "_Name1",
//		Age:  18,
//	}).Error
//	if err != nil {
//		panic(err)
//	}
//
//	err = _db.Create(&Table3{
//		Id:   id2,
//		Name: id2 + "_Name2",
//		Age:  19,
//	}).Error
//	if err != nil {
//		panic(err)
//	}
//
//	model1, err := table3.GetDbContext(context.Background()).GetById(id1)
//	if err != nil {
//		t.Errorf("GetById faild")
//	}
//	if model1 == nil || model1.Id == "" {
//		t.Errorf("GetById faild")
//	}
//
//	modelList, err := table3.GetDbContext(context.Background()).GetByIds(idList)
//	if err != nil {
//		t.Errorf("GetById faild")
//	}
//	if len(modelList) == 0 {
//		t.Errorf("GetById faild")
//	}
//
//	_, err = table3.GetDbContext(context.Background()).Delete()
//	if err == nil {
//		t.Errorf("Delete faild")
//	}
//
//	////根据条件删除
//	//count, err := table3.GetDbContext(context.Background()).
//	//	WhereCondition(&Condition{Column: &table3.Age, CompareSymbols: Gt, Arg: 1}).
//	//	Delete()
//	//if err != nil {
//	//	t.Errorf("Delete faild")
//	//}
//	//if count <= 0 {
//	//	t.Errorf("Delete faild")
//	//}
//
//	////删除
//	//err = table3.GetDbContext(context.Background()).DeleteById()
//	//if err == nil {
//	//	t.Errorf("DeleteById faild")
//	//}
//	//
//	//err = table3.GetDbContext(context.Background()).DeleteById(idList...)
//	//if err != nil {
//	//	t.Errorf("DeleteById faild")
//	//}
//
//}
//
//func Test_Wrapper(t *testing.T) {
//
//	var dbContext = table3.GetDbContext(context.Background()).WhereCondition(&Condition{Column: &table3.Age, CompareSymbols: Gt, Arg: 1})
//
//	fmt.Println(dbContext.ToSql())
//
//	list, err := dbContext.ToList()
//	if err != nil {
//		panic(err)
//	}
//	if len(list) == 0 {
//		t.Errorf("ToList faild")
//	}
//
//	//having
//	dbContext = table3.GetDbContext(context.Background()).WhereCondition(&Condition{
//		TableAlias:     "",
//		Column:         &table3.Age,
//		CompareSymbols: Gt,
//		Arg:            1,
//	}).Select(&table3.Name).
//		SelectWithFunc("1", "aaa", Count).
//		GroupBy(&table3.Name).
//		Having(&Condition{
//			Column:         "1",
//			CompareSymbols: Gt,
//			Arg:            0,
//			Func:           "Count",
//		})
//
//	fmt.Println(dbContext.ToSql())
//
//	//join
//	dbContext = table3.GetDbContext(context.Background()).WhereCondition(&Condition{
//		TableAlias:     "b",
//		Column:         &table3.Age,
//		CompareSymbols: Gt,
//		Arg:            1,
//	}).InnerJoin(table3, "b", &table3.Id, &table3.Id)
//
//	fmt.Println(dbContext.ToSql())
//
//	//join 衍生表
//	var childTable = table3.GetDbContext(context.Background()).WhereCondition(&Condition{
//		//TableAlias:     "b",
//		Column:         &table3.Age,
//		CompareSymbols: Gt,
//		Arg:            1,
//	}).BuildForQuery()
//
//	sql, err := table3.GetDbContext(context.Background()).JoinChildTable(childTable, "b", &table3.Id, &table3.Id, LeftJoin).ToSql()
//	if err != nil {
//		panic(err)
//	}
//
//	fmt.Println(sql)
//	fmt.Println("-----------------------")
//
//	//查询衍生表
//	sql, err = table3.GetDbContext(context.Background()).SetTable("c", childTable).ToSql()
//	fmt.Println(sql)
//
//	//字段子查询
//	var db1 = table3.GetDbContext(context.Background()).Select(&table3.Id).Limit(1).BuildForQuery()
//	var db2 = table3.GetDbContext(context.Background()).WhereCondition(&Condition{
//		Column:         &table3.Id,
//		CompareSymbols: Eq,
//		Arg:            db1,
//	})
//	sql, err = db2.ToSql()
//	fmt.Println(sql)
//
//	//sql, err := table3.GetDbContext(context.Background()).
//	//	SetTableAlias("a").
//	//	//SelectColumnOriginal("a.*", "").
//	//	//SelectColumnOriginal("count(*)", "aaa").
//	//	WhereCondition(&Condition{
//	//		Column:         &table3.Id,
//	//		CompareSymbols: Eq,
//	//		Arg:            "123",
//	//		TableAlias:     "a",
//	//	}).
//	//	WhereByColumn(&table3.Name, Like, "abc", "a").
//	//	WhereByColumn(&table3.Name, StartWith, "abc", "b").
//	//	WhereByColumn(&table3.Name, EndWith, "abc", "c").
//	//	WhereCondition(&ExistsCondition{
//	//		Table: table3,
//	//		ConditionBuilder: NewAndConditionBuilder(&TableCondition{
//	//			InnerColumn:    &table3.Id,
//	//			OuterAlias:     "a",
//	//			OuterColumn:    &table3.Id,
//	//			CompareSymbols: Eq,
//	//		}, &TableCondition{
//	//			InnerColumn:    &table3.Name,
//	//			OuterAlias:     "a",
//	//			OuterColumn:    &table3.Name,
//	//			CompareSymbols: NotEq,
//	//		}, &Condition{
//	//			TableAlias:     "",
//	//			Column:         &table3.Age,
//	//			CompareSymbols: Gt,
//	//			Arg:            18,
//	//		}),
//	//		IsNotExists: true,
//	//		Func:        "",
//	//	}).
//	//	LeftJoin(table4, "b", &table3.Id, &table4.Id).
//	//	LeftJoin(table5, "c", &table4.Name, &table5.Name).
//	//	//Unscoped().
//	//	ToSql()
//	//
//	//if err != nil {
//	//	fmt.Println(err)
//	//}
//	//
//	//fmt.Println(sql)
//	//if err != nil {
//	//	t.Errorf("Test_Wrapper faild")
//	//}
//	//
//	//fmt.Println(sql)
//
//	//table3.GetDbContext(context.Background()).
//	//	LeftJoin(table4, "b", &table3.Id, &table4.Id).
//	//	ToPagerList(&Pager{Page: 0, PageSize: 0})
//
//}
