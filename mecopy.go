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
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/foobaz/lossypng/lossypng"
	"golang.design/x/clipboard"
)

var (
	MecopyVersion     = "v3.1"
	AutoZipSize       = 8.5
	UseJpg            = false
	JpgQuality    int = 90
	PngQuality    int = 5
	OutFilename       = "mecopy.png"
	FlagOut           = false
	Force             = false
)

func main() {
	err := clipboard.Init()
	if err != nil {
		fmt.Println("初始化剪贴板失败：\n", err)
		return
	}
	if runtime.GOOS == "windows" {
		AutoZipSize = 6.5
	}

	// 从剪贴板读取文件
	data := clipboard.Read(clipboard.FmtImage)

	if len(os.Args) > 1 {
		// 参数flag
		if flg, fStr := findArg("-jpg"); flg {
			UseJpg = true
			if fStr != "" {
				num, err := strconv.ParseInt(fStr, 10, 8)
				if err == nil {
					JpgQuality = int(num)
				}
			}
		} else if flg, fStr := findArg("-png"); flg {
			UseJpg = false
			if fStr != "" {
				num, err := strconv.ParseInt(fStr, 10, 8)
				if err == nil {
					JpgQuality = int(num)
				}
			}
		}
		FlagOut, flagOStr := findArg("-o")
		if FlagOut && flagOStr != "" {
			OutFilename = flagOStr
		}
		if flg, fStr := findArg("-i"); flg {
			// 从文件读取 mecopy -i filename
			fmt.Println("从文件读取：", fStr)
			data, err = os.ReadFile(fStr)
			if err != nil {
				fmt.Println("读取文件失败：", err)
				return
			}
		}
		Force, _ = findArg("-f")

		// 多选一flag
		if flg, _ := findArg("-h"); flg {
			fmt.Println("mecopy", MecopyVersion)
			fmt.Println("https://github.com/zanjie1999/mecopy")
			fmt.Println("Usage: mecopy [options] [filename]")
			fmt.Println("       mecopy                   Compress clipboard image and copy to clipboard")
			fmt.Println("       mecopy -o [filename]     Save clipboard image to file")
			fmt.Println("       mecopy [filename]        Compress image file and copy to clipboard")
			fmt.Println("       mecopy -w [filename]     Write image file to clipboard")
			fmt.Println("       mecopy -d 8.5            Automatically compress clipboard images larger than 8.5MB in the background")
			fmt.Println("       mecopy -png 5            Use png to compress image, 0-20, 0 is lossless")
			fmt.Println("       mecopy -jpg 90           Use jpg to compress image, quality 90%, 100% is very high")
			fmt.Println("       mecopy -i [filename] -o  compress image")
			fmt.Println("       mecopy -f                force compress (even bigger than before)")
			return
		} else if flg, fStr := findArg("-d"); flg {
			num, err := strconv.ParseFloat(fStr, 64)
			if err == nil {
				AutoZipSize = num
			}
			// 后台自动压缩
			runBg()
		} else if FlagOut {
			if len(data) == 0 {
				fmt.Println("你还没有复制图片\n", string(clipboard.Read(clipboard.FmtText)))
				return
			}
			// 保存剪贴板 mecopy -o filename
			save2File(OutFilename, data)
			return
		} else if flg, fStr := findArg("-w"); flg {
			// 文件写入剪贴板 mecopy -w filename
			data, err = os.ReadFile(fStr)
			if err != nil {
				fmt.Println("读取文件失败：", err)
				return
			} else {
				fmt.Println("写入剪贴板大小：", float64(len(data))/1000/1000)
				clipboard.Write(clipboard.FmtImage, data)
			}
			return
		} else if len(os.Args) == 2 {
			// 从文件读取 mecopy filename
			fmt.Println("从文件读取：", os.Args[1])
			data, err = os.ReadFile(os.Args[1])
			if err != nil {
				fmt.Println("读取文件失败：", err)
				return
			}
		}
	}

	if len(data) > 0 {
		zipImg(data)
	} else {
		fmt.Println("你还没有复制图片")
		runBg()
	}
}

