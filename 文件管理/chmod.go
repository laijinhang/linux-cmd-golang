package main

import (
	"fmt"
	"os"
	"strings"
	"strconv"
	"io/ioutil"
	"errors"
)

var (
	u, g, o, a bool
	c, f, v, R bool
	help, version bool
)

func main() {
	for _, args := range os.Args[1:] {
		if args == "--help" {
			help = true
			break
		} else if args == "--version" {
			version = true
			break
		}
	}
	if help || len(os.Args) == 1 {
		fmt.Println(helpInfo)
		return
	} else if version {
		fmt.Println(versionInfo)
		return
	}
	// 解析出参数c, f, v, R，如果c在v后面出现过，则将v设为false，保留c
	for _, args := range os.Args {
		if args[0] == '-' {
			for _, char := range args[1:] {
				if char == 'c' {
					v = false
					c = true
				} else if char == 'v' {
					c = false
					v = true
				} else if char == 'f' {
					f = true
				} else if char == 'R' {
					R = true
				} else {
					fmt.Println("chmod: invalid mode: ’" + args + "’\nTry 'chmod --help' for more information.")
					return
				}
			}
		}
	}
	// 解析出命令a,u,a，如果不存在则找出数字
	modes := make([]string, 0)
	for _, args := range os.Args {
		if args[0] >= '0' && args[0] <= '9' && len(modes) == 0 {
			modes = append(modes, args)
			break
		} else if args[0] == 'a' || args[0] == 'g' || args[0] == 'o'{
			modes = append(modes, strings.Split(args, ",")...)
		}
	}
	// 文件列表
	files := make([]string, 0)
	for _, args := range os.Args[1:] {
		if args[0] != '-' {
			files = append(files, args)
			if R {	// 递归遍历文件
				fd, err := os.Stat(args)
				if err != nil {
					continue
				}
				if fd.IsDir() {
					files = append(files, scanDirs(fd.Name())...)
				}
			}
		}
	}

	// 去掉第一个权限设置
	files = files[1:]
	// 对每个文件进行文件权限设置操作
	for _, fileName := range files {
		fd, err := os.Open(fileName)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fh, err := fd.Stat()
		// 权限字符转成数字
		userMode, err := getModel(fh.Mode(), modes)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		um, _ := strconv.ParseInt(strconv.Itoa(userMode), 8 , 0)

		fd.Chmod(os.FileMode(um))
		if fh.Mode().String()[1:] == os.FileMode(um).String()[1:] {
			printData("mode of '" + fileName + "' retained as " + modeStringToOct(fh.Mode().String()) + " (" + fh.Mode().String() + ")", false)
		} else {
			printData("mode of '" + fileName + "' changed from " + modeStringToOct(fh.Mode().String()) + " (" + fh.Mode().String() + ") to 0" + strconv.Itoa(userMode) + " (" + os.FileMode(um).String() + ")", true)
		}
		fd.Close()
	}
}

// 递归扫描目录
func scanDirs(dirName string) []string {
	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		fmt.Println(err)
	}
	var fileList []string
	for _, file := range files {
		fileList = append(fileList, dirName + string(os.PathSeparator) + file.Name())
		if file.IsDir() {
			fileList = append(fileList, scanDirs(dirName + string(os.PathSeparator) + file.Name())...)
		}
	}
	return fileList
}

