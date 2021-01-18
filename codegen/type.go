package codegen

import (
	"bytes"
	"fmt"
	"github.com/zelenin/go-tdlib/tlparser"
)

func GenerateTypes(schema *tlparser.Schema, packageName string) []byte {
	buf := bytes.NewBufferString("")

	buf.WriteString(fmt.Sprintf("%s\n\npackage %s\n\n", header, packageName))

	buf.WriteString(`import (
    "encoding/json"
)

`)

	buf.WriteString("const (\n")
	for _, entity := range schema.Classes {
		tdlibClass := TdlibClass(entity.Name, schema)
		buf.WriteString(fmt.Sprintf("    %s = %q\n", tdlibClass.ToClassConst(), entity.Name))
	}
	for _, entity := range schema.Types {
		tdlibType := TdlibType(entity.Name, schema)
		if tdlibType.IsInternal() || tdlibType.HasClass() {
			continue
		}
		buf.WriteString(fmt.Sprintf("    %s = %q\n", tdlibType.ToClassConst(), entity.Class))
	}
	buf.WriteString(")")

	buf.WriteString("\n\n")

	buf.WriteString("const (\n")
	for _, entity := range schema.Types {
		tdlibType := TdlibType(entity.Name, schema)
		if tdlibType.IsInternal() {
			continue
		}
		buf.WriteString(fmt.Sprintf("    %s = %q\n", tdlibType.ToTypeConst(), entity.Name))
	}
	buf.WriteString(")")

	buf.WriteString("\n\n")

	for _, class := range schema.Classes {
		tdlibClass := TdlibClass(class.Name, schema)

		buf.WriteString(fmt.Sprintf(`// %s
type %s interface {
    %sType() string
}

`, class.Description, tdlibClass.ToGoType(), tdlibClass.ToGoType()))
	}

	for _, typ := range schema.Types {
		tdlibType := TdlibType(typ.Name, schema)
		if tdlibType.IsInternal() {
			continue
		}

		buf.WriteString("// " + typ.Description + "\n")

		if len(typ.Properties) > 0 {
			buf.WriteString(`type ` + tdlibType.ToGoType() + ` struct {
    meta
`)
			for _, property := range typ.Properties {
				tdlibTypeProperty := TdlibTypeProperty(property.Name, property.Type, schema)

				buf.WriteString(fmt.Sprintf("    // %s\n", property.Description))
				buf.WriteString(fmt.Sprintf("    %s %s `json:\"%s\"`\n", tdlibTypeProperty.ToGoName(), tdlibTypeProperty.ToGoType(), property.Name))
			}

			buf.WriteString("}\n\n")
		} else {
			buf.WriteString(`type ` + tdlibType.ToGoType() + ` struct{
    meta
}

`)
		}

		buf.WriteString(fmt.Sprintf(`func (entity *%s) MarshalJSON() ([]byte, error) {
    entity.meta.Type = entity.GetType()

    type stub %s

    return json.Marshal((*stub)(entity))
}

`, tdlibType.ToGoType(), tdlibType.ToGoType()))

		buf.WriteString(fmt.Sprintf(`func (*%s) GetClass() string {
    return %s
}

func (*%s) GetType() string {
    return %s
}

`, tdlibType.ToGoType(), tdlibType.ToClassConst(), tdlibType.ToGoType(), tdlibType.ToTypeConst()))

		if tdlibType.HasClass() {
			tdlibClass := TdlibClass(tdlibType.GetClass().Name, schema)

			buf.WriteString(fmt.Sprintf(`func (*%s) %sType() string {
    return %s
}

`, tdlibType.ToGoType(), tdlibClass.ToGoType(), tdlibType.ToTypeConst()))
		}

		if tdlibType.HasClassProperties() {
			buf.WriteString(fmt.Sprintf(`func (%s *%s) UnmarshalJSON(data []byte) error {
    var tmp struct {
`, typ.Name, tdlibType.ToGoType()))

			var countSimpleProperties int

			for _, property := range typ.Properties {
				tdlibTypeProperty := TdlibTypeProperty(property.Name, property.Type, schema)

				if !tdlibTypeProperty.IsClass() {
					buf.WriteString(fmt.Sprintf("        %s %s `json:\"%s\"`\n", tdlibTypeProperty.ToGoName(), tdlibTypeProperty.ToGoType(), property.Name))
					countSimpleProperties++
				} else {
					if tdlibTypeProperty.IsList() {
						buf.WriteString(fmt.Sprintf("        %s %s `json:\"%s\"`\n", tdlibTypeProperty.ToGoName(), "[]json.RawMessage", property.Name))
					} else {
						buf.WriteString(fmt.Sprintf("        %s %s `json:\"%s\"`\n", tdlibTypeProperty.ToGoName(), "json.RawMessage", property.Name))
					}
				}
			}

			buf.WriteString(`    }

    err := json.Unmarshal(data, &tmp)
    if err != nil {
        return err
    }

`)

			for _, property := range typ.Properties {
				tdlibTypeProperty := TdlibTypeProperty(property.Name, property.Type, schema)

				if !tdlibTypeProperty.IsClass() {
					buf.WriteString(fmt.Sprintf("    %s.%s = tmp.%s\n", typ.Name, tdlibTypeProperty.ToGoName(), tdlibTypeProperty.ToGoName()))
				}
			}

			if countSimpleProperties > 0 {
				buf.WriteString("\n")
			}

			for _, property := range typ.Properties {
				tdlibTypeProperty := TdlibTypeProperty(property.Name, property.Type, schema)

				if tdlibTypeProperty.IsClass() && !tdlibTypeProperty.IsList() {
					buf.WriteString(fmt.Sprintf(`    field%s, _ := Unmarshal%s(tmp.%s)
    %s.%s = field%s

`, tdlibTypeProperty.ToGoName(), tdlibTypeProperty.ToGoType(), tdlibTypeProperty.ToGoName(), typ.Name, tdlibTypeProperty.ToGoName(), tdlibTypeProperty.ToGoName()))
				}
				if tdlibTypeProperty.IsClass() && tdlibTypeProperty.IsList() {
					buf.WriteString(fmt.Sprintf(`    field%s, _ := UnmarshalListOf%s(tmp.%s)
    %s.%s = field%s

`, tdlibTypeProperty.ToGoName(), tdlibTypeProperty.GetClass().ToGoType(), tdlibTypeProperty.ToGoName(), typ.Name, tdlibTypeProperty.ToGoName(), tdlibTypeProperty.ToGoName()))
				}
			}

			buf.WriteString(`    return nil
}

`)
		}
	}

	return buf.Bytes()
}
