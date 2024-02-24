//go:build windows

package meclipboard

/*
 * 咩复制 使用Windows Api实现的剪贴板库
 * 弥补目前剪的贴板库在辣鸡Windows用起来像屎一样的问题
 * zyyme 20240224
 */

import (
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"syscall"
	"unsafe"

	"github.com/lxn/win"
	"golang.org/x/sys/windows"
)

const (
	TypeText    = "text"
	TypeFile    = "file"
	TypeMedia   = "media"
	TypeBitmap  = "bitmap"
	TypeUnknown = "unknown"
)

type MeClipboardService struct {
	hwnd       win.HWND
	clipUpdate chan bool
	lastHMen   win.HGLOBAL
}

var (
	meClip  MeClipboardService
	Formats = []uint32{win.CF_HDROP, win.CF_DIBV5, win.CF_UNICODETEXT}
)

// 注册订阅
func MustRegisterWindowClassWithWndProcPtrAndStyle(className string, wndProcPtr uintptr, style uint32) {
	hInst := win.GetModuleHandle(nil)
	if hInst == 0 {
		panic("GetModuleHandle")
	}

	hIcon := win.LoadIcon(hInst, win.MAKEINTRESOURCE(7)) // rsrc uses 7 for app icon
	if hIcon == 0 {
		hIcon = win.LoadIcon(0, win.MAKEINTRESOURCE(win.IDI_APPLICATION))
	}
	if hIcon == 0 {
		panic("LoadIcon")
	}

	hCursor := win.LoadCursor(0, win.MAKEINTRESOURCE(win.IDC_ARROW))
	if hCursor == 0 {
		panic("LoadCursor")
	}

	var wc win.WNDCLASSEX
	wc.CbSize = uint32(unsafe.Sizeof(wc))
	wc.LpfnWndProc = wndProcPtr
	wc.HInstance = hInst
	wc.HIcon = hIcon
	wc.HCursor = hCursor
	wc.HbrBackground = win.COLOR_BTNFACE + 1
	wc.LpszClassName = syscall.StringToUTF16Ptr(className)
	wc.Style = style

	if atom := win.RegisterClassEx(&wc); atom == 0 {
		panic("RegisterClassEx")
	}
}

// 给win的回调 内容变化
func clipWndProc(hwnd win.HWND, msg uint32, wp, lp uintptr) uintptr {
	switch msg {
	case win.WM_CLIPBOARDUPDATE:
		fmt.Println("回调回调回调")
		meClip.clipUpdate <- true
		return 0
	}
	return win.DefWindowProc(hwnd, msg, wp, lp)
}

func init() {
	MustRegisterWindowClassWithWndProcPtrAndStyle("meClipboard", syscall.NewCallback(clipWndProc), 0)

	hwnd := win.CreateWindowEx(
		0,
		syscall.StringToUTF16Ptr("meClipboard"),
		nil,
		0,
		0,
		0,
		0,
		0,
		win.HWND_MESSAGE,
		0,
		0,
		nil)

	if hwnd == 0 {
		panic("failed to create clipboard window")
	}

	if !win.AddClipboardFormatListener(hwnd) {
		newErr("AddClipboardFormatListener")
	}

	meClip.hwnd = hwnd
}

func MeClipboard() *MeClipboardService {
	return &meClip
}

func (c *MeClipboardService) Watch() chan bool {
	if c.clipUpdate == nil {
		c.clipUpdate = make(chan bool)
	}
	return c.clipUpdate
}

// 清空
func (c *MeClipboardService) Clear() error {
	return c.withOpenClipboard(func() error {
		if !win.EmptyClipboard() {
			return newErr("EmptyClipboard")
		}

		return nil
	})
}

// 是文本
func (c *MeClipboardService) ContainsText() (available bool, err error) {
	err = c.withOpenClipboard(func() error {
		available = win.IsClipboardFormatAvailable(win.CF_UNICODETEXT)

		return nil
	})

	return
}

