package codegen

import (
	"bytes"
	"fmt"

	"github.com/zelenin/go-tdlib/tlparser"
)

func GenerateFunctions(schema *tlparser.Schema, packageName string) []byte {
	buf := bytes.NewBufferString("")

	buf.WriteString(fmt.Sprintf("%s\n\npackage %s\n\n", header, packageName))

	buf.WriteString(`import (
    "errors"
)`)

	buf.WriteString("\n")

	for _, function := range schema.Functions {
		tdlibFunction := TdlibFunction(function.Name, schema)
		tdlibFunctionReturn := TdlibFunctionReturn(function.Class, schema)

		if len(function.Properties) > 0 {
			buf.WriteString("\n")
			buf.WriteString(fmt.Sprintf("type %sRequest struct { \n", tdlibFunction.ToGoName()))
			for _, property := range function.Properties {
				tdlibTypeProperty := TdlibTypeProperty(property.Name, property.Type, schema)

				buf.WriteString(fmt.Sprintf("    // %s\n", property.Description))
				buf.WriteString(fmt.Sprintf("    %s %s `json:\"%s\"`\n", tdlibTypeProperty.ToGoName(), tdlibTypeProperty.ToGoType(), property.Name))
			}
			buf.WriteString("}\n")
		}

		if function.IsSynchronous {
			buf.WriteString("\n")
			buf.WriteString("// " + function.Description)
			buf.WriteString("\n")

			requestArgument := ""
			if len(function.Properties) > 0 {
				requestArgument = fmt.Sprintf("req *%sRequest", tdlibFunction.ToGoName())
			}

			buf.WriteString(fmt.Sprintf("func %s(%s) (%s, error) {\n", tdlibFunction.ToGoName(), requestArgument, tdlibFunctionReturn.ToGoReturn()))

			if len(function.Properties) > 0 {
				buf.WriteString(fmt.Sprintf(`    result, err := Execute(Request{
        meta: meta{
            Type: "%s",
        },
        Data: map[string]interface{}{
`, function.Name))

				for _, property := range function.Properties {
					tdlibTypeProperty := TdlibTypeProperty(property.Name, property.Type, schema)

					buf.WriteString(fmt.Sprintf("            \"%s\": req.%s,\n", property.Name, tdlibTypeProperty.ToGoName()))
				}

				buf.WriteString(`        },
    })
`)
			} else {
				buf.WriteString(fmt.Sprintf(`    result, err := Execute(Request{
        meta: meta{
            Type: "%s",
        },
        Data: map[string]interface{}{},
    })
`, function.Name))
			}

			buf.WriteString(`    if err != nil {
        return nil, err
    }

    if result.Type == "error" {
        return nil, buildResponseError(result.Data)
    }

`)

			if tdlibFunctionReturn.IsClass() {
				buf.WriteString("    switch result.Type {\n")

				for _, subType := range tdlibFunctionReturn.GetClass().GetSubTypes() {
					buf.WriteString(fmt.Sprintf(`    case %s:
        return Unmarshal%s(result.Data)

`, subType.ToTypeConst(), subType.ToGoType()))

				}

				buf.WriteString(`    default:
        return nil, errors.New("invalid type")
`)

				buf.WriteString("   }\n")
			} else {
				buf.WriteString(fmt.Sprintf(`    return Unmarshal%s(result.Data)
`, tdlibFunctionReturn.ToGoType()))
			}

			buf.WriteString("}\n")
		}

		buf.WriteString("\n")
		if function.IsSynchronous {
			buf.WriteString("// deprecated")
			buf.WriteString("\n")
		}
		buf.WriteString("// " + function.Description)
		buf.WriteString("\n")

		requestArgument := ""
		if len(function.Properties) > 0 {
			requestArgument = fmt.Sprintf("req *%sRequest", tdlibFunction.ToGoName())
		}

		buf.WriteString(fmt.Sprintf("func (client *Client) %s(%s) (%s, error) {\n", tdlibFunction.ToGoName(), requestArgument, tdlibFunctionReturn.ToGoReturn()))

		if function.IsSynchronous {
			requestArgument = ""
			if len(function.Properties) > 0 {
				requestArgument = "req"
			}
			buf.WriteString(fmt.Sprintf(`    return %s(%s)`, tdlibFunction.ToGoName(), requestArgument))
		} else {
			if len(function.Properties) > 0 {
				buf.WriteString(fmt.Sprintf(`    result, err := client.Send(Request{
        meta: meta{
            Type: "%s",
        },
        Data: map[string]interface{}{
`, function.Name))

				for _, property := range function.Properties {
					tdlibTypeProperty := TdlibTypeProperty(property.Name, property.Type, schema)

					buf.WriteString(fmt.Sprintf("            \"%s\": req.%s,\n", property.Name, tdlibTypeProperty.ToGoName()))
				}

				buf.WriteString(`        },
    })
`)
			} else {
				buf.WriteString(fmt.Sprintf(`    result, err := client.Send(Request{
        meta: meta{
            Type: "%s",
        },
        Data: map[string]interface{}{},
    })
`, function.Name))
			}

			buf.WriteString(`    if err != nil {
        return nil, err
    }

    if result.Type == "error" {
        return nil, buildResponseError(result.Data)
    }

`)

			if tdlibFunctionReturn.IsClass() {
				buf.WriteString("    switch result.Type {\n")

				for _, subType := range tdlibFunctionReturn.GetClass().GetSubTypes() {
					buf.WriteString(fmt.Sprintf(`    case %s:
        return Unmarshal%s(result.Data)

`, subType.ToTypeConst(), subType.ToGoType()))

				}

				buf.WriteString(`    default:
        return nil, errors.New("invalid type")
`)

				buf.WriteString("   }\n")
			} else {
				buf.WriteString(fmt.Sprintf(`    return Unmarshal%s(result.Data)
`, tdlibFunctionReturn.ToGoType()))
			}
		}

		buf.WriteString("}\n")
	}

	return buf.Bytes()
}