func getModel(old os.FileMode, modes []string) (int, error) {
	// 权限字符转成数字
	userMode := [3][4]int{}
	for i := 0;i < 3;i++ {
		if old.String()[1+i*3] == 'r' {
			userMode[i][0] = 1
		}
		if old.String()[2+i*3] == 'w' {
			userMode[i][1] = 1
		}
		if old.String()[3+i*3] == 'x'{
			userMode[i][2] = 1
		}
		if old.String()[3+i*3] == 's' {
			userMode[i][3] = 1
		}
	}
	um := -1
	for _, mode := range modes {
		if mode >= "0" && mode <= "777" {	// 相当于选择全部然后赋值
			var err error
			um, err = strconv.Atoi(mode)
			if err != nil {
				fmt.Println(err)
			}
			break
		}
		obj := [3]bool{}	// 0表示所属用户，1表示组，2表示其他用户
		ope := 0	// 0表示-，1表示+，2表示=
		modeVal := make(map[string]int)
		modeVal["r"] = 0
		modeVal["w"] = 1
		modeVal["x"] = 2
		modeVal["s"] = 3
		var modeU bool
		for i := 0;i < len(mode); {
			for ;mode[i] == 'a' || mode[i] == 'u' || mode[i] == 'g' || mode[i] == 'o';i++ {
				switch mode[0] {
				case 'a':
					obj = [3]bool{true, true, true}
				case 'u':
					obj[0] = true
				case 'g':
					obj[1] = true
				case 'o':
					obj[2] = true
				}
				modeU = true
			}
			for ;mode[i] == '-' || mode[i] == '+' || mode[i] == '=';i++ {
				switch mode[i] {
				case '-':
					ope = 0
				case '+':
					ope = 1
				case '=':
					ope = 2
					for j := 0;j < 3;j++ {
						if obj[j] {
							userMode[j] = [4]int{}
						}
					}
				}
			}
			for ;i < len(mode) && (mode[i] == 'r' || mode[i] == 'w' || mode[i] == 'x' || mode[i] == 's');i++ {
				for j := 0; j < 3; j++ {
					if obj[j] {
						if ope == 0 {
							userMode[j][modeVal[string(mode[i])]] = 0
						} else if ope == 1 {
							userMode[j][modeVal[string(mode[i])]] = 1
						} else if ope == 2 {
							userMode[j][modeVal[string(mode[i])]] = 1
						}
					}
				}
			}
			if i < len(mode) && mode[i] != '-' && mode[i] != '+' && mode[i] != '=' &&
				mode[i] != 'r' && mode[i] != 'w' && mode[i] != 'x' && mode[i] != 's' {
				err := "chmod: invalid mode: ‘" + mode + "\nTry 'chmod --help' for more information."
				return -1, errors.New(err)
			} else if i < len(mode) && !modeU && mode[i] != 'a' && mode[i] != 'u' && mode[i] != 'g' && mode[i] != 'o' {
				err := "chmod: invalid mode: ‘" + mode + "\nTry 'chmod --help' for more information."
				return -1, errors.New(err)
			}
		}
	}
	if um == -1 {
		um = ((userMode[0][0] & 1) * 4 | (userMode[0][1] & 1) * 2 | userMode[0][2]) * 100
		um += ((userMode[1][0] & 1) * 4 | (userMode[1][1] & 1) * 2 | userMode[1][2]) * 10
		um += ((userMode[2][0] & 1) * 4 | (userMode[2][1] & 1) * 2 | userMode[2][2])
	}
	return um, nil
}

func printData(data string, change bool) {
	if v || (c && change) {
		fmt.Println(data)
	}
}

func modeStringToOct(mode string) string {
	oct := "0"
	for i := 0;i < 3;i++ {
		temp := 0
		if mode[1+3*i] == 'r' {
			temp = 4
		}
		if mode[2+3*i] == 'w' {
			temp += 2
		}
		if mode[3+3*i] == 'x' {
			temp += 1
		}
		oct += strconv.Itoa(temp)
	}
	return oct
}

const (
	helpInfo = `Usage: chmod [OPTION]... MODE[,MODE]... FILE...
  or:  chmod [OPTION]... OCTAL-MODE FILE...
  or:  chmod [OPTION]... --reference=RFILE FILE...
Change the mode of each FILE to MODE.
With --reference, change the mode of each FILE to that of RFILE.

  -c, --changes          like verbose but report only when a change is made
  -f, --silent, --quiet  suppress most error messages
  -v, --verbose          output a diagnostic for every file processed
      --no-preserve-root  do not treat '/' specially (the default)
      --preserve-root    fail to operate recursively on '/'
      --reference=RFILE  use RFILE's mode instead of MODE values
  -R, --recursive        change files and directories recursively
      --help     display this help and exit
      --version  output version information and exit

Each MODE is of the form '[ugoa]*([-+=]([rwxXst]*|[ugo]))+|[-+=][0-7]+'.

GNU coreutils online help: <http://www.gnu.org/software/coreutils/>
Full documentation at: <http://www.gnu.org/software/coreutils/chmod>
or available locally via: info '(coreutils) chmod invocation'`
	versionInfo = 	`chmod (GNU coreutils) 8.25
Copyright (C) 2016 Free Software Foundation, Inc.
License GPLv3+: GNU GPL version 3 or later <http://gnu.org/licenses/gpl.html>.
This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.

Written by David MacKenzie and Jim Meyering.`
)
