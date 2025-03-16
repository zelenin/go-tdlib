package codegen

import (
	"bytes"
	"fmt"
	"github.com/zelenin/go-tdlib/internal/tlparser"
	"strings"
)

func GenerateFunctions(schema *tlparser.Schema, packageName string) []byte {
	buf := bytes.NewBufferString("")

	buf.WriteString(fmt.Sprintf("%s\npackage %s\n\n", header, packageName))

	buf.WriteString(`import (
    "context"
    "errors"
)`)

	buf.WriteString("\n")

	for _, function := range schema.Functions {
		tdlibFunction := TdlibFunction(function.Name, schema)
		tdlibFunctionReturn := TdlibFunctionReturn(function.ResultType, schema)

		buf.WriteString("\n")
		buf.WriteString(fmt.Sprintf("type %sRequest struct { \n", tdlibFunction.ToGoName()))
		buf.WriteString(fmt.Sprintf("    request\n"))
		for _, arg := range function.Args {
			tdlibTypeArg := TdlibTypeArg(arg.Name, arg.Type, schema)

			buf.WriteString(fmt.Sprintf("    // %s\n", arg.Description))
			buf.WriteString(fmt.Sprintf("    %s %s `json:\"%s\"`\n", tdlibTypeArg.ToGoName(), tdlibTypeArg.ToGoType(), arg.Name))
		}
		buf.WriteString("}\n\n")

		buf.WriteString(fmt.Sprintf(`func(req %sRequest) GetFunctionName() string {
    return %q
}
`, tdlibFunction.ToGoName(), function.Name))

		if function.IsSynchronous {
			buf.WriteString("\n")
			buf.WriteString("// " + function.Description)
			buf.WriteString("\n")

			requestArgument := ""
			if len(function.Args) > 0 {
				requestArgument = fmt.Sprintf("req *%sRequest", tdlibFunction.ToGoName())
			}

			buf.WriteString(fmt.Sprintf("func %s(%s) (%s, error) {\n", tdlibFunction.ToGoName(), requestArgument, tdlibFunctionReturn.ToGoReturn()))

			if len(function.Args) == 0 {
				buf.WriteString(fmt.Sprintf("    req := &%sRequest{}\n", tdlibFunction.ToGoName()))
			}
			buf.WriteString(fmt.Sprintf(`    result, err := Execute(req)
`))
			buf.WriteString(`    if err != nil {
        return nil, err
    }

    if result.MetaType == "error" {
        return nil, buildResponseError(result.Data)
    }

`)

			if tdlibFunctionReturn.IsType() {
				buf.WriteString("    switch result.MetaType {\n")

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

		funcArguments := []string{}
		if !function.IsSynchronous {
			funcArguments = append(funcArguments, "ctx context.Context")
		}
		if len(function.Args) > 0 {
			funcArguments = append(funcArguments, fmt.Sprintf("req *%sRequest", tdlibFunction.ToGoName()))
		}

		buf.WriteString(fmt.Sprintf("func (client *Client) %s(%s) (%s, error) {\n", tdlibFunction.ToGoName(), strings.Join(funcArguments, ", "), tdlibFunctionReturn.ToGoReturn()))

		if function.IsSynchronous {
			requestArgument := ""
			if len(function.Args) > 0 {
				requestArgument = "req"
			}
			buf.WriteString(fmt.Sprintf(`    return %s(%s)
`, tdlibFunction.ToGoName(), requestArgument))
		} else {
			if len(function.Args) == 0 {
				buf.WriteString(fmt.Sprintf("    req := &%sRequest{}\n", tdlibFunction.ToGoName()))
			}
			buf.WriteString(fmt.Sprintf(`    result, err := client.Send(ctx, req)
`))
			buf.WriteString(`    if err != nil {
        return nil, err
    }

    if result.MetaType == "error" {
        return nil, buildResponseError(result.Data)
    }

`)

			if tdlibFunctionReturn.IsType() {
				buf.WriteString("    switch result.MetaType {\n")

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
