package tlparser

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func ParseCode(reader io.Reader, schema *Schema) error {
	var prevLine string
	var curLine string

	userMethods := map[string]bool{}
	botMethods := map[string]bool{}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		prevLine = curLine
		curLine = scanner.Text()

		if strings.Contains(curLine, "CHECK_IS_USER();") {
			fields := strings.Fields(prevLine)
			for _, field := range fields {
				var methodName string
				n, err := fmt.Sscanf(field, "td_api::%s", &methodName)
				if err == nil && n > 0 {
					userMethods[methodName] = true
				}
			}
		}

		if strings.Contains(curLine, "CHECK_IS_BOT();") {
			fields := strings.Fields(prevLine)
			for _, field := range fields {
				var methodName string
				n, err := fmt.Sscanf(field, "td_api::%s", &methodName)
				if err == nil && n > 0 {
					botMethods[methodName] = true
				}
			}
		}
	}

	err := scanner.Err()
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
