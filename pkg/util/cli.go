package util

import (
	"os"
	"strings"
)

var cliUtilArgs map[string]string
var cliUtilFlags map[string]string

func GetCLIArgs() map[string]string {
	args := os.Args
	if len(args) < 2 {
		return nil
	}

	if len(cliUtilArgs) < 1 {
		cliUtilArgs = make(map[string]string)
	} else {
		return cliUtilArgs
	}

	args = args[1:]
	for _, val := range args {
		valSplit := strings.Split(val, "=")
		if len(valSplit) < 2 {
			valSplit = append(valSplit, "")
		}
		cliUtilArgs[valSplit[0]] = valSplit[1]
	}

	return cliUtilArgs
}

func GetCLIArg(key string) string {
	cliArgs := GetCLIArgs()
	if val, ok := cliArgs[key]; ok {
		return val
	}
	return ""
}

func HasCLIArgs(hasArgs ...string) bool {
	cliArgs := GetCLIArgs()
	for _, arg := range hasArgs {
		if _, ok := cliArgs[arg]; ok {
			return true
		}
	}
	return false
}

func GetCLIFlags() map[string]string {
	cliArgs := GetCLIArgs()

	if len(cliUtilFlags) < 1 {
		cliUtilFlags = make(map[string]string)
	} else {
		return cliUtilFlags
	}

	for key, val := range cliArgs {
		key = strings.TrimPrefix(key, "--")
		key = strings.TrimPrefix(key, "-")
		cliUtilFlags[key] = val
	}
	return cliUtilFlags
}

func GetCLIFlag(key string) string {
	cliFlags := GetCLIFlags()
	if val, ok := cliFlags[key]; ok {
		return val
	}
	return ""
}

func HasCLIFlags(hasFlags ...string) bool {
	cliFlags := GetCLIFlags()
	for _, flag := range hasFlags {
		if _, ok := cliFlags[flag]; ok {
			return true
		}
	}
	return false
}
