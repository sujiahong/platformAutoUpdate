package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const rootPath = "F:\\u"
const limitTime = 2 * 86400

func walkCallback(ph string, info os.FileInfo, err error) error {
	//fmt.Println(ph, info.Name(), info.IsDir(), err)
	if err == nil {
		if !info.IsDir() {
			fmt.Println(info.Name(), info.ModTime())
			mtime := info.ModTime()
			nowTime := time.Now()
			dTime := nowTime.Unix() - mtime.Unix()
			fmt.Println(mtime.Unix(), nowTime.Unix(), dTime, limitTime)
			if dTime >= limitTime {
				rerr := os.Remove(ph)
				fmt.Println("查看删除错误  ", rerr)
			}
		}
	}
	return err
}

func main() {
	for {
		fmt.Println("自动清除日志 ")
		err := filepath.Walk(rootPath, walkCallback)
		fmt.Println("Walk err: ", err)
		time.Sleep(60 * time.Second)
	}
}
