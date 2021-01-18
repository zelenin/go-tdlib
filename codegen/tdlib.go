package codegen

import (
	"github.com/zelenin/go-tdlib/tlparser"
	"log"
	"strings"
)

type tdlibFunction struct {
	name   string
	schema *tlparser.Schema
}

func TdlibFunction(name string, schema *tlparser.Schema) *tdlibFunction {
	return &tdlibFunction{
		name:   name,
		schema: schema,
	}
}

func (entity *tdlibFunction) ToGoName() string {
	return firstUpper(entity.name)
}

type tdlibFunctionReturn struct {
	name   string
	schema *tlparser.Schema
}

func TdlibFunctionReturn(name string, schema *tlparser.Schema) *tdlibFunctionReturn {
	return &tdlibFunctionReturn{
		name:   name,
		schema: schema,
	}
}

func (entity *tdlibFunctionReturn) IsType() bool {
	return isType(entity.name, func(entity *tlparser.Type) string {
		return entity.Class
	}, entity.schema)
}

func (entity *tdlibFunctionReturn) GetType() *tdlibType {
	return getType(entity.name, func(entity *tlparser.Type) string {
		return entity.Class
	}, entity.schema)
}

func (entity *tdlibFunctionReturn) IsClass() bool {
	return isClass(entity.name, func(entity *tlparser.Class) string {
		return entity.Name
	}, entity.schema)
}

func (entity *tdlibFunctionReturn) GetClass() *tdlibClass {
	return getClass(entity.name, func(entity *tlparser.Class) string {
		return entity.Name
	}, entity.schema)
}

func (entity *tdlibFunctionReturn) ToGoReturn() string {
	if strings.HasPrefix(entity.name, "vector<") {
		log.Fatal("vectors are not supported")
	}

	if entity.IsClass() {
		return entity.GetClass().ToGoType()
	}

	if entity.GetType().IsInternal() {
		return entity.GetType().ToGoType()
	}

	return "*" + entity.GetType().ToGoType()
}

func (entity *tdlibFunctionReturn) ToGoType() string {
	if strings.HasPrefix(entity.name, "vector<") {
		log.Fatal("vectors are not supported")
	}

	if entity.IsClass() {
		return entity.GetClass().ToGoType()
	}

	return entity.GetType().ToGoType()
}

type tdlibFunctionProperty struct {
	name         string
	propertyType string
	schema       *tlparser.Schema
}

func TdlibFunctionProperty(name string, propertyType string, schema *tlparser.Schema) *tdlibFunctionProperty {
	return &tdlibFunctionProperty{
		name:         name,
		propertyType: propertyType,
		schema:       schema,
	}
}

func (entity *tdlibFunctionProperty) GetPrimitive() string {
	primitive := entity.propertyType

	for strings.HasPrefix(primitive, "vector<") {
		primitive = strings.TrimSuffix(strings.TrimPrefix(primitive, "vector<"), ">")
	}

	return primitive
}

func (entity *tdlibFunctionProperty) IsType() bool {
	primitive := entity.GetPrimitive()
	return isType(primitive, func(entity *tlparser.Type) string {
		return entity.Name
	}, entity.schema)
}

func (entity *tdlibFunctionProperty) GetType() *tdlibType {
	primitive := entity.GetPrimitive()
	return getType(primitive, func(entity *tlparser.Type) string {
		return entity.Name
	}, entity.schema)
}

func (entity *tdlibFunctionProperty) IsClass() bool {
	primitive := entity.GetPrimitive()
	return isClass(primitive, func(entity *tlparser.Class) string {
		return entity.Name
	}, entity.schema)
}

func (entity *tdlibFunctionProperty) GetClass() *tdlibClass {
	primitive := entity.GetPrimitive()
	return getClass(primitive, func(entity *tlparser.Class) string {
		return entity.Name
	}, entity.schema)
}

