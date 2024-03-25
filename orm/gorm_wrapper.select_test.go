package orm

import "testing"

func getOrmWrapper[T interface{}]() *OrmWrapper[T] {
	return &OrmWrapper[T]{builder: &ormWrapperBuilder[T]{}}
}

func Test_Select(t *testing.T) {
	var orm = getOrmWrapper[Table1]()

	//Select
	orm.Select("name")
	if len(orm.builder.selectModes) != 1 {
		t.Errorf("Test_Select faild")
	}
	var column = orm.builder.selectModes[0].Column
	if column.(string) != "name" {
		t.Errorf("Test_Select faild")
	}

	orm.Select(_db)
	if orm.Error == nil {
		t.Errorf("Test_Select faild")
	}

	//SelectTable
	orm = getOrmWrapper[Table1]()
	orm.SelectTable("a", "name")
	if len(orm.builder.selectModes) != 1 {
		t.Errorf("Test_Select faild")
	}
	var item = orm.builder.selectModes[0]
	if item.Column.(string) != "name" || item.TableAlias != "a" {
		t.Errorf("Test_Select faild")
	}

	//SelectTable
	orm = getOrmWrapper[Table1]()
	orm.SelectTable("a")
	if len(orm.builder.selectModes) != 0 {
		t.Errorf("Test_Select faild")
	}

	//SelectOne
	orm = getOrmWrapper[Table1]()
	orm.SelectOne("name", "a", "b")
	if len(orm.builder.selectModes) != 1 {
		t.Errorf("Test_Select faild")
	}
	item = orm.builder.selectModes[0]
	if item.Column.(string) != "name" || item.ColumnAlias != "a" || item.TableAlias != "b" {
		t.Errorf("Test_Select faild")
	}

	//SelectOneWithFunc
	orm = getOrmWrapper[Table1]()
	orm.SelectOneWithFunc("name", "a", "max", "b")
	if len(orm.builder.selectModes) != 1 {
		t.Errorf("Test_Select faild")
	}
	item = orm.builder.selectModes[0]
	if item.Column.(string) != "name" || item.ColumnAlias != "a" || item.TableAlias != "b" || item.Func != "max" {
		t.Errorf("Test_Select faild")
	}
}
