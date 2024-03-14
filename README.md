# gorm 扩展
作为一名.net 开发者，用惯了 ef 和 FreeSql 等orm，对 gorm 中需要手写部分sql不是很适应，所以对gorm进行简单包装，尽量用强类型方便维护。

灵感来自于：https://github.com/acmestack/gorm-plus

支持特性：
- 字段、表 强类型
- 无限层级条件构建
- 简单地连表
- 子查询

### 构建 grom 包装器
```golang
//引入包
import "github.com/ad313/gorm_wrapper/orm"

//初始化包装器，需要传入 *gorm.DB 实例，并指定是何种数据库
orm.Init(db实例, orm.MySql)
```
### where

```golang
//1、通过字段查询
model, err := table1.GetDbContext(context.Background()).WhereByColumn(&table1.Name, orm.Eq, "a").FirstOrDefault()
if err != nil {
	fmt.Println(err)
}
//sql：SELECT * FROM `Table1` WHERE `name` = 'a' AND `deleted_at` = 0 LIMIT 1
```