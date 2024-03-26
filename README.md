# gorm 扩展
作为一名.net 开发者，用惯了 ef 和 FreeSql 等orm，对 gorm 中需要手写部分sql不是很适应，所以对gorm进行简单包装，尽量用强类型方便维护。

灵感来自于 gorm-plus：https://github.com/acmestack/gorm-plus

[//]: # (* [入门]&#40;a.md&#41;)

支持特性：
- 字段、表 强类型
- 无限层级条件构建
- 简单地连表查询
- 子查询

### 建表
``` mysql
CREATE TABLE `Table1`  (
  `id` varchar(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `age` int(0) NULL DEFAULT NULL,
  `is_deleted` bigint(0) UNSIGNED NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE
);
```
### 构建 gorm_wrapper
```golang
//安装包
go get github.com/ad313/gorm_wrapper

//引入包
import "github.com/ad313/gorm_wrapper/orm"

//初始化包装器，需要传入 *gorm.DB 实例。这里不创建实例，由外部传入
orm.Init(yourdb)
```
### 构建实体
``` golang
type Table1 struct {
	Id        string                `gorm:"column:id;type:varchar(36);primaryKey;not null"` //标识
	Name      string                `gorm:"column:name;type:varchar(200)" json:"name"`
	Age       int32                 `gorm:"column:age;type:int" json:"age"`
	IsDeleted soft_delete.DeletedAt `gorm:"column:is_deleted;softDelete:flag"`
}

func (t *Table1) TableName() string {
	return "Table1"
}

// GetDbContext 获取DbContext。当外部开启事务时，传入开启事务后的db
func (a *Table1) GetDbContext(ctx context.Context, db ...*gorm.DB) *orm.OrmWrapper[Table1] {
	return orm.BuildOrmWrapper[Table1](ctx, db...)
}

// Table1表对应的操作实体，每个表对应一个实例
var table1 = orm.BuildOrmTable[Table1]().Table.T
```

## 查询 Where
```
支持的操作符，也可以传字符串
* orm.Eq         // 等于
* orm.NotEq      // 不等于
* orm.Gt         // 大于
* orm.GtAndEq    // 大于等于
* orm.Less       // 小于
* orm.LessAndEq  // 小于等于
* orm.In         // IN (?)
* orm.NotIn      // NOT IN (?)
* orm.Like       // Like "%a%"
* orm.NotLike    // NOT Like "%a%"
* orm.StartWith  // Like "a%"
* orm.EndWith    // Like "%a"
* orm.IsNull     // IS NULL
* orm.IsNotNull  // IS NOT NULL
```

### 1、.Where（字段查询） 参数：字段名（强类型或字符串）、操作符、参数值（可以是子查询）、指定字段的表别名，可不传

```golang
//1、通过字段查询
model, err := table1.GetDbContext(context.Background()).	
    //强类型字段
    Where(&table1.Name, orm.Eq, "a").	
    FirstOrDefault()
//Sql：SELECT * FROM `Table1` WHERE `name` = 'a' AND `is_deleted` = 0 LIMIT 1 //默认会加上软删除

//2、表别名
model, err = table1.GetDbContext(context.Background()).	
    //表别名
    SetTable("t").
    Where(&table1.Name, orm.Eq, "a", "t").
    FirstOrDefault()
//Sql：SELECT * FROM `Table1` as t WHERE `t`.`name` = 'a' AND `t`.`is_deleted` = 0 LIMIT 1

//3、字符串字段
model, err = table1.GetDbContext(context.Background()).
    SetTable("t").	
    //字符串字段
    Where("name", orm.Eq, "a", "t").
    FirstOrDefault()
//Sql：SELECT * FROM `Table1` as t WHERE `t`.`name` = 'a' AND `t`.`is_deleted` = 0 LIMIT 1
```

### 2、.WhereCondition 传入条件模型
```
共支持 5 种条件模型，都继承 OrmCondition
* Condition         //字段与值比较
* ColumnCondition   //表与表之间的字段比较
* ExistsCondition   //Exists语句
* ConditionBuilder  //条件构造器，可以无限层级构建条件
* OriginalCondition //gorm原始条件
```

#### 2.1 Condition
```golang
model, err = table1.GetDbContext(context.Background()).WhereCondition(&orm.Condition{
    TableAlias:     "",           //指定字段的表别名
    Column:         &table1.Name, //强类型或字符串
    CompareSymbols: orm.Eq,       //操作比较符
    Arg:            "a",          //值
    Func:           "",           //可以对字段包装一个数据库函数
}).FirstOrDefault()
//Sql：SELECT * FROM `Table1` WHERE `name` = 'a' AND `is_deleted` = 0 LIMIT 1
```

#### 2.2 ColumnCondition 常用语exists、子查询等，参考下面

#### 2.3 ExistsCondition
```golang
//第一个条件
var cond1 = &orm.Condition{
    Column:         &table2.Name, //强类型或字符串
    CompareSymbols: orm.Eq,       //操作比较符
    Arg:            "name2",      //值
}

//第二个条件
var cond2 = &orm.Condition{
    Column:         &table2.Age, //强类型或字符串
    CompareSymbols: orm.Gt,      //操作比较符
    Arg:            18,          //值
}

//第三个条件
var cond3 = &orm.ColumnCondition{
    InnerAlias:     "", //如果是在join中，则是左边表；如果是exists，则是内部表
    InnerColumn:    &table2.Id,
    OuterAlias:     "outer", //外部表别名，exists或者join时必须取别名
    OuterColumn:    &table1.Id,
    CompareSymbols: orm.Eq,
}

//组合条件，这里三个条件之间是 And
var existsConditionBuilder = orm.NewAnd(cond1, cond2, cond3)

//组装exists条件
var existsCondition = &orm.ExistsCondition{
    Table:            table2,
    ConditionBuilder: existsConditionBuilder,
    IsNotExists:      false, //默认 exists，true 就是 not exists
}

//ColumnCondition 常用语exists、子查询等
model, err = table1.GetDbContext(context.Background()).

    //设置主表别名，对应 cond3 中的 OuterAlias
    SetTable("outer").
    WhereCondition(existsCondition).
    FirstOrDefault()
//Sql：SELECT * FROM `Table1` AS `outer`
//WHERE	(
//EXISTS (SELECT 1 FROM `Table2`
//		WHERE `is_deleted` = 0 AND ( `name` = 'name2' AND `age` > 18 AND `id` = `outer`.`id` )))
//AND `outer`.`is_deleted` = 0 	LIMIT 1
```

#### 2.4 ConditionBuilder
```golang
//OriginalCondition gorm 原生条件
var cond4 = &orm.OriginalCondition{
    Sql: "age > ?",
    Arg: 1,
}
var cond5 = &orm.OriginalCondition{
    Sql: "name IN (?)",
    Arg: []string{"aaa", "bbb"},
}

//组合条件，随意嵌套，无限层级
var builder = orm.NewAnd(
    cond1,
    cond2,
    orm.NewOr(cond4, cond5, existsCondition),
)

//执行
list, err := table1.GetDbContext(context.Background()).

    //由于使用了 exists，这里必须设置主表表别名
    SetTable("outer").
    WhereCondition(builder).
    ToList()
//Sql： SELECT * FROM `Table1` as `outer`
// WHERE (
//     `name` = 'name2'
// AND `age` > 18
// AND (
//      age > 18
//      OR name IN ('aaa','bbb')
//      OR Exists (SELECT 1 FROM `Table2` WHERE
//                  `is_deleted` = 0
//                  AND (`name` = 'name2' AND `age` > 18 AND `id` = `outer`.`id`))))
//AND `outer`.`is_deleted` = 0
```
#### 2.5 OriginalCondition，参照 2.4

### 3、WhereOriginal，和 gorm 一样，这里省略

## Select
### .Select 同时传入多个字段，无法指定字段的表名
```golang
//select
model, err := table1.GetDbContext(context.Background()).Select(&table1.Id, &table1.Name).FirstOrDefault()
//sql：SELECT `id`,`name` FROM `Table1` WHERE `is_deleted` = 0 LIMIT 1
```
### .SelectTable 指定查询某个表下的多个字段，如果不指定表别名，则和 Select 一样
```golang
model, err = table1.GetDbContext(context.Background()).
    SetTable("a").
    //指定查询 a 表下的字段
    SelectTable("a", &table1.Id, &table1.Name).
    FirstOrDefault()
//sql：SELECT `a`.`id`,`a`.`name` FROM `Table1` as `a` WHERE `a`.`is_deleted` = 0 LIMIT 1
```
### .SelectOne 查询单个字段，可以指定字段别名和表别名
```golang
model, err = table1.GetDbContext(context.Background()).
    SelectOne(&table1.Id, "Id_column").
    SelectOne(&table1.Name, "Name_column").
    FirstOrDefault()
//sql：SELECT `id` as `Id_column`,`name` as `Name_column` FROM `Table1` WHERE `is_deleted` = 0 LIMIT 1
```
### .SelectOneWithFunc 查询单个字段，SelectOne 的基础上给字段加上数据库函数（你需要判断当前数据库是否支持这个函数）
```
目前支持的函数
* Max
* Min
* Avg
* Sum
* Count
* Abs
* Sqrt
* Ceil
* Floor
* Round
* Upper
* Lower
```
```golang
model, err = table1.GetDbContext(context.Background()).
    //给 id 加上函数 Upper
    SelectOneWithFunc(&table1.Id, "Id_column", orm.Upper).
    SelectOneWithFunc(&table1.Name, "Name_column", "").
    FirstOrDefault()
//sql：SELECT Upper(`id`) as `Id_column`,`name` as `Name_column` FROM `Table1` WHERE `is_deleted` = 0 LIMIT 1
```

## Join（left join、right join、inner join、outer join）
### .LeftJoin
```golang
	model, err := table1.GetDbContext(context.Background()).
		SetTable("t1").
		SelectTable("t1", "*").
		SelectOne(&table2.Age, "Age_t2", "t2").
		LeftJoin(table2, "t2", &table1.Id, &table2.Id).
		WhereCondition(&orm.ColumnCondition{
			InnerAlias:     "t1",
			InnerColumn:    &table1.Age,
			OuterAlias:     "t2",
			OuterColumn:    &table2.Age,
			CompareSymbols: orm.Gt,
		}).
		Distinct().
		OrderBy(&table1.Name, "t1").
		OrderByDesc(&table2.Age, "t2").
		FirstOrDefault()
	if err != nil {
		panic(err)
	}
	fmt.Println(model)
	//sql：SELECT DISTINCT
	//	`t1`.*,
	//	`t2`.`age` AS `Age_t2`
	//FROM
	//	`Table1` AS t1
	//	LEFT JOIN `Table2` AS t2 ON `t1`.`id` = `t2`.`id`
	//	AND `t1`.`is_deleted` = 0
	//	AND `t2`.`is_deleted` = 0
	//WHERE
	//	`t1`.`age` > `t2`.`age`
	//	AND `t1`.`is_deleted` = 0
	//ORDER BY
	//	`t1`.`name`,
	//	`t2`.`age` DESC
	//	LIMIT 1
```
## OrderBy 正序、OrderByDesc 倒序

## GroupBy

## Having

## Limit、Offset

## Unscoped 和gorm一样，忽略软删除字段

## Debug 控制台打印sql

## ToSql 返回sql；ToFirstOrDefaultSql 加了 limit 1

## 子查询
* Select 子句
* Table 子句
* Where 子句