func (c *MeClipboardService) ContainsFile() (available bool, err error) {
	err = c.withOpenClipboard(func() error {
		available = win.IsClipboardFormatAvailable(win.CF_HDROP)

		return nil
	})

	return
}

func (c *MeClipboardService) ContainsBitmap() (available bool, err error) {
	err = c.withOpenClipboard(func() error {
		available = win.IsClipboardFormatAvailable(win.CF_DIBV5)

		return nil
	})

	return
}

// 当前类型 是全枚举一遍，建议直接读判断err或者用上面的
func (c *MeClipboardService) ContentType() (string, error) {
	var format uint32
	err := c.withOpenClipboard(func() error {
		for _, f := range Formats {
			isAvaliable := win.IsClipboardFormatAvailable(f)
			if isAvaliable {
				format = f
				return nil
			}
		}
		return newErr("get content type of clipboard")
	})
	if err != nil {
		return "", err
	}
	switch format {
	case win.CF_HDROP:
		return TypeFile, nil
	case win.CF_DIBV5:
		return TypeBitmap, nil
	case win.CF_UNICODETEXT:
		return TypeText, nil
	default:
		return TypeUnknown, nil
	}
}

// 文本
func (c *MeClipboardService) Text() (text string, err error) {
	err = c.withOpenClipboard(func() error {
		hMem := win.HGLOBAL(win.GetClipboardData(win.CF_UNICODETEXT))
		if hMem == 0 {
			return newErr("GetClipboardData")
		}

		p := win.GlobalLock(hMem)
		if p == nil {
			return newErr("GlobalLock()")
		}
		defer win.GlobalUnlock(hMem)

		text = win.UTF16PtrToString((*uint16)(p))

		return nil
	})

	return
}

func int32Abs(val int32) uint32 {
	if val < 0 {
		return uint32(-val)
	}
	return uint32(val)
}

// 辣鸡Windows剪贴板里的图只能是bmp可太草了
func (c *MeClipboardService) Bitmap() (bmpBytes []byte, err error) {
	err = c.withOpenClipboard(func() error {
		hMem := win.HGLOBAL(win.GetClipboardData(win.CF_DIBV5))
		if hMem == 0 {
			return newErr("GetClipboardData")
		}

		p := win.GlobalLock(hMem)
		if p == nil {
			return newErr("GlobalLock()")
		}
		defer win.GlobalUnlock(hMem)

		header := (*win.BITMAPV5HEADER)(unsafe.Pointer(p))
		var biSizeImage uint32
		// BiSizeImage is 0 when use tencent TIM
		if header.BiBitCount == 32 {
			biSizeImage = 4 * int32Abs(header.BiWidth) * int32Abs(header.BiHeight)
		} else {
			biSizeImage = header.BiSizeImage
		}

		var data []byte
		sh := (*reflect.SliceHeader)(unsafe.Pointer(&data))
		sh.Data = uintptr(p)
		sh.Cap = int(header.BiSize + biSizeImage)
		sh.Len = int(header.BiSize + biSizeImage)

		// In this place, we omit AlphaMask to make sure the BiV5Header can be decoded by image/bmp
		// https://github.com/golang/image/blob/35266b937fa69456d24ed72a04d75eb6857f7d52/bmp/reader.go#L177
		if header.BiCompression == 3 && header.BV4RedMask == 0xff0000 && header.BV4GreenMask == 0xff00 && header.BV4BlueMask == 0xff {
			header.BiCompression = win.BI_RGB

			// always set alpha channel value as 0xFF to make image untransparent
			// to fix screenshot from PicPick is transparent when converted to png
			pixelStartAt := header.BiSize
			for i := pixelStartAt + 3; i < uint32(len(data)); i += 4 {
				data[i] = 0xff
			}
		}

		bmpFileSize := 14 + header.BiSize + biSizeImage
		bmpBytes = make([]byte, bmpFileSize)

		binary.LittleEndian.PutUint16(bmpBytes[0:], 0x4d42) // start with 'BM'
		binary.LittleEndian.PutUint32(bmpBytes[2:], bmpFileSize)
		binary.LittleEndian.PutUint16(bmpBytes[6:], 0)
		binary.LittleEndian.PutUint16(bmpBytes[8:], 0)
		binary.LittleEndian.PutUint32(bmpBytes[10:], 14+header.BiSize)
		copy(bmpBytes[14:], data[:])

		return nil
	})
	return
}

