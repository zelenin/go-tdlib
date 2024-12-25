package codegen

import (
	"github.com/zelenin/go-tdlib/internal/tlparser"
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

func (entity *tdlibFunctionReturn) IsConstructor() bool {
	return isConstructor(entity.name, func(entity *tlparser.Constructor) string {
		return entity.ResultType
	}, entity.schema)
}

func (entity *tdlibFunctionReturn) GetConstructor() *tdlibConstructor {
	return getConstructor(entity.name, func(entity *tlparser.Constructor) string {
		return entity.ResultType
	}, entity.schema)
}

func (entity *tdlibFunctionReturn) IsType() bool {
	return isType(entity.name, func(entity *tlparser.Type) string {
		return entity.Name
	}, entity.schema)
}

func (entity *tdlibFunctionReturn) GetType() *tdlibType {
	return getType(entity.name, func(entity *tlparser.Type) string {
		return entity.Name
	}, entity.schema)
}

func (entity *tdlibFunctionReturn) ToGoReturn() string {
	if strings.HasPrefix(entity.name, "vector<") {
		log.Fatal("vectors are not supported")
	}

	if entity.IsType() {
		return entity.GetType().ToGoType()
	}

	if entity.GetConstructor().IsInternal() {
		return entity.GetConstructor().ToGoType()
	}

	return "*" + entity.GetConstructor().ToGoType()
}

func (entity *tdlibFunctionReturn) ToGoType() string {
	if strings.HasPrefix(entity.name, "vector<") {
		log.Fatal("vectors are not supported")
	}

	if entity.IsType() {
		return entity.GetType().ToGoType()
	}

	return entity.GetConstructor().ToGoType()
}

type tdlibConstructor struct {
	name   string
	schema *tlparser.Schema
}

func TdlibConstructor(name string, schema *tlparser.Schema) *tdlibConstructor {
	return &tdlibConstructor{
		name:   name,
		schema: schema,
	}
}

func (entity *tdlibConstructor) IsInternal() bool {
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

func (entity *tdlibConstructor) GetConstructor() *tlparser.Constructor {
	name := normalizeEntityName(entity.name)
	for _, constructor := range entity.schema.Constructors {
		if constructor.Name == name {
			return constructor
		}
	}
	return nil
}

func (entity *tdlibConstructor) ToGoType() string {
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

func (entity *tdlibConstructor) ToConstructor() string {
	return entity.ToGoType() + "Constructor"
}

func (entity *tdlibConstructor) HasType() bool {
	typeName := entity.GetConstructor().ResultType
	for _, typ := range entity.schema.Types {
		if typ.Name == typeName {
			return true
		}
	}

	return false
}

func (entity *tdlibConstructor) GetType() *tlparser.Type {
	typeName := entity.GetConstructor().ResultType
	for _, typ := range entity.schema.Types {
		if typ.Name == typeName {
			return typ
		}
	}

	return nil
}

func (entity *tdlibConstructor) HasTypeArgs() bool {
	for _, arg := range entity.GetConstructor().Args {
		tdlibTypeArg := TdlibTypeArg(arg.Name, arg.Type, entity.schema)
		if tdlibTypeArg.IsType() {
			return true
		}
	}

	return false
}

func (entity *tdlibConstructor) IsList() bool {
	return strings.HasPrefix(entity.name, "vector<")
}

func (entity *tdlibConstructor) ToTypeConst() string {
	if entity.HasType() {
		return "Type" + TdlibType(entity.GetConstructor().ResultType, entity.schema).ToGoType()
	}
	return "Type" + entity.ToGoType()
}

func (entity *tdlibConstructor) ToConstructorConst() string {
	return "Constructor" + entity.ToGoType()
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

func (entity *tdlibType) ToGoType() string {
	return firstUpper(entity.name)
}

func (entity *tdlibType) ToConstructor() string {
	return entity.ToGoType() + "Constructor"
}

func (entity *tdlibType) GetConstructors() []*tdlibConstructor {
	constructors := []*tdlibConstructor{}

	for _, constructor := range entity.schema.Constructors {
		if constructor.ResultType == entity.name {
			constructors = append(constructors, TdlibConstructor(constructor.Name, entity.schema))
		}
	}

	return constructors
}

func (entity *tdlibType) ToTypeConst() string {
	return "Type" + entity.ToGoType()
}

type tdlibTypeArg struct {
	name    string
	argType string
	schema  *tlparser.Schema
}

func TdlibTypeArg(name string, argType string, schema *tlparser.Schema) *tdlibTypeArg {
	return &tdlibTypeArg{
		name:    name,
		argType: argType,
		schema:  schema,
	}
}

func (entity *tdlibTypeArg) IsList() bool {
	return strings.HasPrefix(entity.argType, "vector<")
}

func (entity *tdlibTypeArg) GetPrimitive() string {
	primitive := entity.argType

	for strings.HasPrefix(primitive, "vector<") {
		primitive = strings.TrimSuffix(strings.TrimPrefix(primitive, "vector<"), ">")
	}

	return primitive
}

func (entity *tdlibTypeArg) IsConstructor() bool {
	primitive := entity.GetPrimitive()
	return isConstructor(primitive, func(entity *tlparser.Constructor) string {
		return entity.Name
	}, entity.schema)
}

func (entity *tdlibTypeArg) GetConstructor() *tdlibConstructor {
	primitive := entity.GetPrimitive()
	return getConstructor(primitive, func(entity *tlparser.Constructor) string {
		return entity.Name
	}, entity.schema)
}

func (entity *tdlibTypeArg) IsType() bool {
	primitive := entity.GetPrimitive()
	return isType(primitive, func(entity *tlparser.Type) string {
		return entity.Name
	}, entity.schema)
}

func (entity *tdlibTypeArg) GetType() *tdlibType {
	primitive := entity.GetPrimitive()
	return getType(primitive, func(entity *tlparser.Type) string {
		return entity.Name
	}, entity.schema)
}

func (entity *tdlibTypeArg) ToGoName() string {
	return firstUpper(underscoreToCamelCase(entity.name))
}

func (entity *tdlibTypeArg) ToGoFunctionArgName() string {
	name := firstLower(underscoreToCamelCase(entity.name))
	if name == "type" {
		name += "Param"
	}

	return name
}

func (entity *tdlibTypeArg) ToGoType() string {
	tdlibType := entity.argType
	goType := ""

	for strings.HasPrefix(tdlibType, "vector<") {
		goType = goType + "[]"
		tdlibType = strings.TrimSuffix(strings.TrimPrefix(tdlibType, "vector<"), ">")
	}

	if entity.IsType() {
		return goType + entity.GetType().ToGoType()
	}

	if entity.GetConstructor().IsInternal() {
		return goType + entity.GetConstructor().ToGoType()
	}

	return goType + "*" + entity.GetConstructor().ToGoType()
}

func isConstructor(name string, field func(entity *tlparser.Constructor) string, schema *tlparser.Schema) bool {
	name = normalizeEntityName(name)
	for _, entity := range schema.Constructors {
		if name == field(entity) {
			return true
		}
	}

	return false
}

func getConstructor(name string, field func(entity *tlparser.Constructor) string, schema *tlparser.Schema) *tdlibConstructor {
	name = normalizeEntityName(name)
	for _, entity := range schema.Constructors {
		if name == field(entity) {
			return TdlibConstructor(entity.Name, schema)
		}
	}

	return nil
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

func normalizeEntityName(name string) string {
	if name == "Bool" {
		name = "boolFalse"
	}

	return name
}
