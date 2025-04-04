# 咩复制 单文件图片压缩转换器
专治QQ微信“过大图片将转换成文件发送  
解决高分屏用户截图过大发不出去的痛点  
采用png压缩，支持透明，文件大幅缩小  
另外支持自定义压缩为jpg或者设置压缩率，具体看下面说明
## 如何使用
在右边Releases下载二进制文件  
当你复制的图片过大的时候，双击运行，会自动压缩图片并复制到剪贴板  
如果你使用的是Windows，在复制的同时会自动在当前目录输出压缩好的`mecopy.png`便于你对抗某些应用
### 进阶使用教程
```
# 将剪贴板图片保存
mecopy -o 文件名
# 从文件读取压缩并复制到剪贴板
mecopy 文件名
# 在后台自动压缩超过8.5MB的剪贴板图片（macOS QQ超过8.5MB就发不出去了）
mecopy -d 8.5
# Windows在后台自动压缩复制的图片和图片文件
mecopy -file -d
# 将图片文件直接写入剪贴板
mecopy -w 文件名
# 使用 jpg 压缩 1-90 默认90% 越高质量越好
mecopy -jpg 90
# 使用 png 压缩 0-20 默认5 越低质量越好
mecopy -png 5
# 压缩图片文件
mecopy -i 输入文件名 -o 输出文件名
# 转换图片格式  -f 在新文件更大时使用新文件
mecopy -i 原文件.png -jpg -o 输出.jpg -f
# 压缩剪贴板后输出到文件
mecopy -o 输出文件名
# 混合使用多个参数
mecopy -o mecopy.png -png 10 -d 6.5 -f
```
以上在v3.0以及之后版本均可以组合使用，请举一反三

## 4.0 以上版本
为Windows这碟醋包了顿饺子，写了个基于Windows Api的剪贴板库，在Windows下会使用meclipboard来实现剪贴板的管理  
但由于技术有限（Windows我实在是不熟），无法成功订阅Windows的剪贴板更新消息，目前采用每2秒检查一次剪贴板内容所处内存地址的方式来判断剪贴板是否被更新实现自动压缩剪贴板图片

## FAQ
- 为什么byte转mb是除以两次1000而不是1024？  
    因为QQ就是这么计算的
- 压缩后有损吗？  
    轻微有损，肉眼看不出  
- 为什么最后选用png不用其他格式？  
    因为支持透明，并且最终系统都会将他转为png  
- 为什么我在Windows QQ上粘贴图片，还是会过大？  
    因为Windows上的QQ会“压缩”剪贴板图片，将3m的图压成12m   
    我建议QQ点击发送图片的按钮，并手动选择程序输出的copy.png发送  
    <b>在v4.0，造新轮子解决了这个问题</b>



### 协议 咩License
使用此项目视为您已阅读并同意遵守 [此LICENSE](https://github.com/zanjie1999/LICENSE)   
Using this project is deemed to indicate that you have read and agreed to abide by [this LICENSE](https://github.com/zanjie1999/LICENSE)   