func (entity *tdlibFunctionProperty) ToGoName() string {
	name := firstLower(underscoreToCamelCase(entity.name))
	if name == "type" {
		name += "Param"
	}

	return name
}

func (entity *tdlibFunctionProperty) ToGoType() string {
	tdlibType := entity.propertyType
	goType := ""

	for strings.HasPrefix(tdlibType, "vector<") {
		goType = goType + "[]"
		tdlibType = strings.TrimSuffix(strings.TrimPrefix(tdlibType, "vector<"), ">")
	}

	if entity.IsClass() {
		return goType + entity.GetClass().ToGoType()
	}

	if entity.GetType().IsInternal() {
		return goType + entity.GetType().ToGoType()
	}

	return goType + "*" + entity.GetType().ToGoType()
}

type tdlibType struct {
	name   string
	schema *tlparser.Schema
}

func TdlibType(name string, schema *tlparser.Schema) *tdlibType {
	return &tdlibType{
		name:   name,
		schema: schema,
	}
}

func (entity *tdlibType) IsInternal() bool {
	switch entity.name {
	case "double":
		return true

	case "string":
		return true

	case "int32":
		return true

	case "int53":
		return true

	case "int64":
		return true

	case "bytes":
		return true

	case "boolFalse":
		return true

	case "boolTrue":
		return true

	case "vector<t>":
		return true
	}

	return false
}

func (entity *tdlibType) GetType() *tlparser.Type {
	name := normalizeEntityName(entity.name)
	for _, typ := range entity.schema.Types {
		if typ.Name == name {
			return typ
		}
	}
	return nil
}

func (entity *tdlibType) ToGoType() string {
	if strings.HasPrefix(entity.name, "vector<") {
		log.Fatal("vectors are not supported")
	}

	switch entity.name {
	case "double":
		return "float64"

	case "string":
		return "string"

	case "int32":
		return "int32"

	case "int53":
		return "int64"

	case "int64":
		return "JsonInt64"

	case "bytes":
		return "[]byte"

	case "boolFalse":
		return "bool"

	case "boolTrue":
		return "bool"
	}

	return firstUpper(entity.name)
}

func (entity *tdlibType) ToType() string {
	return entity.ToGoType() + "Type"
}

func (entity *tdlibType) HasClass() bool {
	className := entity.GetType().Class
	for _, class := range entity.schema.Classes {
		if class.Name == className {
			return true
		}
	}

	return false
}

func (entity *tdlibType) GetClass() *tlparser.Class {
	className := entity.GetType().Class
	for _, class := range entity.schema.Classes {
		if class.Name == className {
			return class
		}
	}

	return nil
}

func (entity *tdlibType) HasClassProperties() bool {
	for _, prop := range entity.GetType().Properties {
		tdlibTypeProperty := TdlibTypeProperty(prop.Name, prop.Type, entity.schema)
		if tdlibTypeProperty.IsClass() {
			return true
		}
	}

	return false
}

func (entity *tdlibType) IsList() bool {
	return strings.HasPrefix(entity.name, "vector<")
}

func (entity *tdlibType) ToClassConst() string {
	if entity.HasClass() {
		return "Class" + TdlibClass(entity.GetType().Class, entity.schema).ToGoType()
	}
	return "Class" + entity.ToGoType()
}

func (entity *tdlibType) ToTypeConst() string {
	return "Type" + entity.ToGoType()
}

type tdlibClass struct {
	name   string
	schema *tlparser.Schema
}

func TdlibClass(name string, schema *tlparser.Schema) *tdlibClass {
	return &tdlibClass{
		name:   name,
		schema: schema,
	}
}

func (entity *tdlibClass) ToGoType() string {
	return firstUpper(entity.name)
}

func (entity *tdlibClass) ToType() string {
	return entity.ToGoType() + "Type"
}

func (entity *tdlibClass) GetSubTypes() []*tdlibType {
	types := []*tdlibType{}

	for _, t := range entity.schema.Types {
		if t.Class == entity.name {
			types = append(types, TdlibType(t.Name, entity.schema))
		}
	}

	return types
}

