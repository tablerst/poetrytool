package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Dependency struct {
	Name         string        `json:"name"`
	Version      string        `json:"version"`
	Description  string        `json:"description,omitempty"`
	Dependencies []*Dependency `json:"dependencies,omitempty"`
}

func main() {
	// 打开输入文件
	file, err := os.Open("dependencies.txt")
	if err != nil {
		fmt.Println("无法打开文件：", err)
		return
	}
	defer file.Close()

	var rootDependencies []*Dependency
	stack := []*Dependency{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// 跳过空行
		if strings.TrimSpace(line) == "" {
			continue
		}

		// 解析行的层级和内容
		level, content := parseLine(line)

		// 解析依赖项信息
		dep := parseDependency(content)

		if level == 0 {
			// 根依赖项
			rootDependencies = append(rootDependencies, dep)
			stack = stack[:0]
			stack = append(stack, dep)
		} else {
			// 调整堆栈到当前层级
			if level > len(stack) {
				fmt.Printf("错误：层级超过堆栈大小，在行：%s\n", line)
				continue
			}
			stack = stack[:level]
			parent := stack[level-1]
			parent.Dependencies = append(parent.Dependencies, dep)
			stack = append(stack, dep)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("读取文件出错：", err)
		return
	}

	// 将结果写入 JSON 文件
	outputFile, err := os.Create("dependencies.json")
	if err != nil {
		fmt.Println("无法创建输出文件：", err)
		return
	}
	defer outputFile.Close()

	encoder := json.NewEncoder(outputFile)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false) // 防止转义 < 和 > 符号
	if err := encoder.Encode(rootDependencies); err != nil {
		fmt.Println("编码 JSON 出错：", err)
		return
	}

	fmt.Println("依赖树已写入 dependencies.json")
}

func parseLine(line string) (int, string) {
	level := 0
	pos := 0
	runes := []rune(line)
	length := len(runes)

	// 检查行首是否直接有树形符号，没有缩进
	if pos+4 <= length && (string(runes[pos:pos+4]) == "├── " || string(runes[pos:pos+4]) == "└── ") {
		level = 1
		pos += 4
	} else {
		// 解析缩进
		for pos < length {
			if pos+4 <= length && (string(runes[pos:pos+4]) == "    " || string(runes[pos:pos+4]) == "│   ") {
				level++
				pos += 4
			} else {
				break
			}
		}
		// 检查缩进后是否有树形符号
		if pos+4 <= length && (string(runes[pos:pos+4]) == "├── " || string(runes[pos:pos+4]) == "└── ") {
			pos += 4
			level++ // 树形符号表示进入下一级
		}
	}

	content := strings.TrimSpace(string(runes[pos:]))
	return level, content
}

func parseDependency(content string) *Dependency {
	// 使用正则表达式解析依赖项
	re := regexp.MustCompile(`^(\S+)\s+([^\s]+)\s*(.*)$`)
	matches := re.FindStringSubmatch(content)
	if matches == nil {
		// 行不符合预期格式
		return &Dependency{
			Name: content,
		}
	}
	name := matches[1]
	version := matches[2]
	description := strings.TrimSpace(matches[3])

	return &Dependency{
		Name:        name,
		Version:     version,
		Description: description,
	}
}
