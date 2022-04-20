# filedrop
使用webdav在多个设备之间传输文件

## 使用方法

1. 创建配置文件
```
filedrop config
```
根据提示输入配置信息，该信息会以json形式**非加密地**保存在用户目录下的`.filedrop.json`文件中。

2. 上传文件
```
filedrop up <filepath>
```
在第一次运行时，会在webdav根目录中创建`filedrop`文件夹。文件上传完成后，会生成一个6位字母数字混合的code。上传的文件会以`code,base64(filename)`的形式保存在该目录中。

3. 下载文件到当前目录
```
filedrop down [code]
```
给出code时，会下载code所代表的文件。省略code时，会下载最新上传的文件。文件名称为上传时的名称，当前目录若存在同名文件，会**导致覆盖**。

4. 列出文件
```
filedrop list
```
列出filedrop目录内的所有文件，显示文件代码、上传日期和文件名。

5. 清理文件
```
filedrop prune
```
清理filedrop目录内，上传日期超过24小时的文件。

## 可选的webdav服务
### 坚果云
1. 注册一个[坚果云](https://www.jianguoyun.com/)帐号
2. 注册完成后，在右上角查看"用户信息"
3. 查看"安全选项"选项卡
4. 在"第三方应用管理"中添加一个应用，应用名称任意，然后生成密码
5. 将服务器地址 https://dav.jianguoyun.com/dav/ ，用户名，生成的密码，保存到filedrop的配置文件中
