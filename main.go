package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"srt2lrc/translate"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

var i18n = false

func main() {

	args := len(os.Args)

	if args == 3 {
		i18n = os.Args[2] == "y"
	}

	if args >= 2 {
		create(os.Args[1])
	} else {
		fmt.Println("please enter file info")
	}
}

func create(filename string) {
	okbyte, err := os.ReadFile(filename)

	if err != nil {
		fmt.Println(err)
		return
	}

	list := []LRC{}

	txt := strings.Split(string(okbyte), "\n\r")
	for _, str := range txt {
		obj := coverTime(str)
		if obj.Time == "" || obj.Subtitle == "" {
			continue
		}
		list = append(list, obj)
	}

	appLength := len(translate.Apps)
	dataLength := len(list)

	app, _ := strconv.ParseFloat(strconv.Itoa(appLength), 64)
	data, _ := strconv.ParseFloat(strconv.Itoa(dataLength), 64)

	total := int(math.Ceil(data / app))
	var lrc []string

	res := []LRC{}
	for i := 0; i < total; i++ {

		if i18n {
			time.Sleep(1 * time.Second)
		}

		start := i * appLength
		end := start + appLength

		var arr []LRC

		if end > dataLength {
			arr = list[start:]
		} else {
			arr = list[start:end]
		}

		for zhi, obj := range arr {

			if i18n {
				obj.Translation = translate.Translator(obj.Subtitle, zhi)
				fmt.Printf("(%v / %v) [%v] => [%v]\n", start+zhi+1, dataLength, obj.Subtitle, obj.Translation)
			}

			res = append(res, obj)
			lrc = append(lrc, fmt.Sprintf("[%v]%v", obj.Time[:len(obj.Time)-1], obj.Subtitle+" | "+obj.Translation))
		}

	}

	txtbyte, _ := json.Marshal(res)

	filePath := strings.TrimSuffix(filename, path.Ext(filename))

	fs := filePath + ".json"
	old := filePath + "old.lrc"
	new := filePath + ".lrc"
	os.WriteFile(fs, txtbyte, os.ModePerm)
	os.WriteFile(old, []byte(strings.Join(lrc, "\n")), os.ModePerm)
	fmt.Println(fs + " done!")
	fmt.Println(new + " done!")

	if err := convertUTF8ToGBK(old, new); err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Conversion successful!")
	}
}

type LRC struct {
	Translation string `json:"i18n"`
	Subtitle    string `json:"s"`
	Time        string `json:"t"`
}

/**
1
00:00:16,766 --> 00:00:18,066
reporting work
*/

func coverTime(session string) LRC {

	obj := LRC{}

	info := strings.Split(session, "\r")
	if len(info) < 3 {
		return obj
	}

	// 00:00:16,766 --> 00:00:18,066
	start := strings.Split(info[1], "-->")

	if len(start) == 2 {
		timeline := strings.TrimSpace(start[0])
		// 00:00:16,766
		detail := strings.Split(timeline, ",")
		// 00:00:16
		arr := strings.Split(detail[0], ":")

		if len(arr) == 3 {

			h, _ := strconv.Atoi(arr[0])
			m, _ := strconv.Atoi(arr[1])
			seconds := arr[2]

			m = h*60 + m

			minutes := strconv.Itoa(m)

			if m < 10 {
				minutes = "0" + strconv.Itoa(m)
			}

			obj.Time = fmt.Sprintf("%s:%s.%s", minutes, seconds, detail[1])
		} else {
			obj.Time = timeline
		}

		fmt.Println("==============")
		fmt.Println(info[2])
		obj.Subtitle = strings.TrimLeft(info[2], "\n\r")
		obj.Translation = strings.TrimLeft(info[3], "\n\r")
		return obj
	}

	return obj
}

func convertUTF8ToGBK(inputPath, outputPath string) error {
	// 打开输入文件
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %v", err)
	}
	defer inputFile.Close()

	// 创建输出文件
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outputFile.Close()

	// 使用transform.NewReader进行编码转换
	reader := transform.NewReader(bufio.NewReader(inputFile), simplifiedchinese.GBK.NewEncoder())

	// 读取转换后的数据并写入到输出文件
	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read transformed data: %v", err)
	}

	_, err = outputFile.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to output file: %v", err)
	}

	return nil
}
