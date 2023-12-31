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
	pattern := fmt.Sprintf(`\%s\((.*)\)\[(.*)\]`, variable)
	varMatches, err := FindStringSubmatch(varStr, pattern)
	if err != nil {
		fmt.Printf("* err: Failed to match: %s.\n", err.Error())
		return "", "", err
	}
	if len(varMatches) == 3 {
		group, err := strconv.Atoi(varMatches[2])
		if err != nil {
			fmt.Printf("* err: %s is not an int number, %s.\n", varMatches[2], err.Error())
		}
		matches, err := FindStringSubmatch(str, varMatches[1])
		if err != nil {
			fmt.Printf("* err: Failed to match: %s.\n", err.Error())
		}
		if len(matches) > group {
			return varMatches[0], matches[group], nil
		} else {
			fmt.Printf("* err: Group is out of range, index: %d, size: %d.\n", group, len(matches))
			fmt.Printf("* err: Current capturing groups: %s.\n", strings.Join(matches, ", "))
			return "", "", fmt.Errorf("")
		}
	} else {
		fmt.Printf("* err: Only support two capturing groups: %s.\n", pattern)
		fmt.Printf("* err: Current capturing groups: %s.\n", strings.Join(varMatches, ", "))
		return "", "", fmt.Errorf("")
	}
}
