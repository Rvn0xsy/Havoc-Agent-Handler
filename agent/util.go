package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func GetRandomItem(list []string) string {
	if len(list) == 0 {
		return ""
	}
	// 设置随机数种子
	rand.Seed(time.Now().UnixNano())
	return list[rand.Intn(len(list))]
}

func PostWithHeaders(body []byte) ([]byte, error) {
	protocol := "http://"
	if OptionsConfig.Listener.Secure {
		protocol = "https://"
	}

	port := OptionsConfig.Listener.PortConn
	if port == "" {
		port = OptionsConfig.Listener.PortBind
	}

	address := GetRandomItem(OptionsConfig.Listener.Hosts)
	host := OptionsConfig.Listener.HostHeader
	uri := GetRandomItem(OptionsConfig.Listener.Uris)
	url := fmt.Sprintf("%s%s:%s/%s", protocol, host, port, uri)
	portInt, _ := strconv.Atoi(port)
	transport := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			// 替换为您的目标 IP 地址和端口
			ip := net.ParseIP(address)
			targetAddr := &net.TCPAddr{
				IP:   ip,
				Port: portInt,
			}
			// 创建连接
			conn, err := net.DialTCP(network, nil, targetAddr)
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
	}

	// 创建一个自定义的 Client
	client := &http.Client{
		Transport: transport,
		Timeout:   3 * time.Second,
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	req.Header.Set("User-Agent", OptionsConfig.Listener.UserAgent)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println(err)
		}
	}(resp.Body)
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return respBody, nil
}

func RemoveEmptyStrings(arr []string) []string {
	var result []string
	for _, str := range arr {
		trimmedStr := strings.TrimSpace(str)
		if trimmedStr != "" {
			result = append(result, trimmedStr)
		}
	}
	return result
}

func Base64Encode(data []byte) string {
	encodedString := base64.StdEncoding.EncodeToString(data)
	return encodedString
}

func Base64Decode(encodedString string) []byte {
	decodedData, err := base64.StdEncoding.DecodeString(encodedString)
	if err != nil {
		log.Println(err)
		return []byte{}
	}
	return decodedData
}
