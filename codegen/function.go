package codegen

import (
    "github.com/zelenin/go-tdlib/tlparser"
    "fmt"
    "strings"
    "bytes"
)

func GenerateFunctions(schema *tlparser.Schema, packageName string) []byte {
    buf := bytes.NewBufferString("")

    buf.WriteString(fmt.Sprintf("%s\n\npackage %s\n\n", header, packageName))

    buf.WriteString(`import (
    "errors"
)`)

    buf.WriteString("\n")

    for _, function := range schema.Functions {
        buf.WriteString("\n")
        buf.WriteString("// " + function.Description)
        buf.WriteString("\n")

        if len(function.Properties) > 0 {
            buf.WriteString("//")
            buf.WriteString("\n")
        }

        propertiesParts := []string{}
        for _, property := range function.Properties {
            tdlibFunctionProperty := TdlibFunctionProperty(property.Name, property.Type, schema)

            buf.WriteString(fmt.Sprintf("// @param %s %s", tdlibFunctionProperty.ToGoName(), property.Description))
            buf.WriteString("\n")

            propertiesParts = append(propertiesParts, tdlibFunctionProperty.ToGoName()+" "+tdlibFunctionProperty.ToGoType())
        }

        tdlibFunction := TdlibFunction(function.Name, schema)
        tdlibFunctionReturn := TdlibFunctionReturn(function.Class, schema)

        buf.WriteString(fmt.Sprintf("func (client *Client) %s(%s) (%s, error) {\n", tdlibFunction.ToGoName(), strings.Join(propertiesParts, ", "), tdlibFunctionReturn.ToGoReturn()))

        sendMethod := "Send"
        if function.IsSynchronous {
            sendMethod = "jsonClient.Execute"
        }

        if len(function.Properties) > 0 {
            buf.WriteString(fmt.Sprintf(`    result, err := client.%s(Request{
        meta: meta{
            Type: "%s",
        },
        Data: map[string]interface{}{
`, sendMethod, function.Name))

            for _, property := range function.Properties {
                tdlibFunctionProperty := TdlibFunctionProperty(property.Name, property.Type, schema)
                buf.WriteString(fmt.Sprintf("            \"%s\": %s,\n", property.Name, tdlibFunctionProperty.ToGoName()))
            }

            buf.WriteString(`        },
    })
`)
        } else {
            buf.WriteString(fmt.Sprintf(`    result, err := client.%s(Request{
        meta: meta{
            Type: "%s",
        },
        Data: map[string]interface{}{},
    })
`, sendMethod, function.Name))
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

    return buf.Bytes()
}
