package main

import (
	"GitHacker/recovery"
	"flag"
	"fmt"
)

func main() {
	t := flag.String("t", "url", "指定类型,默认为url")

	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("传入参数有误，请查看使用说明")
		return
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
