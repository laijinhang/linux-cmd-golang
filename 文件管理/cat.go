package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	parse()
}

func parse() {
	var help = flag.Bool("help", false, "")
	var n = flag.Bool("n", false, "")
	var number = flag.Bool("number", false, "")
	var b = flag.Bool("b", false, "")
	var numbernonblank = flag.Bool("number-nonblank", false, "")
	var s = flag.Bool("s", false, "")
	var sqb = flag.Bool("squeeze-blank", false, "")
	var v = flag.Bool("v", false, "")
	var sn = flag.Bool("show-nonprinting", false, "")
	var e = flag.Bool("e", false, "")
	var E = flag.Bool("E", false, "")
	var se = flag.Bool("show-ends", false, "")
	var t = flag.Bool("t", false, "")
	var T = flag.Bool("T", false, "")
	var st = flag.Bool("show-tabs", false, "")
	var A = flag.Bool("A", false, "")

	flag.Parse()

	if *n {
		*number = *n
	}
	if *b {
		*numbernonblank = *b
	}
	if *s {
		*sqb = *s
	}
	if *v {
		*sn = *v
	}
	if *e {
		*se = *e
	}
	if *E {
		*se = *E
	}
	if *t {*st = *t}
	if *T {*st = *T}
	if *A {
		*sn = *A
		*se = *A
		*st = *A
	}

	args := os.Args

	if len(args) == 1 {
		reader := bufio.NewReader(os.Stdin)
		for {
			data, _, _ := reader.ReadLine()
			fmt.Println(string(data))
		}
	}
	if *help {
		helpFun()
		return
	}
	args = args[1:]
	i := 0
	for _, fn := range args {
		if fn[0] != '-' {
			fd, err := os.Open(fn)
			if err != nil {
				fmt.Println(err)
				continue
			}
			input := bufio.NewScanner(fd)
			var space bool
			if *numbernonblank {
				for input.Scan() {
					temp := input.Text()
					if temp == "" {
						if !space {
							if *sqb {
								space = true
							}
							if *se {
								fmt.Print("$")
							}
							fmt.Println()
						}
					} else {
						space = false
						i++
						if *st {
							temp = strings.Replace(temp, "\t", "^|", -1)
						}
						if *se {
							fmt.Printf("%6d  %s$\n", i, temp)
						} else {
							fmt.Printf("%6d  %s\n", i, temp)
						}
					}
				}
			} else if *number {
				for input.Scan() {
					temp := input.Text()
					if *sqb {
						if temp == "" && space {
							continue
						} else if temp == "" && !space {
							space = true
						} else {
							space = false
						}
					}
					i++
					if *st {
						temp = strings.Replace(temp, "\t", "^|", -1)
					}
					if *se {
						fmt.Printf("%6d  %s$\n", i, input.Text())
					} else {
						fmt.Printf("%6d  %s\n", i, input.Text())
					}
				}
			} else {
				for input.Scan() {
					temp := input.Text()
					if *st {
						temp = strings.Replace(temp, "\t", "^|", -1)
					}
					fmt.Print(temp)
					if *se {
						fmt.Print("$")
					}
					fmt.Println()
				}
			}
			fd.Close()
		}
	}

	return
}

func helpFun() {
	fmt.Println(`
Usage: cat [OPTION]... [FILE]...
Concatenate FILE(s) to standard output.

With no FILE, or when FILE is -, read standard input.

  -A, --show-all           equivalent to -vET
  -b, --number-nonblank    number nonempty output lines, overrides -n
  -e                       equivalent to -vE
  -E, --show-ends          display $ at end of each line
  -n, --number             number all output lines
  -s, --squeeze-blank      suppress repeated empty output lines
  -t                       equivalent to -vT
  -T, --show-tabs          display TAB characters as ^I
  -u                       (ignored)
  -v, --show-nonprinting   use ^ and M- notation, except for LFD and TAB
      --help     display this help and exit
      --version  output version information and exit

Examples:
  cat f - g  Output f's contents, then standard input, then g's contents.
  cat        Copy standard input to standard output.

GNU coreutils online help: <http://www.gnu.org/software/coreutils/>
Full documentation at: <http://www.gnu.org/software/coreutils/cat>
or available locally via: info '(coreutils) cat invocation'`)
}
