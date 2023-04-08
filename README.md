# V2bX

[![](https://img.shields.io/badge/TgChat-%E4%BA%A4%E6%B5%81%E7%BE%A4-blue)](https://t.me/YuzukiProjects)

A V2board node server based on Xray-core, modified from XrayR

一个基于Xray-core的V2board节点服务端，修改自XrayR，支持V2ay,Trojan,Shadowsocks协议。

如果您喜欢本项目，可以右上角点个star+watch，持续关注本项目的进展。

使用教程：[详细使用教程](https://yuzuki-1.gitbook.io/v2bx-doc/)

如对脚本不放心，可使用此沙箱先测一遍再使用：https://killercoda.com/playgrounds/scenario/ubuntu

**注意：1.1.0将更换为V2board1.7.0之后新增的Api，原Api将被移除，请1.7.0之前的用户使用1.1.0之前的版本。**

## 免责声明

本项目只是本人个人学习开发并维护，本人不保证任何可用性，也不对使用本软件造成的任何后果负责。

## 特点

* 永久开源且免费。
* 支持V2ray，Trojan， Shadowsocks多种协议。
* 支持Vless和XTLS等新特性。
* 支持单实例对接多面板、多节点，无需重复启动。
* 支持限制在线IP
* 支持节点端口级别、用户级别限速。
* 配置简单明了。
* 修改配置自动重启实例。
* 方便编译和升级，可以快速更新核心版本， 支持Xray-core新特性。

## 功能介绍

| 功能        | v2ray | trojan | shadowsocks |
|-----------|-------|--------|-------------|
| 获取节点信息    | √     | √      | √           |
| 获取用户信息    | √     | √      | √           |
| 用户流量统计    | √     | √      | √           |
| 自动申请tls证书 | √     | √      | √           |
| 自动续签tls证书 | √     | √      | √           |
| 在线人数统计    | √     | √      | √           |
| 在线IP数限制   | √     | √      | √           |
| 跨节点IP数限制  | √     | √      | √           |
| 审计规则      | √     | √      | √           |
| 按照用户限速    | √     | √      | √           |
| 自定义DNS    | √     | √      | √           |
| 动态限速(未测试) | √     | √      | √           |

## 支持前端

| 前端      | v2ray | trojan | shadowsocks |
|---------|-------|--------|-------------|
| v2board | √     | √      | √           |

## TODO

* 抽象化Core包以摆脱对Xray-Core的强依赖性
* 增加Hysteria支持

## 软件安装

### 一键安装

```
wget -N https://raw.githubusercontents.com/Yuzuki616/V2bX-script/master/install.sh && bash install.sh
```

### 手动安装

[手动安装教程](https://crackair.gitbook.io/xrayr-project/xrayr-xia-zai-he-an-zhuang/install/manual)

## 配置文件及详细使用教程

[详细使用教程](https://crackair.gitbook.io/xrayr-project/)

## Thanks

* [Project X](https://github.com/XTLS/)
* [V2Fly](https://github.com/v2fly)
* [VNet-V2ray](https://github.com/ProxyPanel/VNet-V2ray)
* [Air-Universe](https://github.com/crossfw/Air-Universe)
* [XrayR](https://github.com/Misaka-blog/XrayR)

## Licence

[Mozilla Public License Version 2.0](https://github.com/XrayR-project/XrayR/blob/master/LICENSE)

## Telgram

## Stars 增长记录

[![Stargazers over time](https://starchart.cc/Yuzuki616/V2bX.svg)](https://starchart.cc/Yuzuki616/V2bX)
