package main

import (
	"fmt"
	"github.com/arloor/proxygo/pac"
	"github.com/arloor/proxygo/util"
	"net"
)

var localAddr = ":8081"
var proxyAddr = "proxy:8080"

func main() {
	go pac.ServePAC()
	printLocalIPs()

	ln, err := net.Listen("tcp", localAddr)
	if err != nil {
		fmt.Println("监听", localAddr, "失败 ", err)
		return
	}
	defer ln.Close()
	fmt.Println("成功监听 ", ln.Addr())
	for {
		c, err := ln.Accept()
		if err != nil {
			fmt.Println("接受连接失败 ", err)
		} else {
			go handleBrowserConnnection(c)
		}
	}
}

func handleBrowserConnnection(localConn net.Conn) {
	var proxyConn, err = net.Dial("tcp", proxyAddr)
	if err != nil {
		fmt.Println("连接到远程服务器失败 ", proxyAddr)
		for {
			var _, err = localConn.Write(util.Http503)
			if err == nil {
				break
			}
		}
		for localConn.Close() != nil {
			fmt.Println("上次关闭channel失败，再次尝试", localConn.RemoteAddr())
		}
		fmt.Println("关闭channel成功 ", localConn.RemoteAddr())
		return
	}
	fmt.Println("连接到远程服务器成功 ", proxyConn.RemoteAddr())
	go handleProxyConnection(proxyConn, localConn)
	for {
		var buf = make([]byte, 2048)
		numRead, err := localConn.Read(buf)
		qufan(&buf, numRead)
		if nil != err {
			fmt.Println("读本地出错，", err)
			localConn.Close()
			proxyConn.Close()
			break
		}
		fmt.Println("从本地读到：", numRead, "字节")
		writeAllBytes(proxyConn, localConn, buf, numRead)
	}
}

func handleProxyConnection(proxyConn, localConn net.Conn) {
	for {
		var buf = make([]byte, 2048)
		numRead, err := proxyConn.Read(buf)
		qufan(&buf, numRead)
		if nil != err {
			fmt.Println("读远程出错，", err)
			proxyConn.Close()
			proxyConn.Close()
			break
		}
		fmt.Println("从远程读到：", numRead, "字节")
		writeAllBytes(localConn, proxyConn, buf, numRead)
	}
}

func qufan(bufPtr *[]byte, num int) {
	buf := *bufPtr
	for i := 0; i < num; i++ {
		buf[i] = ^buf[i]
	}
}

func printLocalIPs() {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("net.Interfaces failed, err:", err.Error())
	}

	for i := 0; i < len(netInterfaces); i++ {
		if (netInterfaces[i].Flags & net.FlagUp) != 0 {
			addrs, _ := netInterfaces[i].Addrs()

			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						fmt.Println(ipnet.IP.String())
					}
				}
			}
		}
	}
}

func writeAllBytes(dstConn net.Conn, otherConn net.Conn, buf []byte, num int) {
	for writtenNum := 0; writtenNum != num; {
		tempNum, err := dstConn.Write(buf[writtenNum:num])
		if err != nil {
			fmt.Println("写出错 ", err)
			dstConn.Close()
			otherConn.Close()
			break
		}
		writtenNum += tempNum
	}
}
