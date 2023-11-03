/*
 * 咩复制 自动压缩剪贴板大图片
 * 专治QQ微信“过大图片将转换成文件发送”
 * zyyme 20231103
 */

package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"

	"golang.design/x/clipboard"
)

func main() {
	err := clipboard.Init()
	if err != nil {
		fmt.Println("初始化剪贴板失败：\n", err)
		return
	}

	// 从剪贴板读取文件
	data := clipboard.Read(clipboard.FmtImage)

	if len(os.Args) > 1 {
		if os.Args[1] == "-o" {
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

	if len(data) == 0 {
		fmt.Println("你还没有复制图片\n", string(clipboard.Read(clipboard.FmtText)))
		return
	} else {
		fmt.Println("文件大小：", len(data))
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
	fmt.Println("压缩后大小：", len(out))
	clipboard.Write(clipboard.FmtImage, out)
}
