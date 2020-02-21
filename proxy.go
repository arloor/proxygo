package main

import (
	"encoding/json"
	"github.com/arloor/proxygo/extent"
	"github.com/arloor/proxygo/util"
	"log"
	"net"
	"os"
	"strconv"
)

type Info struct {
	ProxyAddr  string
	ProxyPort  int
	ClientPort int  //8081，请不要修改
	Relay      bool //如果设为true ，则只做转发，不做加解密
	Dev        bool
	Local      bool //如果为true，则只监听127.0.0.1
}

var config = Info{
	"proxy",
	8080,
	8081, //8081，请不要修改
	false,
	true,
	true,
}

func init() {
	log.Println("！！！请务必在运行前将proxy.json和pac.txt放置到", util.GetWorkDir(), "路径下")
	configinit()
	log.Println("配置信息为：", config)
	if config.Local {
		localAddr = "127.0.0.1"
	} else {
		localAddr = ""
	}
	localAddr += ":" + strconv.Itoa(config.ClientPort)
	proxyAddr = config.ProxyAddr + ":" + strconv.Itoa(config.ProxyPort)
}

var localAddr string
var proxyAddr string

func main() {

	extent.SetAutoRun()

	go extent.ServePAC()
	//printLocalIPs()

	ln, err := net.Listen("tcp", localAddr)
	if err != nil {
		log.Println("监听", localAddr, "失败 ", err)
		return
	}
	defer ln.Close()
	log.Println("成功监听 ", ln.Addr())
	for {
		c, err := ln.Accept()
		if err != nil {
			log.Println("接受连接失败 ", err)
		} else {
			log.Println("接受连接 ", c.RemoteAddr())
			go handleBrowserConnnection(c)
		}
	}
}

func handleBrowserConnnection(localConn net.Conn) {
	var proxyConn, err = net.Dial("tcp", proxyAddr)
	if err != nil {
		log.Println("连接到远程服务器失败 ", proxyAddr)
		for {
			var _, err = localConn.Write(util.Http503)
			if err == nil {
				break
			}
		}
		for localConn.Close() != nil {
			log.Println("上次关闭channel失败，再次尝试", localConn.RemoteAddr())
		}
		//log.Println("关闭channel成功 ", localConn.RemoteAddr())
		return
	}
	//log.Println("连接到远程服务器成功 ", proxyConn.RemoteAddr())
	go handleProxyConnection(proxyConn, localConn)
	var buf = make([]byte, 8192)
	for {
		numRead, err := localConn.Read(buf)
		simple(&buf, numRead)
		if nil != err {
			//log.Println("读本地出错，", err)
			localConn.Close()
			proxyConn.Close()
			break
		}
		//log.Println("从本地读到：", numRead, "字节", "from", localConn.RemoteAddr())
		writeAllBytes(proxyConn, localConn, buf, numRead)
	}
}

func handleProxyConnection(proxyConn, localConn net.Conn) {
	var buf = make([]byte, 8192)
	for {
		numRead, err := proxyConn.Read(buf)
		simple(&buf, numRead)
		if nil != err {
			//log.Println("读远程出错，", err)
			localConn.Close()
			proxyConn.Close()
			break
		}
		//log.Println("从远程读到：", numRead, "字节")
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
		log.Println("netio.Interfaces failed, err:", err.Error())
	}

	for i := 0; i < len(netInterfaces); i++ {
		if (netInterfaces[i].Flags & net.FlagUp) != 0 {
			addrs, _ := netInterfaces[i].Addrs()

			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						log.Println(ipnet.IP.String())
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
			//log.Println("写出错 ", err)
			dstConn.Close()
			otherConn.Close()
			break
		}
		writtenNum += tempNum
	}
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

func configinit() {
	log.SetFlags(log.Lshortfile | log.Flags())
	configFile, err := os.Open(util.GetWorkDir() + "proxy.json")
	if err != nil {
		log.Println("Error", "打开proxy.json失败，使用默认配置", err)
		return
	}

	bufSize := 1024
	buf := make([]byte, bufSize)
	for {
		total := 0
		n, err := configFile.Read(buf)
		total += n
		if err != nil {
			log.Println("Error", "读取proxy.json失败，使用默认配置", err)
			return
		} else if n < bufSize {
			log.Println("OK", "读取proxy.json成功")
			buf = buf[:total]
			break
		}

	}
	err = json.Unmarshal(buf, &config)
	if !config.Dev {
		logFile, _ := os.OpenFile(util.GetWorkDir()+"log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		log.SetOutput(logFile)
	}
	if err != nil {
		log.Println("Error", "读取proxy.json失败，使用默认配置", err)
		return
	}

}
