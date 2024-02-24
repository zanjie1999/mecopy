//go:build !windows

package meclipboard

type MeClipboardService struct{}

func MeClipboard() *MeClipboardService                                       { return nil }
func (c *MeClipboardService) Watch() chan bool                               { return nil }
func (c *MeClipboardService) Clear() error                                   { return nil }
func (c *MeClipboardService) ContainsText() (available bool, err error)      { return false, nil }
func (c *MeClipboardService) ContainsFile() (available bool, err error)      { return false, nil }
func (c *MeClipboardService) ContainsBitmap() (available bool, err error)    { return false, nil }
func (c *MeClipboardService) ContentType() (string, error)                   { return "", nil }
func (c *MeClipboardService) Text() (text string, err error)                 { return "", nil }
func (c *MeClipboardService) Bitmap() (bmpBytes []byte, err error)           { return nil, nil }
func (c *MeClipboardService) Files() (filenames []string, err error)         { return nil, nil }
func (c *MeClipboardService) BitmapOnChange() (bmpBytes []byte, err error)   { return nil, nil }
func (c *MeClipboardService) FilesOnChange() (filenames []string, err error) { return nil, nil }
func (c *MeClipboardService) SetText(s string) error                         { return nil }
func (c *MeClipboardService) SetFiles(paths []string) error                  { return nil }
func (c *MeClipboardService) UpdateLastHMemFiles() error                     { return nil }
func (c *MeClipboardService) withOpenClipboard(f func() error) error         { return nil }
