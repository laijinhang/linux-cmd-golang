package main
/*
说明：目前实现以下几种功能
1）打印网卡信息
2）关闭和启动网卡
3）设置ip
4）设置掩码
5）设置mtu
*/

/*
#include <string.h>
#include<sys/ioctl.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <net/if.h>

// 关闭网卡
int up(char* interfaceName)
{
	int sockfd;
	struct ifreq ifr;
	bzero(&ifr,sizeof(struct ifreq));

	sockfd = socket(AF_INET,SOCK_DGRAM,0);
	strcpy(ifr.ifr_name, interfaceName);
	ioctl(sockfd, SIOCGIFFLAGS, &ifr);
	ifr.ifr_flags |= (IFF_UP | IFF_RUNNING);
	ioctl(sockfd, SIOCSIFFLAGS, &ifr);
}

// 开启网卡
int down(char* interfaceName)
{
	int sockfd;
	struct ifreq ifr;
	bzero(&ifr,sizeof(struct ifreq));
	sockfd = socket(AF_INET,SOCK_DGRAM,0);
	strcpy(ifr.ifr_name, interfaceName);
	ioctl(sockfd, SIOCGIFFLAGS, &ifr);
	ifr.ifr_flags &= ~IFF_UP;
	ioctl(sockfd, SIOCSIFFLAGS, &ifr);
}

// 设置ip
int setip(char *interName, char *ip) {
	struct ifreq temp;
	struct sockaddr_in *addr;
	int fd = 0;
	int ret =-1;
	strcpy(temp.ifr_name, interName);
	if((fd = socket(AF_INET, SOCK_STREAM, 0)) < 0) {
		return -1;
	}
	addr = (struct sockaddr_in *)&(temp.ifr_addr);
	addr->sin_family = AF_INET;
	addr->sin_addr.s_addr = inet_addr(ip);
	ret = ioctl(fd, SIOCSIFADDR, &temp);
	if(ret < 0){
		return -1;
	}
	return 0;
}

// 设置MTU
int setmtu(char *interName, int mtu)
{
	struct ifreq temp;
	struct sockaddr_in *addr;
	int fd = 0;
	int ret =-1;
	strcpy(temp.ifr_name, interName);
	if((fd = socket(AF_INET, SOCK_STREAM, 0)) < 0) {
		return -1;
	}
	temp.ifr_mtu = mtu;
	ret = ioctl(fd, SIOCSIFMTU, &temp);
	if(ret < 0){
		return -1;
	}
	return 0;
}

// 设置掩码
int setnetmask(char *interName, char *mask) {
    struct ifreq ifr;
	struct sockaddr_in *addr;
	int fd = 0;
	int ret =-1;
	strcpy(ifr.ifr_name, interName);
	if((fd = socket(AF_INET, SOCK_STREAM, 0)) < 0) {
		return -1;
	}
	struct sockaddr_in *sin = (struct sockaddr_in *) &ifr.ifr_dstaddr;
    sin->sin_family = AF_INET;
    sin->sin_port = 0;
    sin->sin_addr.s_addr = inet_addr(mask);
    return ioctl(fd, SIOCSIFNETMASK, &ifr);
}
*/
import "C"
import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"unsafe"
)

var InterInfo map[string]netDev = make(map[string]netDev)

type netDev struct {
	InterName string
	receive   Receive
	transmit  Transmit
}

type Receive struct {
	bytes      uint64
	packets    uint64
	errs       uint64
	drop       uint64
	fifo       uint64
	frame      uint64
	compressed uint64
	multicast  uint64
}

type Transmit struct {
	bytes      uint64
	packets    uint64
	errs       uint64
	drop       uint64
	fifo       uint64
	colls      uint64
	carrier    uint64
	compressed uint64
}

var upNetDev = make(map[string]bool)

func main() {
	args := os.Args
	// 打印help和version
	if len(args) > 1 {
		switch args[1] {
		case "--version":
			fmt.Printf("net-tools %.2f\nifconfig %.2f (%s)\n", 1.00, 1.00, "2018-10-10")
			return
		case "--help":
			fmt.Printf("%s\n", help)
			return
		}
	}
	interfaces, err := net.Interfaces()
	if err != nil {
		panic("Error : " + err.Error())
	}
	for _, inter := range interfaces {
		if inter.Flags.String()[0:2] == "up" {
			upNetDev[inter.Name] = true
		} else {
			upNetDev[inter.Name] = false
		}
	}
	parseNerDev()	// 解析网卡设备信息
	switch len(args) {
	case 1:
		for _, inter := range interfaces {
			if b, err := upNetDev[inter.Name];err == true && b == true {
				netInfo(inter)
			}
		}
		return
	case 2:
		for _, inter := range interfaces {
			if _, err := upNetDev[inter.Name];err == false {
				fmt.Printf("ifconfig: option `%s' not recognised.\nifconfig: `--help' gives usage information.\n", args[1])
			}
			if _, err := upNetDev[inter.Name];err == true && inter.Name == args[1] {
				netInfo(inter)
			}
		}
		return
	case 3:
		switch args[2] {
		case "up":		// 启动网卡
			C.up((*C.char)(unsafe.Pointer(C.CString(args[1]))))
			return
		case "down":	// 关网卡
			C.down((*C.char)(unsafe.Pointer(C.CString(args[1]))))
			return
		}
		if net.ParseIP(args[2]) != nil {	// 设置IPv4
			C.setip((*C.char)(unsafe.Pointer(C.CString(args[1]))), (*C.char)(unsafe.Pointer(C.CString(args[2]))))
			return
		}
	case 4:
		switch args[2] {
		case "mtu":
			mtu, _ := strconv.Atoi(args[3])
			C.setmtu((*C.char)(unsafe.Pointer(C.CString(args[1]))), C.int(mtu))
		case "netmask":
			C.setnetmask((*C.char)(unsafe.Pointer(C.CString(args[1]))), (*C.char)(unsafe.Pointer(C.CString(args[3]))))
			return
		}
	}
}

