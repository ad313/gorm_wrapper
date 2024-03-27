package orm

import (
	"errors"
	"gorm.io/gorm"
	"strings"
)

// Count 查询总条数
func (o *OrmWrapper[T]) Count() (int64, error) {

	//Build sql
	o.BuildForQuery()

	//创建语句过程中的错误
	if o.Error != nil {
		return 0, o.Error
	}

	var err error
	var total int64

	//left join 加上 distinct
	if len(o.builder.joinModels) > 0 {
		err = o.GetNewDb().Table("(?) as leftJoinTableWrapper", o.builder.DbContext).Count(&total).Error
	} else {
		err = o.builder.DbContext.Count(&total).Error
	}
	if err != nil {
		return 0, err
	}

	return total, nil
}

// GetById 通过id获取数据
func (o *OrmWrapper[T]) GetById(id interface{}) (*T, error) {
	if id == nil {
		o.Error = errors.New("id 不能为空")
		return nil, o.Error
	}

	key, err := o.builder.getPrimaryKey()
	if err != nil {
		return nil, err
	}

	sql, _, err := (&Condition{
		Column:         key,
		CompareSymbols: Eq,
		Arg:            id,
	}).BuildSql(_dbType)
	if err != nil {
		return nil, err
	}

	//执行完毕清理痕迹
	defer o.clearBuilder()

	var result = new(T)
	err = o.GetNewDb().Model(new(T)).Where(sql, id).Take(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return result, err
}

// GetByIds 通过id获取数据
func (o *OrmWrapper[T]) GetByIds(idList []interface{}) ([]*T, error) {
	if len(idList) == 0 {
		o.Error = errors.New("idList 不能为空")
		return nil, o.Error
	}

	key, err := o.builder.getPrimaryKey()
	if err != nil {
		return nil, err
	}

	sql, _, err := (&Condition{
		Column:         key,
		CompareSymbols: In,
		Arg:            idList,
	}).BuildSql(_dbType)
	if err != nil {
		return nil, err
	}

	//执行完毕清理痕迹
	defer o.clearBuilder()

	var result = make([]*T, 0)
	return result, o.GetNewDb().Model(new(T)).Where(sql, idList).Scan(&result).Error
}

// FirstOrDefault 返回第一条，没命中返回nil，可以传入自定义scan，自定义接收数据
func (o *OrmWrapper[T]) FirstOrDefault(scan ...func(db *gorm.DB) error) (*T, error) {

	//Build sql
	o.BuildForQuery()

	//创建语句过程中的错误
	if o.Error != nil {
		return nil, o.Error
	}

	//执行完毕清理痕迹
	defer o.clearBuilder()

	var err error
	var result *T
	if len(scan) > 0 {
		if scan[0] == nil {
			return nil, errors.New("scan 函数不能为空")
		}
		err = scan[0](o.builder.DbContext)
	} else {
		//First 会自动添加主键排序
		err = o.builder.DbContext.Limit(1).Take(&result).Error
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return result, nil
}

// ToList 返回列表，可以传入自定义scan，自定义接收数据
func (o *OrmWrapper[T]) ToList(scan ...func(db *gorm.DB) error) ([]*T, error) {

	//Build sql
	o.BuildForQuery()

	//创建语句过程中的错误
	if o.Error != nil {
		return nil, o.Error
	}

	if len(scan) > 0 {
		if scan[0] == nil {
			return nil, errors.New("scan 函数不能为空")
		}
		return nil, scan[0](o.builder.DbContext)
	}

	//执行完毕清理痕迹
	defer o.clearBuilder()

	var list = make([]*T, 0)
	err := o.builder.DbContext.Scan(&list).Error
	if err != nil {
		return nil, err
	}

	return list, nil
}

// ToPagerList 分页查询，可以自定义scan，否则返回当前实体分页结果
func (o *OrmWrapper[T]) ToPagerList(pager *Pager, scan ...func(db *gorm.DB) error) (*PagerList[T], error) {

	//Build sql
	o.BuildForQuery()

	//创建语句过程中的错误
	if o.Error != nil {
		return nil, o.Error
	}

	if pager == nil {
		return nil, errors.New("传入分页数据不能为空")
	}

	//执行完毕清理痕迹
	defer o.clearBuilder()

	//包含空格 asc desc
	if strings.Contains(pager.Order, " ") {
		var arr = strings.Split(pager.Order, " ")
		if strings.ToUpper(arr[1]) == "DESC" {
			o.OrderByDesc(arr[0])
		} else {
			o.OrderBy(arr[0])
		}
	} else if pager.Order != "" {
		o.OrderBy(pager.Order)
	}

	if pager.Page <= 0 {
		pager.Page = 1
	}

	if pager.PageSize <= 0 {
		pager.PageSize = 20
	}

	//总条数
	var total int64
	var err error

	//left join 加上 distinct
	if len(o.builder.joinModels) > 0 {
		err = o.builder._DbContext.Table("(?) as leftJoinTableWrapper", o.builder.DbContext).Count(&total).Error
	} else {
		err = o.builder.DbContext.Count(&total).Error
	}
	if err != nil {
		return nil, err
	}

	var result = &PagerList[T]{
		Page:       pager.Page,
		PageSize:   pager.PageSize,
		TotalCount: int32(total),
		Order:      pager.Order,
	}

	if result.TotalCount == 0 {
		return result, nil
	}

	if len(scan) == 0 {
		err = o.builder.DbContext.Offset(int((pager.Page - 1) * pager.PageSize)).Limit(int(pager.PageSize)).Scan(&result.Data).Error
	} else {
		if scan[0] == nil {
			return nil, errors.New("scan 函数不能为空")
		}

		err = scan[0](o.builder.DbContext.Offset(int((pager.Page - 1) * pager.PageSize)).Limit(int(pager.PageSize)))
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

// UpdateOne 更新单个，主键作为条件；传了字段只更新出入字段，否则更新全部
func (o *OrmWrapper[T]) UpdateOne(item *T, updateColumns ...interface{}) (int64, error) {
	if item == nil {
		return 0, nil
	}

	var isUpdateAll = false
	if len(updateColumns) > 0 {
		o.Select(updateColumns...)
	} else {
		isUpdateAll = true
	}

	o.BuildForQuery()

	//创建语句过程中的错误
	if o.Error != nil {
		return 0, o.Error
	}

	//执行完毕清理痕迹
	defer o.clearBuilder()
	var db = o.builder.DbContext.Model(item)

	if isUpdateAll {
		o.builder.DbContext = db.Save(item)
	} else {
		o.builder.DbContext = db.Updates(item)
	}
	return o.builder.DbContext.RowsAffected, o.builder.DbContext.Error
}

// UpdateList 更新多个，主键作为条件；传了字段只更新传入字段，否则更新全部
func (o *OrmWrapper[T]) UpdateList(items []*T, updateColumns ...interface{}) (int64, error) {
	if len(items) == 0 {
		return 0, nil
	}

	//执行完毕清理痕迹
	defer o.clearBuilder()

	var total int64 = 0

	//外部开启了事务
	if o.builder.isOuterDb {
		for _, item := range items {
			//重新设置db
			o.SetDb(o.builder.ctx, o.builder._DbContext)
			c, err := o.UpdateOne(item, updateColumns...)
			if err != nil {
				return 0, err
			}

			total += c
		}

		return total, nil
	}

	//本地开事务
	var db = o.builder.DbContext

	err := db.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			//重新设置db
			o.SetDb(o.builder.ctx, tx)
			c, err := o.UpdateOne(item, updateColumns...)
			if err != nil {
				return err
			}

			total += c
		}

		o.SetDb(o.builder.ctx, db)

		return nil
	})

	if err != nil {
		return 0, err
	}

	return total, nil
}

// Update 传入字典更新，必须添加查询条件
func (o *OrmWrapper[T]) Update(columnMap map[interface{}]interface{}) (int64, error) {
	if len(columnMap) == 0 {
		o.Error = errors.New("update 更新字段不能为空")
		return 0, o.Error
	}

	if len(o.builder.where) == 0 && len(o.builder.WhereCondition) == 0 {
		o.Error = errors.New("update 更新操作必须加条件")
		return 0, o.Error
	}

	o.BuildForQuery()

	//创建语句过程中的错误
	if o.Error != nil {
		return 0, o.Error
	}

	//执行完毕清理痕迹
	defer o.clearBuilder()

	var m = make(map[string]interface{})
	for key, value := range columnMap {
		name, err := resolveColumnName(key, "")
		if err != nil {
			o.Error = err
			return 0, err
		}

		m[name] = value
	}

	var result *gorm.DB
	result = o.builder.DbContext.Updates(m)
	return result.RowsAffected, result.Error
}

//todo 自定义更新

// Insert 插入单条
func (o *OrmWrapper[T]) Insert(item *T) error {
	if item == nil {
		o.Error = errors.New("item 不能为空")
		return o.Error
	}

	return o.builder.DbContext.Create(item).Error
}

// Inserts 插入多条
func (o *OrmWrapper[T]) Inserts(items []*T) error {
	if len(items) == 0 {
		return nil
	}

	return o.builder.DbContext.Model(new(T)).Create(items).Error
}

// DeleteById 通过id删除数据，可以传入id集合
func (o *OrmWrapper[T]) DeleteById(idList ...interface{}) error {
	if len(idList) == 0 {
		o.Error = errors.New("idList 不能为空")
		return o.Error
	}

	return o.builder.DbContext.Delete(new(T), idList).Error
}

// Delete 根据条件删除
func (o *OrmWrapper[T]) Delete() (int64, error) {
	if len(o.builder.where) == 0 && len(o.builder.WhereCondition) == 0 {
		o.Error = errors.New("删除操作必须加条件")
		return 0, o.Error
	}

	var exc = o.Build().Delete(new(T))
	return exc.RowsAffected, exc.Error
}

// 清理操作痕迹
func (o *OrmWrapper[T]) clearBuilder() {
	o.builder.where = nil
	o.builder.WhereCondition = nil
	o.builder.HavingCondition = nil
	o.builder.joinModels = nil
	o.builder.selectColumns = nil
	o.builder.selectModes = nil
	o.builder.groupByColumns = nil
	o.builder.orderByColumns = nil

	//重新设置db
	o.SetDb(o.builder.ctx, o.builder._DbContext)
	o.Error = nil
}
