package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
)

type TaskPackage struct {
	Task     string      `json:"task"`
	Data     string      `json:"data"`
	External interface{} `json:"external"`
}

type Package struct {
	// size_bytes + magic + agentid
	Size         uint32
	Magic        uint32
	AgentID      uint32
	RegisterData RegisterAgentRequest
}

type RegisterAgentRequest struct {
	AgentID         string `json:"AgentID"`
	Hostname        string `json:"Hostname"`
	Username        string `json:"Username"`
	Domain          string `json:"Domain"`
	InternalIP      string `json:"InternalIP"`
	ProcessPath     string `json:"Process Path"`
	ProcessID       string `json:"Process ID"`
	ProcessParentID string `json:"Process Parent ID"`
	ProcessArch     string `json:"Process Arch"`
	ProcessElevated uint32 `json:"Process Elevated"`
	OSBuild         string `json:"OS Build"`
	Sleep           uint32 `json:"Sleep"`
	ProcessName     string `json:"Process Name"`
	OSVersion       string `json:"OS Version"`
}

func NewPackage(TaskData *TaskPackage, AgentID string) []byte {
	var PackageBuffer bytes.Buffer
	blobSizeBytes := make([]byte, 4)
	TaskJson, err := json.Marshal(TaskData)
	if err != nil {
		panic(err)
	}
	blobSize := len(TaskJson) + 12
	binary.BigEndian.PutUint32(blobSizeBytes, uint32(blobSize))
	PackageBuffer.Write(blobSizeBytes)
	PackageBuffer.Write(Magic)
	PackageBuffer.Write([]byte(AgentID))
	PackageBuffer.Write(TaskJson)
	return PackageBuffer.Bytes()
}
