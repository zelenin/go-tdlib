package codegen

import (
	"bytes"
	"fmt"
	"github.com/zelenin/go-tdlib/internal/tlparser"
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
		tdlibFunctionReturn := TdlibFunctionReturn(function.ResultType, schema)

		if len(function.Args) > 0 {
			buf.WriteString("\n")
			buf.WriteString(fmt.Sprintf("type %sRequest struct { \n", tdlibFunction.ToGoName()))
			for _, arg := range function.Args {
				tdlibTypeArg := TdlibTypeArg(arg.Name, arg.Type, schema)

				buf.WriteString(fmt.Sprintf("    // %s\n", arg.Description))
				buf.WriteString(fmt.Sprintf("    %s %s `json:\"%s\"`\n", tdlibTypeArg.ToGoName(), tdlibTypeArg.ToGoType(), arg.Name))
			}
			buf.WriteString("}\n")
		}

		if function.IsSynchronous {
			buf.WriteString("\n")
			buf.WriteString("// " + function.Description)
			buf.WriteString("\n")

			requestArgument := ""
			if len(function.Args) > 0 {
				requestArgument = fmt.Sprintf("req *%sRequest", tdlibFunction.ToGoName())
			}

			buf.WriteString(fmt.Sprintf("func %s(%s) (%s, error) {\n", tdlibFunction.ToGoName(), requestArgument, tdlibFunctionReturn.ToGoReturn()))

			if len(function.Args) > 0 {
				buf.WriteString(fmt.Sprintf(`    result, err := Execute(Request{
        meta: meta{
            Type: "%s",
        },
        Data: map[string]interface{}{
`, function.Name))

				for _, arg := range function.Args {
					tdlibTypeArg := TdlibTypeArg(arg.Name, arg.Type, schema)

					buf.WriteString(fmt.Sprintf("            \"%s\": req.%s,\n", arg.Name, tdlibTypeArg.ToGoName()))
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

			if tdlibFunctionReturn.IsType() {
				buf.WriteString("    switch result.Type {\n")

				for _, constructor := range tdlibFunctionReturn.GetType().GetConstructors() {
					buf.WriteString(fmt.Sprintf(`    case %s:
        return Unmarshal%s(result.Data)

`, constructor.ToConstructorConst(), constructor.ToGoType()))

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
		if len(function.Args) > 0 {
			requestArgument = fmt.Sprintf("req *%sRequest", tdlibFunction.ToGoName())
		}

		buf.WriteString(fmt.Sprintf("func (client *Client) %s(%s) (%s, error) {\n", tdlibFunction.ToGoName(), requestArgument, tdlibFunctionReturn.ToGoReturn()))

		if function.IsSynchronous {
			requestArgument = ""
			if len(function.Args) > 0 {
				requestArgument = "req"
			}
			buf.WriteString(fmt.Sprintf(`    return %s(%s)`, tdlibFunction.ToGoName(), requestArgument))
		} else {
			if len(function.Args) > 0 {
				buf.WriteString(fmt.Sprintf(`    result, err := client.Send(Request{
        meta: meta{
            Type: "%s",
        },
        Data: map[string]interface{}{
`, function.Name))

				for _, arg := range function.Args {
					tdlibTypeArg := TdlibTypeArg(arg.Name, arg.Type, schema)

					buf.WriteString(fmt.Sprintf("            \"%s\": req.%s,\n", arg.Name, tdlibTypeArg.ToGoName()))
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

			if tdlibFunctionReturn.IsType() {
				buf.WriteString("    switch result.Type {\n")

				for _, constructor := range tdlibFunctionReturn.GetType().GetConstructors() {
					buf.WriteString(fmt.Sprintf(`    case %s:
        return Unmarshal%s(result.Data)

`, constructor.ToConstructorConst(), constructor.ToGoType()))

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