func (entity *tdlibClass) ToClassConst() string {
	return "Class" + entity.ToGoType()
}

type tdlibTypeProperty struct {
	name         string
	propertyType string
	schema       *tlparser.Schema
}

func TdlibTypeProperty(name string, propertyType string, schema *tlparser.Schema) *tdlibTypeProperty {
	return &tdlibTypeProperty{
		name:         name,
		propertyType: propertyType,
		schema:       schema,
	}
}

func (entity *tdlibTypeProperty) IsList() bool {
	return strings.HasPrefix(entity.propertyType, "vector<")
}

func (entity *tdlibTypeProperty) GetPrimitive() string {
	primitive := entity.propertyType

	for strings.HasPrefix(primitive, "vector<") {
		primitive = strings.TrimSuffix(strings.TrimPrefix(primitive, "vector<"), ">")
	}

	return primitive
}

func (entity *tdlibTypeProperty) IsType() bool {
	primitive := entity.GetPrimitive()
	return isType(primitive, func(entity *tlparser.Type) string {
		return entity.Name
	}, entity.schema)
}

func (entity *tdlibTypeProperty) GetType() *tdlibType {
	primitive := entity.GetPrimitive()
	return getType(primitive, func(entity *tlparser.Type) string {
		return entity.Name
	}, entity.schema)
}

func (entity *tdlibTypeProperty) IsClass() bool {
	primitive := entity.GetPrimitive()
	return isClass(primitive, func(entity *tlparser.Class) string {
		return entity.Name
	}, entity.schema)
}

func (entity *tdlibTypeProperty) GetClass() *tdlibClass {
	primitive := entity.GetPrimitive()
	return getClass(primitive, func(entity *tlparser.Class) string {
		return entity.Name
	}, entity.schema)
}

func (entity *tdlibTypeProperty) ToGoName() string {
	return firstUpper(underscoreToCamelCase(entity.name))
}

func (entity *tdlibTypeProperty) ToGoFunctionPropertyName() string {
	name := firstLower(underscoreToCamelCase(entity.name))
	if name == "type" {
		name += "Param"
	}

	return name
}

func (entity *tdlibTypeProperty) ToGoType() string {
	tdlibType := entity.propertyType
	goType := ""

	for strings.HasPrefix(tdlibType, "vector<") {
		goType = goType + "[]"
		tdlibType = strings.TrimSuffix(strings.TrimPrefix(tdlibType, "vector<"), ">")
	}

	if entity.IsClass() {
		return goType + entity.GetClass().ToGoType()
	}

	if entity.GetType().IsInternal() {
		return goType + entity.GetType().ToGoType()
	}

	return goType + "*" + entity.GetType().ToGoType()
}

func isType(name string, field func(entity *tlparser.Type) string, schema *tlparser.Schema) bool {
	name = normalizeEntityName(name)
	for _, entity := range schema.Types {
		if name == field(entity) {
			return true
		}
	}

	return false
}

func getType(name string, field func(entity *tlparser.Type) string, schema *tlparser.Schema) *tdlibType {
	name = normalizeEntityName(name)
	for _, entity := range schema.Types {
		if name == field(entity) {
			return TdlibType(entity.Name, schema)
		}
	}

	return nil
}

func isClass(name string, field func(entity *tlparser.Class) string, schema *tlparser.Schema) bool {
	name = normalizeEntityName(name)
	for _, entity := range schema.Classes {
		if name == field(entity) {
			return true
		}
	}

	return false
}

func getClass(name string, field func(entity *tlparser.Class) string, schema *tlparser.Schema) *tdlibClass {
	name = normalizeEntityName(name)
	for _, entity := range schema.Classes {
		if name == field(entity) {
			return TdlibClass(entity.Name, schema)
		}
	}

	return nil
}

func normalizeEntityName(name string) string {
	if name == "Bool" {
		name = "boolFalse"
	}

	return name
}
