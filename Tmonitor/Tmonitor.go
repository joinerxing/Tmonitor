package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/docker"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	shirouNet "github.com/shirou/gopsutil/net"
)

type Config struct {
	Mode  string
	Http  string
	Udp   string
	Sec   int
	Sign  string
	Sjoin string
}

type StatInfo struct {
	Mem      string
	Cpu      string
	Disk     string
	Load     string
	Host     string
	Docker   string
	NetInter string
	NetIo    string
}

type JsonStruct struct {
}

var EtcConf = Config{}
var StatLat = StatInfo{}

//两种模式 http udp
func main() {
	JsonParse := NewJsonStruct()
	JsonParse.Load("../Tetc/config.json", &EtcConf)

var logoStr = `
  _______ _              __                 
 |__   __(_)            / _|                
    | |   _ _ __  _   _| |_ _   _ _ __  ___ 
    | |  | | '_ \| | | |  _| | | | '_ \/ __|
    | |  | | | | | |_| | | | |_| | | | \__ \
    |_|  |_|_| |_|\__, |_|  \__,_|_| |_|___/
                   __/ |                    
                  |___/                     `
	fmt.Println(logoStr)
	//return;
	//log.Println("config.json parse:", EtcConf)

	if EtcConf.Sec < 300 { //5分钟上报一次
		EtcConf.Sec = 300
	}
	if EtcConf.Sjoin == "" {
		EtcConf.Sjoin = ","
	}
	//statAll()
	if strings.EqualFold(EtcConf.Mode, "http") && EtcConf.Http != "" {
		//log.Println("config.json mode Http:")
		flushHttpData()
		return
	}

	log.Println("config.json mode Error:", "http or udp")
	os.Exit(0)
}

func NewJsonStruct() *JsonStruct {
	return &JsonStruct{}
}

func (jst *JsonStruct) Load(filename string, v interface{}) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Println(err)
		return
	}
	err = json.Unmarshal(data, v)
	if err != nil {
		log.Println(err)
		return
	}
}

func flushHttpData() {
	timer := time.NewTicker(time.Duration(EtcConf.Sec) * time.Second)
	for {
		select {
		case <-timer.C:
			statAll()
		}
	}
}

func HttpRepoert() {

	client := &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(25 * time.Second)
				c, err := net.DialTimeout(netw, addr, time.Second*20)
				if err != nil {
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
		},
	}

	form := url.Values{}

	form.Set("sign", EtcConf.Sign)
	form.Set("mem", StatLat.Mem)
	form.Set("cpu", StatLat.Cpu)
	form.Set("disk", StatLat.Disk)
	form.Set("load", StatLat.Load)
	form.Set("host", StatLat.Host)
	form.Set("docker", StatLat.Docker)
	form.Set("netio", StatLat.NetIo)
	form.Set("netinter", StatLat.NetInter)

	b := strings.NewReader(form.Encode())
	req, err := http.NewRequest("POST", EtcConf.Http, b)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := client.Do(req)
	if err != nil {
		log.Println("Fatal error ", err.Error())
		os.Exit(0)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	log.Println(string(body))
}

func tracefile(str_content string) {
	fd, _ := os.OpenFile("./Result.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	fd_time := time.Now().Format("2006-01-02 15:04:05")
	fd_content := strings.Join([]string{"======", fd_time, "=====", str_content, "\n"}, "")
	buf := []byte(fd_content)
	fd.Write(buf)
	fd.Close()
}

func statAll() {
	statMem()
	statCpu()
	statDisk()
	statLoad()
	statHost()
	statDocker()
	statNet()
	data, _ := json.Marshal(StatLat)
	HttpRepoert()
	tracefile(string(data))
}

func statMem() {
	v, _ := mem.VirtualMemory()
	data, _ := json.Marshal(v)
	StatLat.Mem = string(data)
}

func statCpu() {
	v, _ := cpu.Times(false)
	data, _ := json.Marshal(v)
	StatLat.Cpu = string(data)
}

func statDisk() {
	v1, _ := disk.Partitions(true)
	dataTmp := make([]string, 0, 20)
	for _, row := range v1 {
		v, _ := disk.Usage(row.Mountpoint)
		data, _ := json.Marshal(v)
		dataTmp = append(dataTmp, string(data))
	}
	//log.Println(dataTmp)
	StatLat.Disk = strings.Join(dataTmp, EtcConf.Sjoin)
	//log.Println(StatLat.Disk)
}

func statNet() {
	v, _ := shirouNet.Interfaces()
	data, _ := json.Marshal(v)
	StatLat.NetInter = string(data)
	//log.Println(StatLat.NetInter)

	v1, _ := shirouNet.IOCounters(true)
	data1, _ := json.Marshal(v1)
	StatLat.NetIo = string(data1)
	//log.Println(StatLat.NetIo)
}

func statLoad() {
	v, _ := load.Avg()
	data, _ := json.Marshal(v)
	StatLat.Load = string(data)
}

func statHost() {
	v, _ := host.Info()
	data, _ := json.Marshal(v)
	StatLat.Host = string(data)
}

func statDocker() {
	v, _ := docker.GetDockerStat()
	data, _ := json.Marshal(v)
	StatLat.Docker = string(data)
}
