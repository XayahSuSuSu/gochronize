package util

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func MatchString(str, pattern string) (bool, error) {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}
	if regex.MatchString(str) {
		return true, nil
	} else {
		return false, fmt.Errorf("regex not matched")
	}
}

func FindStringSubmatch(str, pattern string) ([]string, error) {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return []string{}, err
	}
	return regex.FindStringSubmatch(str), nil
}

func MatchCustomRegex(variable, varStr, str string) (string, string, error) {
	variableFlag := strings.TrimSuffix(variable, "}")
	if strings.Contains(varStr, variableFlag) {
		pattern := fmt.Sprintf(`\%s\((.*)\)\[(.*)\]}`, variableFlag)
		varMatches, err := FindStringSubmatch(varStr, pattern)
		if err != nil {
			Fprintfln("* err: Failed to match: %v", err)
			return "", "", err
		}
		if len(varMatches) == 3 {
			group, err := strconv.Atoi(varMatches[2])
			if err != nil {
				Fprintfln("* err: %s is not an int number, %v", varMatches[2], err)
			}
			matches, err := FindStringSubmatch(str, varMatches[1])
			if err != nil {
				Fprintfln("* err: Failed to match: %v", err)
			}
			if len(matches) > group {
				return varMatches[0], matches[group], nil
			} else {
				Fprintfln("* err: Group is out of range, index: %d, size: %d.", group, len(matches))
				Fprintfln("* err: Current capturing groups: %s.", strings.Join(matches, ", "))
				return "", "", fmt.Errorf("")
			}
		} else {
			Fprintfln("* err: Only support two capturing groups: %s.", pattern)
			Fprintfln("* err: Current capturing groups: %s.", strings.Join(varMatches, ", "))
			return "", "", fmt.Errorf("")
		}
	} else {
		return "", "", fmt.Errorf("")
	}
}