// 找flag
func findArg(key string) (bool, string) {
	for i, arg := range os.Args {
		if arg == key {
			if i+1 < len(os.Args) && !strings.HasPrefix(os.Args[i+1], "-") {
				return true, os.Args[i+1]
			}
			return true, ""
		}
	}
	return false, ""
}

func save2File(fn string, data []byte) {
	file, err := os.Create(fn)
	if err == nil {
		file.Write(data)
		file.Close()
		fmt.Println("图片已保存到：", fn)
	} else {
		fmt.Println("保存图片失败：", err)
	}
}

// 转换为jpg
func toJpg(data []byte) []byte {
	if len(data) == 0 {
		fmt.Println("你还没有复制图片\n", string(clipboard.Read(clipboard.FmtText)))
		return nil
	} else {
		fmt.Println("文件大小：", float64(len(data))/1024/1024)
	}

	// 解码图片
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		fmt.Println("图片解析失败：")
		fmt.Println(err)
		return nil
	}

	// 压缩成jpg
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, img, &jpeg.Options{Quality: JpgQuality})
	if err != nil {
		fmt.Println("压缩成jpg失败：")
		fmt.Println(err)
		return nil
	}

	out := buf.Bytes()
	fmt.Println("压缩jpg后大小：", float64(len(out))/1024/1024)
	if !Force && len(out) >= len(data) {
		fmt.Println("压缩后比原图还大！使用原图")
		return data
	}
	return out
}

// 转换为png
func toPng(data []byte) []byte {
	if len(data) == 0 {
		fmt.Println("你还没有复制图片\n", string(clipboard.Read(clipboard.FmtText)))
		return nil
	} else {
		fmt.Println("文件大小：", float64(len(data))/1000/1000)
	}

	// 解码图片
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		fmt.Println("图片解析失败：")
		fmt.Println(err)
		return nil
	}

	// 压缩成jpg
	buf := new(bytes.Buffer)
	// 自带的直接反向压缩文件更大
	// encoder := png.Encoder{CompressionLevel: png.BestCompression}
	// err = encoder.Encode(buf, img)

	// 20 quantization threshold, 0 is lossless
	optimized := lossypng.Compress(img, lossypng.RGBAConversion, PngQuality)
	err = png.Encode(buf, optimized)
	if err != nil {
		fmt.Println("压缩成png失败：")
		fmt.Println(err)
		return nil
	}

	out := buf.Bytes()
	fmt.Println("压缩后大小：", float64(len(out))/1000/1000)
	if !Force && len(out) >= len(data) {
		fmt.Println("压缩后比原图还大！使用原图")
		return data
	}
	return out
}

// 压缩图片
func zipImg(data []byte) {
	var out []byte
	if UseJpg {
		out = toJpg(data)
	} else {
		out = toPng(data)
	}
	clipboard.Write(clipboard.FmtImage, out)
	if runtime.GOOS == "windows" || FlagOut {
		save2File(OutFilename, out)
	}
}

// 后台自动压缩
func runBg() {
	var data []byte
	fmt.Println("剪贴板图片超过", AutoZipSize, "MB 时会自动压缩，请保持程序运行，按 Ctrl+C 退出")
	if UseJpg {
		fmt.Println("使用jpg，质量", JpgQuality)
	} else {
		fmt.Println("使用png，质量", PngQuality)
	}
	sizeI := int(AutoZipSize * 1000 * 1000)
	for {
		changed := clipboard.Watch(context.Background(), clipboard.FmtImage)
		data = <-changed
		if len(data) > sizeI {
			zipImg(data)
		} else {
			fmt.Println("文件未超过指定大小：", float64(len(data))/1000/1000)
		}
	}
}
