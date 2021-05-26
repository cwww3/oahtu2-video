package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"sync"
)

// oauth.assets
var path = "https://cdn.jsdelivr.net/gh/cwww3/picture@master/img"
var wg sync.WaitGroup

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("lack of config file")
		os.Exit(-1)
	}
	// 不支持零宽断言
	//r, err := regexp.Compile("(?<=\\()(\\S+)(?=\\/)")
	r, err := regexp.Compile("(?<=\\()(\\S+)(?=\\/)")
	if err != nil {
		fmt.Printf("compile fail err=%v", err)
		os.Exit(-1)
	}
	for i := 1; i < len(os.Args); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fileName := os.Args[i]
			in, err := os.Open(fileName)
			if err != nil {
				fmt.Printf("open %v fail:%v\n", fileName, err)
				os.Exit(-1)
			}
			defer in.Close()

			out, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0766)
			if err != nil {
				fmt.Printf("Open write %v fail:%v\n", fileName, err)
				os.Exit(-1)
			}
			defer out.Close()

			br := bufio.NewReader(in)
			index := 1
			for {
				line, _, err := br.ReadLine()
				if err == io.EOF {
					break
				}
				if err != nil {
					fmt.Printf("read %v err:%v\n", fileName, err)
					os.Exit(-1)
				}
				newLine := r.ReplaceAllString(string(line), path)
				fmt.Println(newLine)
				//newLine := strings.Replace(string(line), os.Args[2], os.Args[3], -1)

				_, err = out.WriteString(newLine + "\n")
				if err != nil {
					fmt.Println("write to file fail:", err)
					os.Exit(-1)
				}
				index++
			}
			fmt.Printf("%v---%v\n", fileName, index)
		}()
	}
	fmt.Println("等待结束")
	wg.Wait()
}
