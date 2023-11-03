/*
 * 咩复制 自动压缩剪贴板大图片
 * 专治QQ微信“过大图片将转换成文件发送”
 * zyyme 20231103
 */

package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"strconv"

	"golang.design/x/clipboard"
)

var mecopyVersion = "v1.2"

func main() {
	err := clipboard.Init()
	if err != nil {
		fmt.Println("初始化剪贴板失败：\n", err)
		return
	}

	// 从剪贴板读取文件
	data := clipboard.Read(clipboard.FmtImage)

	if len(os.Args) > 1 {
		if os.Args[1] == "-h" {
			fmt.Println("mecopy", mecopyVersion)
			fmt.Println("https://github.com/zanjie1999/mecopy")
			fmt.Println("Usage: mecopy [options] [filename]")
			fmt.Println("       mecopy               Compress clipboard image and copy to clipboard")
			fmt.Println("       mecopy -o [filename] Save clipboard image to file")
			fmt.Println("       mecopy [filename]    Compress image file and copy to clipboard")
			fmt.Println("       mecopy -d 8.5        Automatically compress clipboard images larger than 8.5MB in the background")
			return
		} else if os.Args[1] == "-d" {
			// 后台自动压缩
			size := 8.5
			if len(os.Args) > 2 {
				size, err = strconv.ParseFloat(os.Args[2], 64)
				if err != nil {
					size = 12.0
				}
			}
			fmt.Println("剪贴板图片超过", size, "MB 时会自动压缩，请保持程序运行，按 Ctrl+C 退出")
			sizeI := int(size * 1024 * 1024)
			for {
				changed := clipboard.Watch(context.Background(), clipboard.FmtImage)
				data = <-changed
				if len(data) > sizeI {
					toJpgCopy(data)
				} else {
					fmt.Println("文件未超过指定大小：", float64(len(data))/1024/1024)
				}
			}
		} else if os.Args[1] == "-o" {
			if len(data) == 0 {
				fmt.Println("你还没有复制图片\n", string(clipboard.Read(clipboard.FmtText)))
				return
			}
			// 保存剪贴板 mecopy -o filename
			fn := "copy.png"
			if len(os.Args) > 2 {
				fn = os.Args[2]
			}
			file, err := os.Create(fn)
			if err == nil {
				file.Write(data)
				file.Close()
				fmt.Println("图片已保存到：", fn)
			} else {
				fmt.Println("保存图片失败：", err)
			}
			return
		} else {
			// 从文件读取 mecopy filename
			data, err = os.ReadFile(os.Args[1])
			if err != nil {
				fmt.Println("读取文件失败：", err)
				return
			}
		}
	}

	toJpgCopy(data)
}

// 转换为jpg并放入剪贴板
func toJpgCopy(data []byte) {
	if len(data) == 0 {
		fmt.Println("你还没有复制图片\n", string(clipboard.Read(clipboard.FmtText)))
		return
	} else {
		fmt.Println("文件大小：", float64(len(data))/1024/1024)
	}

	// 解码图片 默认png
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		img, err = jpeg.Decode(bytes.NewReader(data))
	}
	if err != nil {
		// 不行就试试通用的，这个解码不了png
		img, _, err = image.Decode(bytes.NewReader(data))
	}
	if err != nil {
		fmt.Println("剪贴板图片解析失败：")
		fmt.Println(err)
		return
	}

	// 压缩成jpgs
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, img, &jpeg.Options{Quality: 90})
	if err != nil {
		fmt.Println("压缩成jpg失败：")
		fmt.Println(err)
		return
	}

	out := buf.Bytes()
	fmt.Println("压缩后大小：", float64(len(out))/1024/1024)
	clipboard.Write(clipboard.FmtImage, out)
}
