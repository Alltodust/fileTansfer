package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
)

//存放文件路径
const filePath = "E:\\GolangProject\\interviewTest\\tmp\\"

// 把接收到的内容append到文件
func writeFile(content []byte, fileName string) {
	if len(content) != 0 {
		f, err := os.OpenFile(filePath+fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0755)
		defer f.Close()
		if err != nil {
			fmt.Println("open file failed:", err)
		}
		_, err = f.Write(content)
		if err != nil {
			fmt.Println("append content to failed:", err)
		}
	}
}

// 获取已接收内容的大小
// 断点续传需要把已接收内容大下通知客户端从哪里开始发送文件内容
func getFileStat(fileName string) int64 {
	fileinfo, err := os.Stat(filePath + fileName)
	if err != nil {
		fmt.Println("get file stat failed: ", err)
	}
	return fileinfo.Size()

}

func recvFile(con net.Conn, fileName string) {

	// 在指定文件路径下创建文件并打开了文件
	f, err := os.Create(filePath + fileName)
	if err != nil {
		fmt.Println("os.Create err:", err)
		return
	}
	defer f.Close()

	for {
		var buf = make([]byte, 10)
		n, err := con.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("server io EOF")
				return
			}
			fmt.Println("server read failed: ", err)
		}
		switch string(buf[:n]) {
		case "start-->":
			off := getFileStat(fileName)
			stringoff := strconv.FormatInt(off, 10)
			_, err = con.Write([]byte(stringoff))
			if err != nil {
				fmt.Println("server write failed:", err)
			}
			continue
		case "<--end":
			return
		}
		writeFile(buf[:n], fileName)
	}
}

func main() {
	l, err := net.Listen("tcp", "127.0.0.1:8000")
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	defer l.Close()
	fmt.Println("接收端启动成功，等待发送端发送文件！")

	//阻塞监听
	con, err := l.Accept()
	if err != nil {
		fmt.Println("l.Accept err:", err)
		return
	}
	defer con.Close()

	//循环接收多个文件数据
	for {
		buf := make([]byte, 4096)
		n, err := con.Read(buf)
		//如果接收完对方发的数据，会返回 EOF，结束协程
		if err != nil {
			if err == io.EOF {
				// 如果已经接收完文件内容，退出循环
				return
			}
			fmt.Println("con.Read err:", err)
			return
		}
		fileName := string(buf[:n])
		con.Write([]byte("ok"))
		recvFile(con, fileName)
		//go func() {
		//	recvFile(con, fileName)
		//}()
	}
	fmt.Println("recevice over")
}
