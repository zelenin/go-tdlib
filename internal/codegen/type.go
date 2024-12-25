package codegen

import (
	"bytes"
	"fmt"
	"github.com/zelenin/go-tdlib/internal/tlparser"
)

func GenerateTypes(schema *tlparser.Schema, packageName string) []byte {
	buf := bytes.NewBufferString("")

	buf.WriteString(fmt.Sprintf("%s\n\npackage %s\n\n", header, packageName))

	buf.WriteString(`import (
    "encoding/json"
)

`)

	buf.WriteString("const (\n")
	for _, entity := range schema.Types {
		tdlibType := TdlibType(entity.Name, schema)
		buf.WriteString(fmt.Sprintf("    %s = %q\n", tdlibType.ToTypeConst(), entity.Name))
	}
	for _, entity := range schema.Constructors {
		tdlibConstructor := TdlibConstructor(entity.Name, schema)
		if tdlibConstructor.IsInternal() || tdlibConstructor.HasType() {
			continue
		}
		buf.WriteString(fmt.Sprintf("    %s = %q\n", tdlibConstructor.ToTypeConst(), entity.ResultType))
	}
	buf.WriteString(")")

	buf.WriteString("\n\n")

	buf.WriteString("const (\n")
	for _, entity := range schema.Constructors {
		tdlibConstructor := TdlibConstructor(entity.Name, schema)
		if tdlibConstructor.IsInternal() {
			continue
		}
		buf.WriteString(fmt.Sprintf("    %s = %q\n", tdlibConstructor.ToConstructorConst(), entity.Name))
	}
	buf.WriteString(")")

	buf.WriteString("\n\n")

	for _, typ := range schema.Types {
		tdlibType := TdlibType(typ.Name, schema)

		buf.WriteString(fmt.Sprintf(`// %s
type %s interface {
    %sConstructor() string
}

`, typ.Description, tdlibType.ToGoType(), tdlibType.ToGoType()))
	}

	for _, constructor := range schema.Constructors {
		tdlibConstructor := TdlibConstructor(constructor.Name, schema)
		if tdlibConstructor.IsInternal() {
			continue
		}

		buf.WriteString("// " + constructor.Description + "\n")

		if len(constructor.Args) > 0 {
			buf.WriteString(`type ` + tdlibConstructor.ToGoType() + ` struct {
    meta
`)
			for _, arg := range constructor.Args {
				tdlibTypeArg := TdlibTypeArg(arg.Name, arg.Type, schema)

				buf.WriteString(fmt.Sprintf("    // %s\n", arg.Description))
				buf.WriteString(fmt.Sprintf("    %s %s `json:\"%s\"`\n", tdlibTypeArg.ToGoName(), tdlibTypeArg.ToGoType(), arg.Name))
			}

			buf.WriteString("}\n\n")
		} else {
			buf.WriteString(`type ` + tdlibConstructor.ToGoType() + ` struct{
    meta
}

`)
		}

		buf.WriteString(fmt.Sprintf(`func (entity *%s) MarshalJSON() ([]byte, error) {
    entity.meta.Type = entity.GetConstructor()

    type stub %s

    return json.Marshal((*stub)(entity))
}

`, tdlibConstructor.ToGoType(), tdlibConstructor.ToGoType()))

		buf.WriteString(fmt.Sprintf(`func (*%s) GetType() string {
    return %s
}

func (*%s) GetConstructor() string {
    return %s
}

`, tdlibConstructor.ToGoType(), tdlibConstructor.ToTypeConst(), tdlibConstructor.ToGoType(), tdlibConstructor.ToConstructorConst()))

		if tdlibConstructor.HasType() {
			tdlibType := TdlibType(tdlibConstructor.GetType().Name, schema)

			buf.WriteString(fmt.Sprintf(`func (*%s) %sConstructor() string {
    return %s
}

`, tdlibConstructor.ToGoType(), tdlibType.ToGoType(), tdlibConstructor.ToConstructorConst()))
		}

		if tdlibConstructor.HasTypeArgs() {
			buf.WriteString(fmt.Sprintf(`func (%s *%s) UnmarshalJSON(data []byte) error {
    var tmp struct {
`, constructor.Name, tdlibConstructor.ToGoType()))

			var countSimpleProperties int

			for _, arg := range constructor.Args {
				tdlibTypeArg := TdlibTypeArg(arg.Name, arg.Type, schema)

				if !tdlibTypeArg.IsType() {
					buf.WriteString(fmt.Sprintf("        %s %s `json:\"%s\"`\n", tdlibTypeArg.ToGoName(), tdlibTypeArg.ToGoType(), arg.Name))
					countSimpleProperties++
				} else {
					if tdlibTypeArg.IsList() {
						buf.WriteString(fmt.Sprintf("        %s %s `json:\"%s\"`\n", tdlibTypeArg.ToGoName(), "[]json.RawMessage", arg.Name))
					} else {
						buf.WriteString(fmt.Sprintf("        %s %s `json:\"%s\"`\n", tdlibTypeArg.ToGoName(), "json.RawMessage", arg.Name))
					}
				}
			}

			buf.WriteString(`    }

    err := json.Unmarshal(data, &tmp)
    if err != nil {
        return err
    }

`)

			for _, arg := range constructor.Args {
				tdlibTypeArg := TdlibTypeArg(arg.Name, arg.Type, schema)

				if !tdlibTypeArg.IsType() {
					buf.WriteString(fmt.Sprintf("    %s.%s = tmp.%s\n", constructor.Name, tdlibTypeArg.ToGoName(), tdlibTypeArg.ToGoName()))
				}
			}

			if countSimpleProperties > 0 {
				buf.WriteString("\n")
			}

			for _, arg := range constructor.Args {
				tdlibTypeArg := TdlibTypeArg(arg.Name, arg.Type, schema)

				if tdlibTypeArg.IsType() && !tdlibTypeArg.IsList() {
					buf.WriteString(fmt.Sprintf(`    field%s, _ := Unmarshal%s(tmp.%s)
    %s.%s = field%s

`, tdlibTypeArg.ToGoName(), tdlibTypeArg.ToGoType(), tdlibTypeArg.ToGoName(), constructor.Name, tdlibTypeArg.ToGoName(), tdlibTypeArg.ToGoName()))
				}
				if tdlibTypeArg.IsType() && tdlibTypeArg.IsList() {
					buf.WriteString(fmt.Sprintf(`    field%s, _ := UnmarshalListOf%s(tmp.%s)
    %s.%s = field%s

`, tdlibTypeArg.ToGoName(), tdlibTypeArg.GetType().ToGoType(), tdlibTypeArg.ToGoName(), constructor.Name, tdlibTypeArg.ToGoName(), tdlibTypeArg.ToGoName()))
				}
			}

			buf.WriteString(`    return nil
}

`)
		}
	}

	return buf.Bytes()
}
