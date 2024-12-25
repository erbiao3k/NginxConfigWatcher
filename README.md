# Nginx Config Watcher 说明文档

## 概述

Nginx Config Watcher 是一个用 Go 语言编写的程序，用于监控 Nginx 配置文件的变化，并在检测到变化时自动重新加载 Nginx 配置。

## 功能

- 监控指定目录下的 Nginx 配置文件。
- 当配置文件发生变化时，自动测试配置文件的语法是否正确。
- 如果配置文件语法正确，自动重新加载 Nginx 配置。

## 配置

### 环境变量

- `NGINX_MAIN_CONF`: Nginx 主配置文件的路径，默认为 `/etc/nginx/nginx.conf`。
- `RELOAD_DIRS`: 需要监控的目录，多个目录可以用逗号分隔，默认为 `/etc/nginx`。
