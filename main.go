package main

import (
	"bufio"
	"matrix-seeker/cmd"
	"os"
	"runtime"
	"strings"
)

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	app := cmd.InitCli()
	//监控用户输入
	for {
		var input string

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input = scanner.Text()

		//构建命令
		s := []string{app.Name}

		//获取命令
		cmdArgs := strings.Split(input, " ")
		if len(cmdArgs) == 0 {
			continue
		}

		s = append(s, cmdArgs...)
		app.Run(s)
	}
}
