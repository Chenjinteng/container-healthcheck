# 容器状态检查工具

## 一 简介
一个能通过接口的方式进行检查特定容器运行状态的工具，主要是用于腾讯蓝鲸平台二进制版的 APPO 服务器。

### 1.1 主要问题
二进制版蓝鲸的 APPO 服务器是 `Nginx+Docker` 单机的架构，多节点部署达到高可用。SaaS 访问请求由 appo 服务器的 Nginx 响应，并代理到对应的 Docker 容器中。而多个 APPO 服务器的高可用也是通过蓝鲸平台整体的入口 Nginx 的 `upstream` 进行配置，这会使得 Nginx 的 `upstream` 配置中以 appo 上的 nginx 的端口是否存活来判断 appo 是否存在这个逻辑会存在一定的问题，appo 的 nginx 存活并不代表对应的 docker 容器也存活。
> 虽然可以使用添加 upstream cheker 的方式，但是上层 Nginx 还负责其它的转发，所以做全局的 checker 会有可能导致其它的服务判断失败，所以不适用。

![架构](https://typoraimgs-1258684427.cos.ap-guangzhou.myqcloud.com/typora_imgs202403191123998.png)

### 1.2 解决
因此需要一个方式告知上层 Nginx 在转发时，应该要以 Container 的状态为准，而不是 Nginx 的状态为准。`container-healthcheck` 就是通过接口的方式，让上层 Nginx 在转发时可以判断容器的运行状态，并转发到对应的 appo Nginx 中。


## 二 编译安装
```bash
# 依赖 golang 1.19
git clone https://github.com/Chenjinteng/container-healthcheck.git

cd container-healthcheck

make build
make install
```

## 三 使用
### 3.1 接口检查
```bash
curl http://localhost:4246/container-healthcheck/health
# 返回:
# {
#     "code": 0,
#     "status": "ok",
#     "message": "i'm healthy"
# }


curl http://localhost:4246/container-healthcheck/apps/health/bk_iam |jq .
# 正常返回：
# {
#     "code": 0,
#     "status": "ok",
#     "message": {
#         "command": "sh /build/builder",
#         "container_id": "037dbe46cd",
#         "created": 1709878119,
#         "image": "python36e:1.1",
#         "names": "bk_iam-1709878117",
#         "port": [],
#         "state": "running",
#         "status": "Up 5 days"
#     }
# }

```

### 3.2 结合 LUA 进行优化 Nginx
- 增加 `rewrite_by_lua_file`
```bash
# 1. 增加 LUA http 模块
tar xf resty-http-lua.tar -C /usr/local/openresty/lualib/resty/

# 2. 修改 paas.conf
vim /etc/consul-template/templates/paas.conf
# 把 for apps prod 的 location 修改成以下形式
location ~ ^/o/([^/]+) {
        set $app "$1";
        set $target_up "";
        # 引入 lua 脚本
        rewrite_by_lua_file $target_up lua/apps_banancer.lua;
        # proxy_pass 由 $app 修改为 $target_up
        proxy_pass http://$target_up;
        proxy_pass_header Server;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Scheme $scheme;
        proxy_set_header Host $http_host;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-Host $http_host;
        proxy_redirect off;
        proxy_read_timeout 600;
    }
```

- 增加 lua 脚本
```lua
# 1. 创建目录
mkdir /usr/local/openresty/nginx/lua

# 2. 把以下内容粘贴到 apps_banancer.lua 文件中
# vi /usr/local/openresty/nginx/lua/apps_banancer.lua
local http = require("resty.http")
local app_upstream = require "ngx.upstream"
local servers = app_upstream.get_servers(ngx.var.app)

local health_addr
local httpc = http:new()

for _, srv in ipairs(servers) do
    local addr = srv.addr
    ngx.log(ngx.INFO, "[LUA] Check upstream health =>: ", addr)
    local urlpath = "/container-healthcheck/apps/health/" .. ngx.var.app
    local host = "http://" .. addr
    local resp, err = httpc:request_uri(host ..  urlpath,
        {
             -- path = urlpath ,
             method = "GET"
        }
    )
    if not resp then
        ngx.log(ngx.ERR, "GET " .. host .. urlpath .. " ERROR, " .. err )
        goto continue
    end

    if resp.status == 200 then
         ngx.log(ngx.INFO, "[LUA] Upstream ", addr ," Health Success. Use it", " Resp:", resp.body)
         health_addr = addr
    else
         ngx.log(ngx.ERR, "[LUA] Upstream ", addr ," Health Fail, Error: ", err, " Resp:", resp.body)
         goto continue
    end

    ::continue::

end

httpc:close()
ngx.log(ngx.INFO, health_addr)
ngx.var.target_up = health_addr
```

- 在 appo 的 nginx 上添加一层转发
```bash
# 1. 添加关于 container-healthcheck 接口的转发
vim /etc/consul-template/templates/paasagent.conf

location /container-healthcheck {
    proxy_pass http://localhost:4246;
}

# 2. 重启 consul-template
systemctl restart consul-template

# 3. 使用 $LAN_IP:8010 端口进行检查上述的容器
curl -s http://$LAN_IP:8010/container-healthcheck/apps/health/bk_iam |jq .
# 正常返回：
# {
#     "code": 0,
#     "status": "ok",
#     "message": {
#         "command": "sh /build/builder",
#         "container_id": "037dbe46cd",
#         "created": 1709878119,
#         "image": "python36e:1.1",
#         "names": "bk_iam-1709878117",
#         "port": [],
#         "state": "running",
#         "status": "Up 5 days"
#     }
# }
```