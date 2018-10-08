package main

import (
	"os"
	"fmt"
	"strings"
	"os/user"
	"syscall"
	"io/ioutil"
	"strconv"
)

const (
	HELPINFO = ""
	VERSIONINFO = ""
)

var (
	userName string		// 新文件拥有者的使用者ID
	group string	// 新文件拥有者的使用者组(group)
	g bool

	c bool	// 显示更改的部分的信息
	f bool	// 忽略错误信息
	h bool	// 修复符号链接
	v bool	// 显示详细的处理信息
	R bool	// 处理指定目录以及其子目录下的所有文件
	help bool		// 显示辅助说明
	version bool	// 显示版本
)

func main() {
	args := os.Args[1:]
	files := []string{}

	// 参数处理
	for _, arg := range args {
		if arg == "--help" {
			fmt.Println(HELPINFO)
			return
		} else if arg == "--version" {
			fmt.Println(VERSIONINFO)
			return
		} else if arg[0] == '-' {	// 参数处理
			for _, char := range arg[1:] {
				if char != 'c' && char != 'f' &&
					char != 'h' && char != 'v' &&
					char != 'R' {
					fmt.Println("chown: invalid option -- '" + string(char) + "'\nTry 'chown --help' for more information.")
					return
				}
				switch char {
				case 'c':
					c = true
					v = false
				case 'f':
					f = true
				case 'h':
					h = true
				case 'v':
					c = false
					v = true
				case 'R':
					R = true
				}
			}
		} else if userName == "" {
			u := strings.Split(arg, ":")
			userName = u[0]
			if len(u) > 1 {
				g = true
				group = u[1]
			}
		} else {	// 剩下的就是文件了
			files = append(files, arg)
		}
	}
	// 获取用户id和组id
	userInfo, err := user.Lookup(userName)
	if err != nil {
		fmt.Println("chown: invalid user: " + userName + "’")
		return
	}
	var groupInfo *user.Group
	if g {
		var err error
		if group == "" {
			groupInfo, err = user.LookupGroupId(userInfo.Gid)
			if err != nil {
				fmt.Println(err)
				return
			}
		} else {
			groupInfo, err = user.LookupGroup(group)
			if err != nil {
				fmt.Println("chown: invalid group: " + userInfo.Name + ":" + group)
				return
			}
		}
	}

	//ownership of 'test.go' retained as lai
	//changed ownership of 'test.go' from lai to root
	//changed ownership of 'test.go' from root:root to lai:lai

	if R { // 递归处理目录
		for _, file := range files {
			files = append(files, scanDirs(file)...)
		}
	}

	for _, file := range files {
		fi, err := os.Stat(file)
		if err != nil {
			if os.IsPermission(err) {
				fmt.Println("chown: cannot read directory '" + file + "': Permission denied\nfailed to change ownership of '" + file + "' to " + userInfo.Name + ":" + groupInfo.Name)
			} else if os.IsNotExist(err) {
				fmt.Println("chown: cannot access '" + file + "': No such file or directory\nfailed to change ownership of '" + file + "' to " + userInfo.Name + ":" + groupInfo.Name)
			}
			continue
		}
		stat_t := fi.Sys().(*syscall.Stat_t)
		uid, err := strconv.Atoi(userInfo.Uid)
		if err != nil {
			fmt.Println(err)
			continue
		}
		var gid int
		if g {
			gid, err = strconv.Atoi(groupInfo.Gid)
		} else {
			gid = int(stat_t.Gid)
		}
		if err != nil {
			fmt.Println(err)
			continue
		}
		err = os.Chown(file, uid, gid)
		if err != nil {
			if os.IsPermission(err) {
				fmt.Println("chown: cannot read directory '" + file + "': Permission denied\nfailed to change ownership of '" + file + "' to " + userInfo.Name + ":" + groupInfo.Name)
			} else if os.IsNotExist(err) {
				fmt.Println("chown: cannot access '" + file + "': No such file or directory\nfailed to change ownership of '" + file + "' to " + userInfo.Name + ":" + groupInfo.Name)
			}
			continue
		}
		if !g {	// 不修改组打印
			if stat_t.Uid == uint32(uid) && stat_t.Gid == uint32(gid) {
				fmt.Println("ownership of '" + file + "' retained as " + userInfo.Name)
			} else {
				// 通过uid得到name
				oldUserInfo, err := user.LookupId(strconv.Itoa(int(stat_t.Uid)))
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println("changed ownership of '" + file + "' from " + oldUserInfo.Name + " to " + userInfo.Name)
			}
		} else {	// 修改组打印
			if stat_t.Uid == uint32(uid) && stat_t.Gid == uint32(gid) {
				fmt.Println("ownership of '" + file + "' retained as " + userInfo.Name + ":" + groupInfo.Name)
			} else {
				// 通过uid得到name
				oldUserInfo, err := user.LookupId(strconv.Itoa(int(stat_t.Uid)))
				if err != nil {
					if os.IsPermission(err) {
						fmt.Println("chown: cannot read directory '" + file + "': Permission denied\nfailed to change ownership of '" + file + "' to " + userInfo.Name + ":" + groupInfo.Name)
					} else if os.IsNotExist(err) {
						fmt.Println("chown: cannot access '" + file + "': No such file or directory\nfailed to change ownership of '" + file + "' to " + userInfo.Name + ":" + groupInfo.Name)
					}
					continue
				}
				oldGroupInfo, err := user.LookupGroupId(strconv.Itoa(int(stat_t.Gid)))
				if err != nil {
					if os.IsPermission(err) {
						fmt.Println("chown: cannot read directory '" + file + "': Permission denied\nfailed to change ownership of '" + file + "' to " + userInfo.Name + ":" + groupInfo.Name)
					} else if os.IsNotExist(err) {
						fmt.Println("chown: cannot access '" + file + "': No such file or directory\nfailed to change ownership of '" + file + "' to " + userInfo.Name + ":" + groupInfo.Name)
					}
					continue
				}
				fmt.Println("changed ownership of '" + file + "' from " + oldUserInfo.Name + ":" + oldGroupInfo.Name + " to " + userInfo.Name + ":" + groupInfo.Name)
			}
		}
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