// 因为读取这个锁的时间实在是太长了，如果没有变，那就跳过
func (c *MeClipboardService) BitmapOnChange() (bmpBytes []byte, err error) {
	err = c.withOpenClipboard(func() error {
		hMem := win.HGLOBAL(win.GetClipboardData(win.CF_DIBV5))
		if hMem == 0 {
			return newErr("GetClipboardData")
		}
		if hMem == c.lastHMen {
			return nil
		} else {
			c.lastHMen = hMem
		}

		p := win.GlobalLock(hMem)
		if p == nil {
			return newErr("GlobalLock()")
		}
		defer win.GlobalUnlock(hMem)

		header := (*win.BITMAPV5HEADER)(unsafe.Pointer(p))
		var biSizeImage uint32
		// BiSizeImage is 0 when use tencent TIM
		if header.BiBitCount == 32 {
			biSizeImage = 4 * int32Abs(header.BiWidth) * int32Abs(header.BiHeight)
		} else {
			biSizeImage = header.BiSizeImage
		}

		var data []byte
		sh := (*reflect.SliceHeader)(unsafe.Pointer(&data))
		sh.Data = uintptr(p)
		sh.Cap = int(header.BiSize + biSizeImage)
		sh.Len = int(header.BiSize + biSizeImage)

		// In this place, we omit AlphaMask to make sure the BiV5Header can be decoded by image/bmp
		// https://github.com/golang/image/blob/35266b937fa69456d24ed72a04d75eb6857f7d52/bmp/reader.go#L177
		if header.BiCompression == 3 && header.BV4RedMask == 0xff0000 && header.BV4GreenMask == 0xff00 && header.BV4BlueMask == 0xff {
			header.BiCompression = win.BI_RGB

			// always set alpha channel value as 0xFF to make image untransparent
			// to fix screenshot from PicPick is transparent when converted to png
			pixelStartAt := header.BiSize
			for i := pixelStartAt + 3; i < uint32(len(data)); i += 4 {
				data[i] = 0xff
			}
		}

		bmpFileSize := 14 + header.BiSize + biSizeImage
		bmpBytes = make([]byte, bmpFileSize)

		binary.LittleEndian.PutUint16(bmpBytes[0:], 0x4d42) // start with 'BM'
		binary.LittleEndian.PutUint32(bmpBytes[2:], bmpFileSize)
		binary.LittleEndian.PutUint16(bmpBytes[6:], 0)
		binary.LittleEndian.PutUint16(bmpBytes[8:], 0)
		binary.LittleEndian.PutUint32(bmpBytes[10:], 14+header.BiSize)
		copy(bmpBytes[14:], data[:])

		return nil
	})
	return
}

