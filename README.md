# gorm 扩展
作为一名.net 开发者，用惯了 ef 和 FreeSql 等orm，对 gorm 中需要手写部分sql不是很适应，所以对gorm进行简单包装，尽量用强类型方便维护。

灵感来自于：https://github.com/acmestack/gorm-plus

[//]: # (* [入门]&#40;a.md&#41;)

支持特性：
- 字段、表 强类型
- 无限层级条件构建
- 简单地连表
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

//初始化包装器，需要传入 *gorm.DB 实例
orm.Init(db实例)
```
### 构建实体
``` golang
type Table1 struct {
	Id        string                `gorm:"column:id;type:varchar(36);primaryKey;not null"` //标识
	Name      string                `gorm:"column:name;type:varchar(200)" json:"name"`
	Age       int32                 `gorm:"column:age;type:int" json:"age"`
	IsDeleted soft_delete.DeletedAt `gorm:"column:is_deleted;softDelete:flag"`
}

// GetDbContext 获取DbContext。当外部开启事务时，传入开启事务后的db
func (a *Table1) GetDbContext(ctx context.Context, db ...*gorm.DB) *orm.OrmWrapper[Table1] {
	return orm.BuildOrmWrapper[Table1](ctx, db...)
}

// Table1表对应的操作实体，每个表对应一个实例
var table1 = orm.BuildOrmTable[Table1]().Table.T
```



### where

```golang
//1、通过字段查询
model, err := table1.GetDbContext(context.Background()).WhereByColumn(&table1.Name, orm.Eq, "a").FirstOrDefault()
if err != nil {
	fmt.Println(err)
}
//sql：SELECT * FROM `Table1` WHERE `name` = 'a' AND `is_deleted` = 0 LIMIT 1
```