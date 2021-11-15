package schema

import (
	"go/ast"
	"orm/dialect"
	"reflect"
)

// Obj <--> Table

// Field represents a column of database
type Field struct {
	// 字段名
	Name string
	// 字段类型
	Type string
	// 额外的约束条件(例如非空、主键等)
	Tag  string
}

// Schema represents a table of database
type Schema struct {
	// 被映射的对象model
	Model      interface{}
	// 表名
	Name       string
	Fields     []*Field
	FieldNames []string
	// 记录fieldname 和 fields的对应关系，不用遍历确定了
	fieldMap   map[string]*Field
}

// GetField returns field by name
func (schema *Schema) GetField(name string) *Field {
	return schema.fieldMap[name]
}

type ITableName interface {
	TableName() string
}

// Values return the values of dest's member variables
func (schema *Schema) RecordValues(dest interface{}) []interface{} {
	destValue := reflect.Indirect(reflect.ValueOf(dest))
	var fieldValues []interface{}
	for _, field := range schema.Fields {
		fieldValues = append(fieldValues, destValue.FieldByName(field.Name).Interface())
	}
	return fieldValues
}

// 将对象解析为Schema实例
func Parse(dest interface{}, d dialect.Dialect) *Schema {
	// 因为设计的入参是一个对象的指针，因此需要 reflect.Indirect() 获取指针指向的实例
	modelType := reflect.Indirect(reflect.ValueOf(dest)).Type()
	var tableName string
	t, ok := dest.(ITableName)
	if !ok {
		tableName = modelType.Name()
	} else {
		tableName = t.TableName()
	}
	schema := &Schema{
		Model:    dest,
		Name:     tableName,
		fieldMap: make(map[string]*Field),
	}

	for i := 0; i < modelType.NumField(); i++ {
		p := modelType.Field(i)
		if !p.Anonymous && ast.IsExported(p.Name) {
			field := &Field{
				Name: p.Name,
				// 获取映射对象结构体类型，转换成对应的数据库类型
				Type: d.DataTypeOf(reflect.Indirect(reflect.New(p.Type))),
			}
			if v, ok := p.Tag.Lookup("geeorm"); ok {
				field.Tag = v
			}
			schema.Fields = append(schema.Fields, field)
			schema.FieldNames = append(schema.FieldNames, p.Name)
			schema.fieldMap[p.Name] = field
		}
	}
	return schema
}
