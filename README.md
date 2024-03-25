# gorm 扩展
作为一名.net 开发者，用惯了 ef 和 FreeSql 等orm，对 gorm 中需要手写部分sql不是很适应，所以对gorm进行简单包装，尽量用强类型方便维护。

灵感来自于：https://github.com/acmestack/gorm-plus

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
orm.Init(你的db实例)
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

### 1、Where 字段：字段名（强类型或字符串）、操作符、参数值（可以是子查询）、指定字段的表别名，可不传

```golang
//1、通过字段查询
model, err := table1.GetDbContext(context.Background()).Where(&table1.Name, orm.Eq, "a").FirstOrDefault()
if err != nil {
    fmt.Println(err)
}
fmt.Println(model)
//Sql：SELECT * FROM `Table1` WHERE `name` = 'a' AND `is_deleted` = 0 LIMIT 1 //默认会加上软删除

//2、表别名
model, err = table1.GetDbContext(context.Background()).
    SetTable("t").
    Where(&table1.Name, orm.Eq, "a", "t").
    FirstOrDefault()
if err != nil {
    panic(err)
}
fmt.Println(model)
//Sql：SELECT * FROM `Table1` as t WHERE `t`.`name` = 'a' AND `t`.`is_deleted` = 0 LIMIT 1

//3、字符串字段
model, err = table1.GetDbContext(context.Background()).
    SetTable("t").
    Where("name", orm.Eq, "a", "t").
    FirstOrDefault()
if err != nil {
    panic(err)
}
fmt.Println(model)
//Sql：SELECT * FROM `Table1` as t WHERE `t`.`name` = 'a' AND `t`.`is_deleted` = 0 LIMIT 1

```

### 2、WhereCondition 传入 condition 模型
```
共支持 5 种条件模型
* Condition         //字段与值比较
* ColumnCondition   //表与表之间比较
* ExistsCondition   //Exists
* OriginalCondition //gorm原始条件
* ConditionBuilder  //条件构造器，可以无限层级构建条件
```

```golang

```