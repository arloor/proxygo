package main

import (
	"encoding/json"
	"fmt"
	"github.com/arloor/proxygo/pac"
	"github.com/arloor/proxygo/util"
	"log"
	"net"
	"os"
	"strconv"
)

type Info struct {
	ProxyAddr  string
	ProxyPort  int
	ClientPort int
	Relay      bool
}

//implement JSONObject
func (configInfo Info) ToJSONString() (str string, error error) {
	b, err := json.Marshal(configInfo)
	if err != nil {
		return "", err
	} else {
		return string(b), nil
	}
}

func (configInfo Info) String() string {
	str, _ := configInfo.ToJSONString()
	return str
}

var config = Info{
	"proxy",
	8080,
	8081,
	false,
}

func init() {
	configinit()
	localAddr = ":" + strconv.Itoa(config.ClientPort)
	proxyAddr = config.ProxyAddr + ":" + strconv.Itoa(config.ProxyPort)
}
func configinit() {
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Println("Error", "打开config.json失败，使用默认配置", err)
		return
	} else {
		log.Println("Done", "打开config.json成功", "下面读取配置文件")
		log.Println("Reading...")
	}
	bufSize := 1024
	buf := make([]byte, bufSize)
	for {
		total := 0
		n, err := configFile.Read(buf)
		total += n
		if err != nil {
			log.Println("Error", "读取config.json失败，使用默认配置", err)
			return
		} else if n < bufSize {
			log.Println("Done", "读取config.json成功")
			buf = buf[:total]
			break
		}

	}
	err = json.Unmarshal(buf, &config)
	if err != nil {
		log.Println("Error", "读取config.json失败，使用默认配置", err)
		return
	} else {
		log.Println("Done", "config被设置为", config)
	}

}

var localAddr string
var proxyAddr string

//var proxyAddr = "193.187.119.219:8080"
//var proxyAddr = "67.230.170.45:8080"
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
		simple(&buf, numRead)
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
		simple(&buf, numRead)
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

func simple(bufPtr *[]byte, num int) {
	if !config.Relay {
		buf := *bufPtr
		for i := 0; i < num; i++ {
			buf[i] = ^buf[i]
		}
	}
}

func printLocalIPs() {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("netio.Interfaces failed, err:", err.Error())
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
