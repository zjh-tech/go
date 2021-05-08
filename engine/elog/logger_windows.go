// +build windows
package elog

import "fmt"

func (l *Logger) out_put_console(content string) {
	fmt.Print(content)
}
