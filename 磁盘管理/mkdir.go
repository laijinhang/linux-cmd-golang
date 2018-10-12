package main

import (
	"os"
	"fmt"
	"strings"
)

func main() {
	var p bool
	args := os.Args[1:]
	dirs := []string{}
	// 解析参数
	for _, arg := range args {
		if arg[0] != '-' {	// 去掉参数
			dirs = append(dirs, strings.TrimRight(arg, "/"))	//如果最右边一个字符是/，则将其删除
		} else if arg == "--help" {
			fmt.Printf("%s\n", help)
		} else if arg == "--version" {
			fmt.Printf("%s\n", version)
		} else if arg == "-p" || arg == "-P" {
			p = true
		} else {
			fmt.Printf("mkdir: invalid option -- '%s'\nTry 'mkdir --help' for more information.\n",
				string(arg[0]))
			return
		}
	}
	// 开始创建目录
	var err error
	for _, dir := range dirs {
		_, err = os.Stat(dir)
		if err == nil {
			fmt.Printf("mkdir: cannot create directory ‘%s’: File exists\n", dir)
			continue
		}
		if p {	// 递归创建
			err = os.MkdirAll(dir, 0777)
		} else {
			err = os.Mkdir(dir, 0777)
		}
		if err != nil {
			if os.IsPermission(err) {
				fmt.Printf("mkdir: cannot create directory '%s’: Permission denied\n", dir)
				continue
			} else if os.IsNotExist(err) {
				fmt.Printf("mkdir: cannot create directory ‘%s': No such file or directory\n", dir)
				continue
			}
		}
	}
}

const (
	help = `Usage: mkdir [OPTION]... DIRECTORY...
Create the DIRECTORY(ies), if they do not already exist.

Mandatory arguments to long options are mandatory for short options too.
  -m, --mode=MODE   set file mode (as in chmod), not a=rwx - umask
  -p, --parents     no error if existing, make parent directories as needed
  -v, --verbose     print a message for each created directory
  -Z                   set SELinux security context of each created directory
                         to the default type
      --context[=CTX]  like -Z, or if CTX is specified then set the SELinux
                         or SMACK security context to CTX
      --help     display this help and exit
      --version  output version information and exit

GNU coreutils online help: <http://www.gnu.org/software/coreutils/>
Full documentation at: <http://www.gnu.org/software/coreutils/mkdir>
or available locally via: info '(coreutils) mkdir invocation'`
	version = `mkdir (GNU coreutils) 8.25
Copyright (C) 2016 Free Software Foundation, Inc.
License GPLv3+: GNU GPL version 3 or later <http://gnu.org/licenses/gpl.html>.
This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.

Written by David MacKenzie.`
)
