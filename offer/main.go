package main

/*
#cgo LDFLAGS: -lpcre2-8
#include <pcre2.h>
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"net"
	"os"
	"unsafe"
)

func main() {
	// 目标文本
	targetText := "a;jhgoqoghqoj0329 u0tyu10hg0h9Y0Y9827342482y(Y0y(G)_)lajf;lqjfgqhgpqjopjqa=)*(^!@#$%^&*())9999999"

	// 筛选规则
	results := filterResults(targetText)

	// 将结果发送给UDP服务器
	sendResultsViaUDP(results)
}

func filterResults(targetText string) []string {
	// 编写PCRE2正则表达式
	pattern := "(\\d{4})([a-zA-Z!@#$%^&*()]+)([a-zA-Z!@#$%^&*()]{3,11})"

	cTargetText := C.CString(targetText)
	defer C.free(unsafe.Pointer(cTargetText))

	cPattern := C.CString(pattern)
	defer C.free(unsafe.Pointer(cPattern))

	re := C.pcre2_compile(cPattern, C.PCRE2_ZERO_TERMINATED, 0, nil)
	if re == nil {
		fmt.Println("Error compiling pattern")
		os.Exit(1)
	}
	defer C.pcre2_code_free(re)

	matchData := C.pcre2_match_data_create_from_pattern(re, nil)
	defer C.pcre2_match_data_free(matchData)

	var results []string

	for rc := C.pcre2_match(re, (*C.uchar)(unsafe.Pointer(cTargetText)), C.PCRE2_ZERO_TERMINATED, 0, 0, matchData, nil); rc > 0; rc = C.pcre2_match(re, (*C.uchar)(unsafe.Pointer(cTargetText)), C.PCRE2_ZERO_TERMINATED, 0, 0, matchData, nil) {
		// Extract the matched string
		start := C.pcre2_get_ovector_pointer(matchData)[2]
		end := C.pcre2_get_ovector_pointer(matchData)[3]

		matchedStr := C.GoStringN((*C.char)(unsafe.Pointer(&cTargetText[start])), C.int(end-start))

		// 检查是否符合筛选规则
		if isValidResult(matchedStr, targetText, start, end) {
			results = append(results, matchedStr)
		}

		// 移动匹配位置
		cTargetText = C.CString(targetText[end:])
		defer C.free(unsafe.Pointer(cTargetText))
	}

	return results
}

func isValidResult(result, targetText string, start, end C.int) bool {
	// 结果字符串长度为3至11个字符
	if len(result) < 3 || len(result) > 11 {
		return false
	}

	// 左侧相邻的字符串是4个数字
	if start >= 4 {
		leftAdjacent := targetText[start-4 : start]
		for _, char := range leftAdjacent {
			if char < '0' || char > '9' {
				return false
			}
		}
	} else {
		return false
	}

	// 右侧相邻的字符串不为空
	if int(end) < len(targetText)-1 {
		rightAdjacent := targetText[end+1 : end+2]
		if rightAdjacent == " " || rightAdjacent == "\t" || rightAdjacent == "\n" || rightAdjacent == "\r" {
			return false
		}
	}

	return true
}

func sendResultsViaUDP(results []string) {
	// UDP服务器地址
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:8989")
	if err != nil {
		fmt.Println("Error resolving UDP address:", err)
		os.Exit(1)
	}

	// 创建UDP连接
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		fmt.Println("Error connecting to UDP server:", err)
		os.Exit(1)
	}
	defer conn.Close()

	// 发送结果字符串
	for _, result := range results {
		_, err := conn.Write([]byte(result))
		if err != nil {
			fmt.Println("Error sending result via UDP:", err)
			os.Exit(1)
		}
	}
}