func netInfo(inter net.Interface) {
	// 第一行内容
	fmt.Print(inter.Name)
	for i := 10 - len(inter.Name); i > 0; i-- {
		fmt.Print(" ")
	}
	addr, _ := inter.Addrs()
	if addr[0].String() == "127.0.0.1/8" {
		fmt.Print("Link encap:Local Loopback")
	} else {
		fmt.Printf("Link encap:Ethernet  HWaddr %s", inter.HardwareAddr.String())
	}
	fmt.Println()
	// 第二行内容
	for i := 10; i > 0; i-- {
		fmt.Print(" ")
	}
	fmt.Print("inet addr:"+strings.Split(addr[0].String(), "/")[0], "  ")
	if addr[0].String() != "127.0.0.1/8" {
		fmt.Print("Bcast:")
		fmt.Print(getBcast(addr[0].String()), "  ")
		_, ipNet, _ := net.ParseCIDR(addr[0].String())
		mask := ipNet.Mask.String()
		a, _ := strconv.ParseInt(mask[0:2], 16, 0)
		b, _ := strconv.ParseInt(mask[2:4], 16, 0)
		c, _ := strconv.ParseInt(mask[4:6], 16, 0)
		d, _ := strconv.ParseInt(mask[6:8], 16, 0)

		fmt.Print("Mask:", net.IPv4(byte(a), byte(b), byte(c), byte(d)))
	}
	fmt.Println()
	// 第三行内容
	if b := upNetDev[inter.Name];b == true {
		for _, a := range addr[1:] {
			for i := 10; i > 0; i-- {
				fmt.Print(" ")
			}
			fmt.Printf("inet6 addr: %s  ", a.String())
			if a.String()[:3] == "127" {
				fmt.Print("Scope:Host")
			} else {
				fmt.Print("Scope:Link")
			}
			fmt.Println()
		}
	}
	// 第四行
	for i := 10; i > 0; i-- {
		fmt.Print(" ")
	}
	status := strings.Replace(strings.ToTitle(inter.Flags.String()), "|", " ", -1)
	if status[:2] == "UP" {
		temps := strings.Split(status, " ")
		temps[1] += " RUNNING"
		status = strings.Join(temps, " ")
	}
	fmt.Printf("%s  MTU:%d  Metric:1\n",
		status,
		inter.MTU)
	// 第五行	读取 /proc/net/dev文件
	for i := 10; i > 0; i-- {
		fmt.Print(" ")
	}
	fmt.Printf("RX packets:%d errors:%d dropped:%d overruns:%d frame:%d\n",
		InterInfo[inter.Name].receive.packets,
		InterInfo[inter.Name].receive.errs,
		InterInfo[inter.Name].receive.drop,
		InterInfo[inter.Name].receive.fifo,
		InterInfo[inter.Name].receive.frame)
	// 第六行
	for i := 10; i > 0; i-- {
		fmt.Print(" ")
	}
	fmt.Printf("TX packets:%d errors:%d dropped:%d overruns:%d carrier:%d\n",
		InterInfo[inter.Name].transmit.packets,
		InterInfo[inter.Name].transmit.errs,
		InterInfo[inter.Name].transmit.drop,
		InterInfo[inter.Name].transmit.fifo,
		InterInfo[inter.Name].transmit.carrier)
	// 第七行
	for i := 10; i > 0; i-- {
		fmt.Print(" ")
	}
	fmt.Printf("collisions:%d txqueuelen:1000\n", InterInfo[inter.Name].transmit.colls)
	// 第八行
	for i := 10; i > 0; i-- {
		fmt.Print(" ")
	}
	fmt.Printf("RX bytes:%d (%.1f MB)  TX bytes:%d (%.1f MB)\n",
		InterInfo[inter.Name].receive.bytes,
		float64(InterInfo[inter.Name].receive.bytes / 100000) / 10.0,
		InterInfo[inter.Name].transmit.bytes,
		float64(InterInfo[inter.Name].transmit.bytes / 100000) / 10.0)

	fmt.Println()
}

