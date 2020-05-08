package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	symbol = "/"
	msgHELLO = "You have been blocked.\n"
	// sec
	timeout = 2
)

var (
	host = "0.0.0.0"
	port = "22"
	msgPre = "SSH-2.0-OpenSSH_7.6p1 Ubuntu-4ubuntu0.3\n"
	blackListpath = ""
)

func main() {
	setupHASHNET()
	setupIPTABLES()
	if component() {
		fmt.Println("Args MODE.")
	} else {
		fmt.Println("Source Mode.")
	}
	checkPath()
	listenPort(host,port)
}

func handle(conn net.Conn) {
	// Only get IP and Ignore the Port
	remoteIP := strings.Split(conn.RemoteAddr().String(),":")[0]
	fmt.Println("Remote Addr:",remoteIP)
	// Sent a msg
	sendmsg(conn,msgPre)
	// Read relpy only a time
	reply := readmsg(conn,timeout)
	// Give the client a echo anyway
	sendmsg(conn,msgHELLO)
	// Add to file
	if writeBlackList(remoteIP,reply) {
		fmt.Println("Write:",remoteIP,"Succeed.")
	} else {
		fmt.Println("Warn:","A exist IP",remoteIP,"has accesses your computer again.\nPlease Check the firewall whether is working or not.")
	}
	// Close Connection
	_ = conn.Close()
	// Bye
	insertHASHNET(remoteIP)
}

func writeBlackList(name string,data string) bool {
	fullPath := blackListpath + name
	if !pathExist(fullPath) {
		file,err := os.Create(fullPath)
		he("Create File",err,true)
		_,_ = file.Write([]byte(data))
		file.Close()
		return true
	}
	return false
}

func component()  bool{
	args := os.Args
	if len(args) > 1 {
		//host = args[1]
		port = args[1]
		blackListpath = args[2]
		msgPre = args[3]
		return true
	}
	return false
}

func checkPath() {
	if len(blackListpath) == 0 {
		wd,_ := os.Getwd()
		blackListpath = wd + symbol + port + symbol
	}
	if !pathExist(blackListpath) {
		_ = os.Mkdir(blackListpath,os.ModePerm)
	}
}

func pathExist(path string) bool{
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func sendmsg(conn net.Conn,msg string) {
	_,err :=conn.Write([]byte(msg))
	he("SEND MSG",err,false)
}

func readmsg(conn net.Conn,t int) (string) {
	// Set a read deadline
	_ = conn.SetReadDeadline(time.Now().Add(time.Second*time.Duration(t)))
	// Setup a buffer for store data (only 256 len)
	tmp := make([]byte, 256)
	// Read data
	_,er := conn.Read(tmp)
	// Decode to string
	tmpStr := string(tmp)
	// Check if nil
	if er != nil {
		he("EMPTY-DATA",er,false)
		return "Empty-Data"
	}
	// nil
	fmt.Println("Received Msg:", tmpStr)
	return tmpStr
}

func listenPort(lhost string,lport string) {
	server,err := net.Listen("tcp",lhost + ":" + lport)
	he("LISTENING PORT",err,true)
	defer server.Close()
	fmt.Println("Listening","Host:",lhost,"Port:",lport)
	for {
		// Listen for an incoming connection.
		conn, err := server.Accept()
		he("ACCEPTING CONNECTION",err,true)
		fmt.Println("Connection Established From Port:",lport)
		go handle(conn)
	}
}

func he(tag string,err error,depen bool)  {
	if err != nil {
		fmt.Println("Error " + tag + " :",err.Error())
		if depen {
			os.Exit(1)
		}
	}
}

func setupHASHNET() {
	e := exec.Command("sudo","ipset","create","SAP","hash:net").Run()
	he("CREATE HASHNET",e,true)
}

func setupIPTABLES()  {
	e := exec.Command("sudo","iptables","-t","filter","-A","INPUT","-m","set","--match-set","SAP","dst","DROP").Run()
	he("SETUP IPTABLES",e,true)
}

func insertHASHNET(ip string) {
	e := exec.Command("sudo","ipset","add","SAP",ip).Run()
	he("INSERT HASHNET",e,true)
}
