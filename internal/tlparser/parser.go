package tlparser

import (
	"bufio"
	"io"
	"strings"
)

func Parse(reader io.Reader) (*Schema, error) {
	schema := &Schema{
		Constructors: []*Constructor{},
		Types:        []*Type{},
		Functions:    []*Function{},
	}

	scanner := bufio.NewScanner(reader)

	hitFunctions := false

	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "//@description"):
			if hitFunctions {
				schema.Functions = append(schema.Functions, parseFunction(line, scanner))
			} else {
				schema.Constructors = append(schema.Constructors, parseConstructor(line, scanner))
			}

		case strings.HasPrefix(line, "//@class"):
			schema.Types = append(schema.Types, parseType(line, scanner))

		case strings.Contains(line, "---functions---"):
			hitFunctions = true

		case line == "":

		default:
			bodyFields := strings.Fields(line)
			name := bodyFields[0]
			resultType := strings.TrimRight(bodyFields[len(bodyFields)-1], ";")
			if hitFunctions {
				schema.Functions = append(schema.Functions, &Function{
					Name:          name,
					Description:   "",
					Args:          []*Arg{},
					ResultType:    resultType,
					IsSynchronous: false,
					Type:          FUNCTION_TYPE_COMMON,
				})
			} else {
				if name == "vector" {
					name = "vector<t>"
					resultType = "Vector<T>"
				}

				schema.Constructors = append(schema.Constructors, &Constructor{
					Name:        name,
					Description: "",
					Args:        []*Arg{},
					ResultType:  resultType,
				})
			}
		}
	}

	return schema, nil
}

func parseConstructor(firstLine string, scanner *bufio.Scanner) *Constructor {
	name, description, args, resultType, _ := parseEntity(firstLine, scanner)
	return &Constructor{
		Name:        name,
		Description: description,
		Args:        args,
		ResultType:  resultType,
	}
}

func parseFunction(firstLine string, scanner *bufio.Scanner) *Function {
	name, description, args, resultType, isSynchronous := parseEntity(firstLine, scanner)
	return &Function{
		Name:          name,
		Description:   description,
		Args:          args,
		ResultType:    resultType,
		IsSynchronous: isSynchronous,
		Type:          FUNCTION_TYPE_COMMON,
	}
}

func parseType(firstLine string, scanner *bufio.Scanner) *Type {
	typeLineParts := strings.Split(firstLine, "@")

	_, name := parseArg(typeLineParts[1])
	_, description := parseArg(typeLineParts[2])

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "//-") {
			break
		}
		description += " " + strings.TrimLeft(line, "//-")
	}

	return &Type{
		Name:        name,
		Description: description,
	}
}

func parseEntity(firstLine string, scanner *bufio.Scanner) (string, string, []*Arg, string, bool) {
	name := ""
	description := ""
	args := []*Arg{}
	resultType := ""

	argsLine := strings.TrimLeft(firstLine, "//")

Loop:
	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "//@"):
			argsLine += " " + strings.TrimLeft(line, "//")

		case strings.HasPrefix(line, "//-"):
			argsLine += " " + strings.TrimLeft(line, "//-")

		default:
			bodyFields := strings.Fields(line)
			name = bodyFields[0]

			for _, rawArg := range bodyFields[1 : len(bodyFields)-2] {
				argParts := strings.Split(rawArg, ":")
				arg := &Arg{
					Name:        argParts[0],
					Description: "",
					Type:        argParts[1],
				}
				args = append(args, arg)
			}
			resultType = strings.TrimRight(bodyFields[len(bodyFields)-1], ";")
			break Loop
		}
	}

	rawArgs := strings.Split(argsLine, "@")
	for _, rawArg := range rawArgs[1:] {
		name, value := parseArg(rawArg)
		switch {
		case name == "description":
			description = value
		default:
			name = strings.TrimPrefix(name, "param_")
			arg := getArg(args, name)
			arg.Description = value
		}
	}

	return name, description, args, resultType, strings.Contains(description, "Can be called synchronously")
}

func parseArg(str string) (string, string) {
	strParts := strings.Fields(str)

	return strParts[0], strings.Join(strParts[1:], " ")
}

func getArg(args []*Arg, name string) *Arg {
	for _, arg := range args {
		if arg.Name == name {
			return arg
		}
	}

	return nil
}
