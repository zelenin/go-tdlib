package codegen

import (
	"bytes"
	"fmt"
	"github.com/zelenin/go-tdlib/tlparser"
)

func GenerateUnmarshalers(schema *tlparser.Schema, packageName string) []byte {
	buf := bytes.NewBufferString("")

	buf.WriteString(fmt.Sprintf("%s\n\npackage %s\n\n", header, packageName))

	buf.WriteString(`import (
    "encoding/json"
    "fmt"
)

`)

	for _, class := range schema.Classes {
		tdlibClass := TdlibClass(class.Name, schema)

		buf.WriteString(fmt.Sprintf(`func Unmarshal%s(data json.RawMessage) (%s, error) {
    var meta meta

    err := json.Unmarshal(data, &meta)
    if err != nil {
        return nil, err
    }

    switch meta.Type {
`, tdlibClass.ToGoType(), tdlibClass.ToGoType()))

		for _, subType := range tdlibClass.GetSubTypes() {
			buf.WriteString(fmt.Sprintf(`    case %s:
        return Unmarshal%s(data)

`, subType.ToTypeConst(), subType.ToGoType()))

		}

		buf.WriteString(`    default:
        return nil, fmt.Errorf("Error unmarshaling. Unknown type: " +  meta.Type)
    }
}

`)

		buf.WriteString(fmt.Sprintf(`func UnmarshalListOf%s(dataList []json.RawMessage) ([]%s, error) {
    list := []%s{}

    for _, data := range dataList {
        entity, err := Unmarshal%s(data)
        if err != nil {
            return nil, err
        }
        list = append(list, entity)
    }

    return list, nil
}

`, tdlibClass.ToGoType(), tdlibClass.ToGoType(), tdlibClass.ToGoType(), tdlibClass.ToGoType()))

	}

	for _, typ := range schema.Types {
		tdlibType := TdlibType(typ.Name, schema)

		if tdlibType.IsList() || tdlibType.IsInternal() {
			continue
		}

		buf.WriteString(fmt.Sprintf(`func Unmarshal%s(data json.RawMessage) (*%s, error) {
    var resp %s

    err := json.Unmarshal(data, &resp)

    return &resp, err
}

`, tdlibType.ToGoType(), tdlibType.ToGoType(), tdlibType.ToGoType()))

	}

	buf.WriteString(`func UnmarshalType(data json.RawMessage) (Type, error) {
    var meta meta

    err := json.Unmarshal(data, &meta)
    if err != nil {
        return nil, err
    }

    switch meta.Type {
`)

	for _, typ := range schema.Types {
		tdlibType := TdlibType(typ.Name, schema)

		if tdlibType.IsList() || tdlibType.IsInternal() {
			continue
		}

		buf.WriteString(fmt.Sprintf(`    case %s:
        return Unmarshal%s(data)

`, tdlibType.ToTypeConst(), tdlibType.ToGoType()))

	}

	buf.WriteString(`    default:
        return nil, fmt.Errorf("Error unmarshaling. Unknown type: " +  meta.Type)
    }
}
`)

	return buf.Bytes()
}
