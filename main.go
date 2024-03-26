/*
Copyright 2024 Chenjinteng

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	// "github.com/docker/docker/api/types"
	"github.com/dlclark/regexp2"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"

	// capi "github.com/hashicorp/consul/api"
	"github.com/sirupsen/logrus"
)

const (
	SUCCESS                = 0
	CONNECT_DOCKER_ERROR   = 10 // Docker 客户端连接失败
	DOCKER_OPERATIOR_ERROR = 11 // Docker 操作失败
	DOCKER_QUERY_ERROR     = 12 // Docker 查询失败

)

type APIReturn struct {
	Code    int         `json:"code"`
	Status  string      `json:"status"`
	Message interface{} `json:"message"`
}

var (
	Version, GoVersion, BuildTime, GitCommit, Author string
	//
	Logpath string

	Prefix        string
	Port          string
	AppoNginxPort string
	Level         string
	PrintVersion  bool
	//

	Log                        = logrus.New()
	GinLevel map[string]string = map[string]string{
		"info":  gin.ReleaseMode,
		"debug": gin.DebugMode,
	}

	LogrusLevel map[string]logrus.Level = map[string]logrus.Level{
		"info":  logrus.InfoLevel,
		"warn":  logrus.WarnLevel,
		"error": logrus.ErrorLevel,
		"debug": logrus.DebugLevel,
	}
)

func main() {
	flag.StringVar(&Prefix, "prefix", "/", "代理转发前缀，使用nginx做转发时需要")
	flag.StringVar(&Port, "port", "4246", "侦听的端口")
	flag.StringVar(&AppoNginxPort, "ngx-port", "8010", "appo 上的 nginx 侦听的端口")
	flag.StringVar(&Level, "level", "info", "日志级别，info|warn|error|debug")
	flag.StringVar(&Logpath, "logpath", "/var/log/container-healthcheck", "日志路径")
	flag.BoolVar(&PrintVersion, "version", false, "打印版本")
	flag.Parse()

	if PrintVersion {
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("Go Version: %s\n", GoVersion)
		fmt.Printf("Build Time: %s\n", BuildTime)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		fmt.Printf("Author: %s\n", Author)
		return
	}

	// 日志

	stdoutlog, _ := os.Create(Logpath + "/gin-stdout.log")
	applog, _ := os.Create(Logpath + "/container-healthcheck.log")

	Log.Formatter = &logrus.TextFormatter{}
	Log.Level = LogrusLevel[Level]

	Log.Out = applog

	gin.DisableConsoleColor()
	gin.SetMode(GinLevel[Level])
	gin.DefaultWriter = io.MultiWriter(stdoutlog, os.Stdout)

	router := gin.Default()
	router.GET(Prefix+"/health", health)
	router.GET(Prefix+"/apps/health/:app_code", get_app_health)
	router.HEAD(Prefix+"/apps/health/:app_code", get_app_health)

	serverAddr := fmt.Sprintf(":%s", Port)

	fmt.Printf("Listen on [%s]...\n", serverAddr)

	router.Run(serverAddr)

}

func health(c *gin.Context) {
	Log.Infof("get health")
	c.JSON(200, APIReturn{
		Code:    SUCCESS,
		Status:  "ok",
		Message: "i'm healthy",
	})
}

func get_app_health(c *gin.Context) {
	app_code := c.Param("app_code")
	all := c.DefaultQuery("all", "false")
	list_all_container := false

	Log.Debugf("Param [%s]", app_code)

	if all == "true" {
		list_all_container = true
	}

	ctx := context.Background()
	dockercli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		Log.Errorf("Create New Docker Client ERROR: %v", err.Error())
		c.JSON(500, APIReturn{
			Code:    CONNECT_DOCKER_ERROR,
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	defer dockercli.Close()

	// 所有运行中的容器
	containers, err := dockercli.ContainerList(ctx, container.ListOptions{All: list_all_container})

	if err != nil {
		Log.Errorf("List Running Containers ERROR: %v", err.Error())
		c.JSON(500, APIReturn{
			Code:    DOCKER_OPERATIOR_ERROR,
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	expr := fmt.Sprintf(`.*%s.*`, app_code)

	Log.Debugf("Expr => %s", expr)

	reg := regexp2.MustCompile(expr, 0)
	res := make([]map[string]any, 0)
	for _, contain := range containers {
		name := string(contain.Names[0])[1:]
		Log.Debugf("Running Container Name: %s", name)
		if isMatch, _ := reg.MatchString(name); isMatch {
			Log.Debugf("Match: %s => %s", expr, name)
			Log.Debugf("%v", contain)
			res = append(res, gin.H{
				"container_id": contain.ID[:10],
				"image":        contain.Image,
				"command":      contain.Command,
				"created":      contain.Created,
				"status":       contain.Status, //Up 2 days
				"port":         contain.Ports,
				"names":        name,
				"state":        contain.State, // running
			})
			// return
		}
	}

	Log.Debugf("Match res: %v", res)
	if len(res) == 0 {
		Log.Errorf("No any Matches")
		c.JSON(500, APIReturn{
			Code:    DOCKER_QUERY_ERROR,
			Status:  "error",
			Message: "no any matches",
		})
		// return
	} else if len(res) > 1 {
		Log.Errorf("Multi Matches: %v", res)
		var multi_names []string
		for _, con := range res {
			multi_names = append(multi_names, con["names"].(string))
		}
		c.JSON(500, APIReturn{
			Code:    DOCKER_QUERY_ERROR,
			Status:  "error",
			Message: fmt.Sprintf("multi container mathes by expr [%s]: %s.  PLEASE CHECK!!!", expr, strings.Join(multi_names, ",")),
		})
	} else {
		Log.Infof("%v", res)
		// 唯一，然后判断 uwsgi 是否正常启动
		b, uwsgiHealth := check_uwsgi_health(AppoNginxPort, app_code)
		if b {
			c.JSON(200, APIReturn{
				Code:    SUCCESS,
				Status:  "ok",
				Message: res[0],
			})
		} else {
			c.JSON(500, APIReturn{
				Code:    DOCKER_QUERY_ERROR,
				Status:  "error",
				Message: fmt.Sprintf("container [%s] is Running, but uwsgi service was wrong: %v", res[0]["names"], uwsgiHealth),
			})
		}
	}
}

func check_uwsgi_health(port, app_code string) (bool, string) {
	url := fmt.Sprintf("http://localhost:%s/o/%s/", port, app_code)

	// req, err := http.NewRequest(http.MethodGet, url, nil)
	// if err != nil {
	// 	Log.Errorf("NewRequest Error: %v", err.Error())
	// 	return false, err.Error()
	// }

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse /* 不跟随重定向 */
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		defer resp.Body.Close()
		Log.Errorf("Do Get Error: %v", err.Error())
		return false, err.Error()

	}

	Log.Infof("StatusCode: %d", resp.StatusCode)

	if resp.StatusCode == 502 {
		return false, "http status code is [502]"
	}

	return true, fmt.Sprint(resp.StatusCode)

}
