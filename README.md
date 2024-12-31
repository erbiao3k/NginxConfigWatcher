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

## 运行要求
#### nginx服务器
- 在nginx服务器上直接运行时，无特殊要求
#### kubernetes集群
- nginx配置文件保存到`configmap`或`secret`，然后将配置以文件形式挂载到指定目录(必须)
- 以sidercar形式运行本程序，基础镜像用带nginx的镜像即可
- 共享主POD的进程信息到sidecar容器:`shareProcessNamespace: true`
- sidecar和主POD都需要挂载配置文件

#### 示例Deployment
```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-gateway
  labels:
    app: nginx-gateway
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx-gateway
  template:
    metadata:
      labels:
        app: nginx-gateway
    spec:
      shareProcessNamespace: true
      imagePullSecrets:
      - name: harbor-auth
      containers:
      - name: nginx-gateway
        image: youhub.com/nginx:base
        command: ["nginx"]
        args: ["-g", "daemon off;", "-c", "/etc/nginx/main/nginx.conf"]
        ports:
        - containerPort: 80
        - containerPort: 443
        volumeMounts:
        - name: nginx-site
          mountPath: /etc/nginx/vhost/
        - name: examplecom-tls
          mountPath: /etc/nginx/ssl/
          readOnly: true
        - name: nginx-main
          mountPath: /etc/nginx/main
        resources:
          requests:
            cpu: "0.1"
            memory: "50Mi"
          limits:
            cpu: "1"
            memory: "2048Mi"
            ephemeral-storage: "30Gi"
        livenessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 3
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 3
          periodSeconds: 5
        startupProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 3
          failureThreshold: 30
          periodSeconds: 10
      - name: nginx-config-watcher
        image: youhub.com/devops/nginx-config-watcher:latest
        env:
        - name: NGINX_MAIN_CONF
          valueFrom:
            configMapKeyRef:
              name: nginx-config-watcher-config
              key: NGINX_MAIN_CONF
        - name: RELOAD_DIRS
          valueFrom:
            configMapKeyRef:
              name: nginx-config-watcher-config
              key: RELOAD_DIRS
        volumeMounts:
        - name: nginx-site
          mountPath: /etc/nginx/vhost/
        - name: examplecom-tls
          mountPath: /etc/nginx/ssl/
          readOnly: true
        - name: nginx-main
          mountPath: /etc/nginx/main
        resources:
          requests:
            cpu: "0.1"
            memory: "50Mi"
          limits:
            cpu: "1"
            memory: "2048Mi"
            ephemeral-storage: "30Gi"
      volumes:
      - name: nginx-main
        configMap:
          name: nginx-main
      - name: nginx-site
        configMap:
          name: nginx-site
      - name: examplecom-tls
        secret:
          secretName: examplecom-tls
      - name: nginx-config-watcher-config
        configMap:
          name: nginx-config-watcher-config
---
apiVersion: v1
kind: Service
metadata:
  name: nginx-gateway
spec:
  selector:
    app: nginx-gateway
  ports:
    - protocol: TCP
      port: 80
      targetPort: 80
      name: http
    - protocol: TCP
      port: 443
      targetPort: 443
      name: https
  type: NodePort
```
