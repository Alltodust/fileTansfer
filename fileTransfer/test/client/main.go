package main

import (
	"fmt"
	"interviewTest/test/client/bar"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

//获取文件上次发送的位置
func clientRead(conn net.Conn) int {
	buf := make([]byte, 5)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatalf("receive server info faild: %s\n", err)
	}
	// string conver int
	off, err := strconv.Atoi(string(buf[:n]))
	if err != nil {
		log.Fatalf("string conver int faild: %s\n", err)
	}
	return off
}

func sendFile(con net.Conn, filePath string) {

	//获取文件上次发送的位置
	_, err := con.Write([]byte("start-->"))
	off := clientRead(con)

	//打开文件
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Println("os.Open err: ", err)
		return
	}
	defer f.Close()

	//设置从文件的哪里开始发送数据
	_, err = f.Seek(int64(off), 0)
	if err != nil {
		fmt.Println("set file seek failed:", err)
	}
	//初始化进度条
	fileInfo, err := os.Stat(filePath)
	b := bar.NewBar(0, int(fileInfo.Size()))
	//持续发送
	for {
		data := make([]byte, 128)
		n, err := f.Read(data)
		if err != nil {
			if err == io.EOF {
				// 如果已经读取完文件内容 ，就发送'<--end'消息通知服务端，文件内容发送完了
				time.Sleep(time.Second * 1)
				_, err = con.Write([]byte("<--end"))
				if err != nil {
					fmt.Println("con.Write err:", err)
				}
				break
			}
		}
		_, err = con.Write(data[:n])
		if err != nil {
			fmt.Println("send content failed: ", err)
		}

		//进度条
		for i := 0; i < n; i++ {
			b.Add(1)
			time.Sleep(time.Millisecond)
		}
	}
}

func clientConn(con net.Conn, filePath string) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		fmt.Println("os.Stat err ", err)
		return
	}
	fileName := fileInfo.Name()

	//发送文件名，接收返回的ok，正式发送文件数据
	_, err = con.Write([]byte(fileName))
	if err != nil {
		fmt.Println("con.Write ", err)
		return
	}
	buf := make([]byte, 16)
	n, err := con.Read(buf)
	if err != nil {
		fmt.Println("con.read err ", err)
		return
	}
	if "ok" == string(buf[:n]) {
		sendFile(con, filePath)
	}
}

func main() {
	//建立连接
	con, err := net.Dial("tcp", "127.0.0.1:8000")
	if err != nil {
		fmt.Println("net.Dial ", err)
		return
	}
	defer con.Close()

	//获取命令行参数
	list := os.Args
	if len(list) != 2 {
		fmt.Println("格式为：go run xxx.go 文件绝对路径")
		return
	}
	//获取文件路径
	filePath := list[1]

	//获取文件信息
	files, err := ioutil.ReadDir(filePath)
	if err != nil {
		log.Fatal("ioutil.ReadDir: ", err)
	}
	len := len(files)

	//WaitGroup实现并发同步
	gw := sync.WaitGroup{}
	gw.Add(len)
	//开一个goroutine并发处理
	go func() {
		for _, file := range files {
			fmt.Println(file.Name())
			clientConn(con, filePath+"\\"+file.Name())
			fmt.Println()
			gw.Done()
		}
	}()
	//一个file安排一个goroutine去处理
	//for _, file := range files {
	//	go func(file fs.FileInfo) {
	//		fmt.Println(file.Name())
	//		clientConn(con, filePath+"\\"+file.Name())
	//		fmt.Println()
	//		gw.Done()
	//	}(file)
	//}
	gw.Wait()

}

//cd E:\GolangProject\interviewTest\test\client
//go run main.go E:\GolangProject\home

//cd E:\GolangProject\interviewTest\test\server
