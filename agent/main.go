package main

import "encoding/json"

var (
	Magic = []byte{0x41, 0x41, 0x41, 0x41}
)

func init() {
	err := json.Unmarshal([]byte(OptionsConfigString), &OptionsConfig)
	if err != nil {
		panic(err)
	}
}

func main() {
	agent := LinuxAgent{}
	Agent := agent.NewAgent()
	if Agent.Register() {
		// 分发任务
		Agent.DispatcherTask()
	}

}