func getBcast(ip string) string {
	ips, ipNet, _ := net.ParseCIDR(ip)
	mask, _ := strconv.ParseInt(ipNet.Mask.String(), 16, 0)
	netNumber := strconv.FormatInt(int64(0xffffffff-mask), 16)

	// IP转换
	i := strings.Split(ips.String(), ".")
	ipBytes := []byte{}
	for j := 3; j >= 0; j-- {
		a, _ := strconv.Atoi(i[j])
		ipBytes = append(ipBytes, byte(a))
	}
	// 网络号转换
	netNumberBytes := []byte{}
	for i := len(netNumber) - 1; i > 0; i -= 2 {
		a, _ := strconv.ParseInt(netNumber[i-1:i+1], 16, 0)
		netNumberBytes = append(netNumberBytes, byte(a))
	}
	if len(netNumber)%2 != 0 {
		a, _ := strconv.ParseInt(string(netNumber[0]), 16, 0)
		netNumberBytes = append(netNumberBytes, byte(a))
	}
	for i := 0; i < len(netNumberBytes); i++ {
		ipBytes[i] = ipBytes[i] | netNumberBytes[i]
	}
	return net.IPv4(ipBytes[3], ipBytes[2], ipBytes[1], ipBytes[0]).String()
}

func parseNerDev() {
	fd, err := os.Open("/proc/net/dev")
	if err != nil {
		fmt.Println(err)
	}
	/*
						 packets   errs  drop fifo frame compressed multicast|  bytes     packets   errs drop fifo colls carrier compressed
		lo:    1085574   19157     0    0    0     0          0         0     1085574   19157     0     0    0     0       0          0
		ens3:  60555659  319822    0    0    0     0          0         0     46601467  351633    0     0    0     0       0          0
	*/
	input := bufio.NewScanner(fd)
	defer fd.Close()
	packData := []string{}
	for input.Scan() {
		packData = append(packData, input.Text())
	}
	packData = packData[2:]
	for _, packets := range packData {
		var nd netDev
		packd := strings.Split(packets, ":")
		nd.InterName = strings.TrimSpace(packd[0])

		n := strings.Split(packd[1], " ")
		num := []uint64{}
		for _, a := range n {
			if a != " " {
				t, err := strconv.Atoi(a)
				if err == nil {
					num = append(num, uint64(t))
				}
			}
		}
		nd.receive.bytes = num[0]
		nd.receive.packets = num[1]
		nd.receive.errs = num[2]
		nd.receive.drop = num[3]
		nd.receive.fifo = num[4]
		nd.receive.frame = num[5]
		nd.receive.compressed = num[6]
		nd.receive.multicast = num[7]

		nd.transmit.bytes = num[8]
		nd.transmit.packets = num[9]
		nd.transmit.errs = num[10]
		nd.transmit.drop = num[11]
		nd.transmit.fifo = num[12]
		nd.transmit.colls = num[13]
		nd.transmit.carrier = num[14]
		nd.transmit.compressed = num[15]
		InterInfo[nd.InterName] = nd
	}
}


const (
	help = `Usage:
  ifconfig [-a] [-v] [-s] <interface> [[<AF>] <address>]
  [add <address>[/<prefixlen>]]
  [del <address>[/<prefixlen>]]
  [[-]broadcast [<address>]]  [[-]pointopoint [<address>]]
  [netmask <address>]  [dstaddr <address>]  [tunnel <address>]
  [outfill <NN>] [keepalive <NN>]
  [hw <HW> <address>]  [metric <NN>]  [mtu <NN>]
  [[-]trailers]  [[-]arp]  [[-]allmulti]
  [multicast]  [[-]promisc]
  [mem_start <NN>]  [io_addr <NN>]  [irq <NN>]  [media <type>]
  [txqueuelen <NN>]
  [[-]dynamic]
  [up|down] ...

  <HW>=Hardware Type.
  List of possible hardware types:
    loop (Local Loopback) slip (Serial Line IP) cslip (VJ Serial Line IP)
    slip6 (6-bit Serial Line IP) cslip6 (VJ 6-bit Serial Line IP) adaptive (Adaptive Serial Line IP)
    ash (Ash) ether (Ethernet) ax25 (AMPR AX.25)
    netrom (AMPR NET/ROM) rose (AMPR ROSE) tunnel (IPIP Tunnel)
    ppp (Point-to-Point Protocol) hdlc ((Cisco)-HDLC) lapb (LAPB)
    arcnet (ARCnet) dlci (Frame Relay DLCI) frad (Frame Relay Access Device)
    sit (IPv6-in-IPv4) fddi (Fiber Distributed Data Interface) hippi (HIPPI)
    irda (IrLAP) ec (Econet) x25 (generic X.25)
    eui64 (Generic EUI-64)
  <AF>=Address family. Default: inet
  List of possible address families:
    unix (UNIX Domain) inet (DARPA Internet) inet6 (IPv6)
    ax25 (AMPR AX.25) netrom (AMPR NET/ROM) rose (AMPR ROSE)
    ipx (Novell IPX) ddp (Appletalk DDP) ec (Econet)
    ash (Ash) x25 (CCITT X.25)`
)