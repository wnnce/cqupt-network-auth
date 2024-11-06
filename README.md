# 重庆邮电大学校园网认证程序

使用`Go`实现，可交叉编译到`Linux`、路由器等多平台使用。使用简单，只需指定运行`flag`即可，程序会自动选择校园网所对应的网络接口，如果有多个接口网络地址具备校园网特征，可以手动选择。

```shell
# 编译到Linux平台使用
go env -w GOOS=linux
go env -w GOARCH=amd64
go build -o network-auth ./main.go

# 编译到windows平台使用
go env -w GOOS=windows
go env -w GOARCH=amd64
go build -o network-auth.exe ./main.go

# 编译到路由器使用
go env -w GOOS=linux
# 自行确定路由器的CPU架构
go env -w GOARCH=mipsle
go build -o network-auth ./main.go
```

编译完成后，一行命令就可以认证校园了，以`Linux`平台为例

```shell
./network-auth -username <username> -password <password> -ua desktop -isp telecom
```

共有四个可选参数
- `username`: 校园网登录用户，**必填**
- `password`: 校园网登录密码，**必填**
- `ua`: 选择以什么设备连接校园网，可选参数`desktop | phone | pad`
- `isp`: 校园网的运营商，可选参数`unicom | telecom | cmcc | xyw`

认证成功或失败，控制台都会打印信息

```shell
2024/11/06 03:56:35 interface lo not active ipv4 address, break
2024/11/06 03:56:35 interface eth0 not active ipv4 address, break
2024/11/06 03:56:35 interface lan1 not active ipv4 address, break
2024/11/06 03:56:35 interface lan2 not active ipv4 address, break
2024/11/06 03:56:35 interface ip6tnl0 not active ipv4 address, break
2024/11/06 03:56:35 interface sit0 not active ipv4 address, break
2024/11/06 03:56:35 interface gre0 not active ipv4 address, break
2024/11/06 03:56:35 interface gretap0 not active ipv4 address, break
2024/11/06 03:56:35 interface erspan0 not active ipv4 address, break
2024/11/06 03:56:35 interface ip6gre0 not active ipv4 address, break
2024/11/06 03:56:35 interface phy1-ap0 not active ipv4 address, break
2024/11/06 03:56:35 当前设备已认证
```