// +build windows
package elog

import "fmt"

func (l *Logger) outPutConsole(content string) {
	fmt.Print(content)
}
