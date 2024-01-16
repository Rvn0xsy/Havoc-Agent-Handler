package test

import (
	"regexp"
	"testing"
)

func TestGetOSVersion(t *testing.T) {
	output := "Microsoft Windows [版本 10.0.19045.3448]"
	// 使用正则表达式匹配版本号
	r, err := regexp.Compile(`\s+(\d+\.\d+\.\d+\.\d+)`)
	if err != nil {
		t.Errorf("Failed to compile regular expression: %v", err)
	}
	// 在命令输出中搜索匹配的子串
	matches := r.FindStringSubmatch(output)
	if matches == nil {
		t.Errorf("Version number not found in output: %s", output)
	}
	t.Log(matches)
	// 获取版本号
	version := matches[1]
	t.Log(version)
}
