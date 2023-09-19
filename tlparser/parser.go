package tlparser

import (
	"bufio"
	"io"
	"strings"
)

func Parse(reader io.Reader) (*Schema, error) {
	schema := &Schema{
		Types:     []*Type{},
		Classes:   []*Class{},
		Functions: []*Function{},
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
				schema.Types = append(schema.Types, parseType(line, scanner))
			}

		case strings.HasPrefix(line, "//@class"):
			schema.Classes = append(schema.Classes, parseClass(line, scanner))

		case strings.Contains(line, "---functions---"):
			hitFunctions = true

		case line == "":

		default:
			bodyFields := strings.Fields(line)
			name := bodyFields[0]
			class := strings.TrimRight(bodyFields[len(bodyFields)-1], ";")
			if hitFunctions {
				schema.Functions = append(schema.Functions, &Function{
					Name:          name,
					Description:   "",
					Class:         class,
					Properties:    []*Property{},
					IsSynchronous: false,
					Type:          FUNCTION_TYPE_UNKNOWN,
				})
			} else {
				if name == "vector" {
					name = "vector<t>"
					class = "Vector<T>"
				}

				schema.Types = append(schema.Types, &Type{
					Name:        name,
					Description: "",
					Class:       class,
					Properties:  []*Property{},
				})
			}
		}
	}

	return schema, nil
}

func parseType(firstLine string, scanner *bufio.Scanner) *Type {
	name, description, class, properties, _ := parseEntity(firstLine, scanner)
	return &Type{
		Name:        name,
		Description: description,
		Class:       class,
		Properties:  properties,
	}
}

func parseFunction(firstLine string, scanner *bufio.Scanner) *Function {
	name, description, class, properties, isSynchronous := parseEntity(firstLine, scanner)
	return &Function{
		Name:          name,
		Description:   description,
		Class:         class,
		Properties:    properties,
		IsSynchronous: isSynchronous,
		Type:          FUNCTION_TYPE_UNKNOWN,
	}
}

func parseClass(firstLine string, scanner *bufio.Scanner) *Class {
	classLineParts := strings.Split(firstLine, "@")

	_, name := parseProperty(classLineParts[1])
	_, description := parseProperty(classLineParts[2])

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "//-") {
			break
		}
		description += " " + strings.TrimLeft(line, "//-")
	}

	return &Class{
		Name:        name,
		Description: description,
	}
}

func parseEntity(firstLine string, scanner *bufio.Scanner) (string, string, string, []*Property, bool) {
	name := ""
	description := ""
	class := ""
	properties := []*Property{}

	propertiesLine := strings.TrimLeft(firstLine, "//")

Loop:
	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "//@"):
			propertiesLine += " " + strings.TrimLeft(line, "//")

		case strings.HasPrefix(line, "//-"):
			propertiesLine += " " + strings.TrimLeft(line, "//-")

		default:
			bodyFields := strings.Fields(line)
			name = bodyFields[0]

			for _, rawProperty := range bodyFields[1 : len(bodyFields)-2] {
				propertyParts := strings.Split(rawProperty, ":")
				property := &Property{
					Name: propertyParts[0],
					Type: propertyParts[1],
				}
				properties = append(properties, property)
			}
			class = strings.TrimRight(bodyFields[len(bodyFields)-1], ";")
			break Loop
		}
	}

	rawProperties := strings.Split(propertiesLine, "@")
	for _, rawProperty := range rawProperties[1:] {
		name, value := parseProperty(rawProperty)
		switch {
		case name == "description":
			description = value
		default:
			name = strings.TrimPrefix(name, "param_")
			property := getProperty(properties, name)
			property.Description = value

		}
	}

	return name, description, class, properties, strings.Contains(description, "Can be called synchronously")
}

func parseProperty(str string) (string, string) {
	strParts := strings.Fields(str)

	return strParts[0], strings.Join(strParts[1:], " ")
}

func getProperty(properties []*Property, name string) *Property {
	for _, property := range properties {
		if property.Name == name {
			return property
		}
	}

	return nil
}
