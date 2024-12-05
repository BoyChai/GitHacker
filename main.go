package main

import (
	"GitHacker/recovery"
	"flag"
	"fmt"
	"strings"
)

func main() {
	t := flag.String("t", "url", "指定类型,默认为url")
	out := flag.String("o", "", "输出目录,默认值为当前位置的GitHacker_Output的主机地址目录下")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("传入参数有误，请查看使用说明")
		return
	}

	if *out == "" {
		withoutProtocol := strings.Split(args[0], "://")[1]
		recovery.OutputDir = "GitHacker_Output/" + strings.Split(withoutProtocol, "/")[0]
	} else {
		recovery.OutputDir = *out
	}

	switch *t {
	case "url":
		recovery.UrlRecovery(args[0])
	case "local":
		recovery.LocalRecovery(args[0])
	default:
		fmt.Println("传入类型有误，请查看使用说明")
		return
	}
}
