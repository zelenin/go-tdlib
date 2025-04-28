package tlparser

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

func getMethodName(line string, re *regexp.Regexp) string {
	return strings.TrimSpace(strings.TrimPrefix(re.FindString(line), "td_api::"))
}

func ParseCode(reader io.Reader, schema *Schema) error {

	reExtractMethodName, err := regexp.Compile(`td_api::(.*?) `)
	if err != nil {
		return fmt.Errorf("invalid regexp to extract method name: %w", err)
	}

	userMethods := map[string]bool{}
	botMethods := map[string]bool{}

	scanner := bufio.NewScanner(reader)

	methodName := ""

	for scanner.Scan() {

		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "void Requests::on_request("):
			methodName = getMethodName(line, reExtractMethodName)
		case strings.Contains(line, "CHECK_IS_USER();"):
			userMethods[methodName] = true
		case strings.Contains(line, "CHECK_IS_BOT();"):
			botMethods[methodName] = true
		case line == "}":
			methodName = ""
		}

	}

	err = scanner.Err()
	if err != nil {
		return err
	}

	for _, fn := range schema.Functions {

		fn.Type = FUNCTION_TYPE_COMMON

		switch {
		case userMethods[fn.Name]:
			fn.Type = FUNCTION_TYPE_USER
		case botMethods[fn.Name]:
			fn.Type = FUNCTION_TYPE_BOT
		}

	}

	return nil

}