func (c *MeClipboardService) Files() (filenames []string, err error) {
	err = c.withOpenClipboard(func() error {
		hMem := win.HGLOBAL(win.GetClipboardData(win.CF_HDROP))
		if hMem == 0 {
			return newErr("GetClipboardData")
		}
		p := win.GlobalLock(hMem)
		if p == nil {
			return newErr("GlobalLock()")
		}
		defer win.GlobalUnlock(hMem)
		filesCount := win.DragQueryFile(win.HDROP(p), 0xFFFFFFFF, nil, 0)
		filenames = make([]string, 0, filesCount)
		buf := make([]uint16, win.MAX_PATH)
		for i := uint(0); i < filesCount; i++ {
			win.DragQueryFile(win.HDROP(p), i, &buf[0], win.MAX_PATH)
			filenames = append(filenames, windows.UTF16ToString(buf))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return
}

func (c *MeClipboardService) FilesOnChange() (filenames []string, err error) {
	err = c.withOpenClipboard(func() error {
		hMem := win.HGLOBAL(win.GetClipboardData(win.CF_HDROP))
		if hMem == 0 {
			return newErr("GetClipboardData")
		}
		if hMem == c.lastHMen {
			return nil
		} else {
			c.lastHMen = hMem
		}
		p := win.GlobalLock(hMem)
		if p == nil {
			return newErr("GlobalLock()")
		}
		defer win.GlobalUnlock(hMem)
		filesCount := win.DragQueryFile(win.HDROP(p), 0xFFFFFFFF, nil, 0)
		filenames = make([]string, 0, filesCount)
		buf := make([]uint16, win.MAX_PATH)
		for i := uint(0); i < filesCount; i++ {
			win.DragQueryFile(win.HDROP(p), i, &buf[0], win.MAX_PATH)
			filenames = append(filenames, windows.UTF16ToString(buf))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return
}

func (c *MeClipboardService) SetText(s string) error {
	return c.withOpenClipboard(func() error {
		win.EmptyClipboard()
		utf16, err := syscall.UTF16FromString(s)
		if err != nil {
			return err
		}

		hMem := win.GlobalAlloc(win.GMEM_MOVEABLE, uintptr(len(utf16)*2))
		if hMem == 0 {
			return newErr("GlobalAlloc")
		}

		p := win.GlobalLock(hMem)
		if p == nil {
			return newErr("GlobalLock()")
		}

		win.MoveMemory(p, unsafe.Pointer(&utf16[0]), uintptr(len(utf16)*2))

		win.GlobalUnlock(hMem)

		if 0 == win.SetClipboardData(win.CF_UNICODETEXT, win.HANDLE(hMem)) {
			// We need to free hMem.
			defer win.GlobalFree(hMem)

			return newErr("SetClipboardData")
		}

		// The system now owns the memory referred to by hMem.
		return nil
	})
}

type DROPFILES struct {
	pFiles uintptr
	pt     uintptr
	fNC    bool
	fWide  bool
	_      uint32 // padding
}

func (c *MeClipboardService) SetFiles(paths []string) error {
	return c.withOpenClipboard(func() error {
		win.EmptyClipboard()
		// https://docs.microsoft.com/en-us/windows/win32/shell/clipboard#cf_hdrop
		var utf16 []uint16
		for _, path := range paths {
			_utf16, err := syscall.UTF16FromString(path)
			if err != nil {
				return err
			}
			utf16 = append(utf16, _utf16...)
		}
		utf16 = append(utf16, uint16(0))

		const dropFilesSize = unsafe.Sizeof(DROPFILES{}) - 4

		size := dropFilesSize + uintptr((len(utf16))*2+2)

		hMem := win.GlobalAlloc(win.GHND, size)
		if hMem == 0 {
			return newErr("GlobalAlloc")
		}

		p := win.GlobalLock(hMem)
		if p == nil {
			return newErr("GlobalLock()")
		}

		zeroMem := make([]byte, size)
		win.MoveMemory(p, unsafe.Pointer(&zeroMem[0]), size)

		pD := (*DROPFILES)(p)
		pD.pFiles = dropFilesSize
		pD.fWide = false
		pD.fNC = true
		win.MoveMemory(unsafe.Pointer(uintptr(p)+dropFilesSize), unsafe.Pointer(&utf16[0]), uintptr(len(utf16)*2))

		win.GlobalUnlock(hMem)

		if 0 == win.SetClipboardData(win.CF_HDROP, win.HANDLE(hMem)) {
			// We need to free hMem.
			defer win.GlobalFree(hMem)

			return newErr("SetClipboardData")
		}
		// The system now owns the memory referred to by hMem.

		return nil
	})
}

func (c *MeClipboardService) withOpenClipboard(f func() error) error {
	if !win.OpenClipboard(c.hwnd) {
		return newErr("OpenClipboard")
	}
	defer win.CloseClipboard()

	return f()
}

func newErr(name string) error {
	return errors.New(fmt.Sprintf("%s failed", name))
}
