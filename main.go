package main

import (
	"context"
	"flag"
	"fmt"
	"io"
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
	Prefix       string
	Port         string
	Level        string
	PrintVersion bool
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
	flag.StringVar(&Prefix, "prefix", "/", "prefix")
	flag.StringVar(&Port, "port", "8080", "port")
	flag.StringVar(&Level, "level", "info", "level")
	flag.BoolVar(&PrintVersion, "print", false, "print")
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
	Log.Formatter = &logrus.TextFormatter{}
	Log.Level = LogrusLevel[Level]
	Log.Out = os.Stdout

	gin.DisableConsoleColor()
	gin.SetMode(GinLevel[Level])
	gin.DefaultWriter = io.MultiWriter(os.Stdout)

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
	c.IndentedJSON(200, APIReturn{
		Code:    SUCCESS,
		Status:  "ok",
		Message: "i'm healthy",
	})
}

func get_app_health(c *gin.Context) {
	app_code := c.Param("app_code")

	Log.Debugf("Param [%s]", app_code)

	ctx := context.Background()
	dockercli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		Log.Errorf("Create New Docker Client ERROR: %v", err.Error())
		c.IndentedJSON(500, APIReturn{
			Code:    CONNECT_DOCKER_ERROR,
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	defer dockercli.Close()

	// 所有运行中的容器
	containers, err := dockercli.ContainerList(ctx, container.ListOptions{All: false})

	if err != nil {
		Log.Errorf("List Running Containers ERROR: %v", err.Error())
		c.IndentedJSON(500, APIReturn{
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
		c.IndentedJSON(500, APIReturn{
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
		c.IndentedJSON(500, APIReturn{
			Code:    DOCKER_QUERY_ERROR,
			Status:  "error",
			Message: fmt.Sprintf("multi container mathes by [%s] => %s, please check!!!", expr, strings.Join(multi_names, ",")),
		})
	} else {
		Log.Infof("%v", res)
		c.IndentedJSON(200, APIReturn{
			Code:    SUCCESS,
			Status:  "ok",
			Message: res[0],
		})
		// update_consul_kv(fmt.Sprintf("bkapps/upstreams/prod/%s", app_code))
	}

}

// TODO: 更新 Consul KV 值
// func update_consul_kv(key string) {
// 	client, err := capi.NewClient(capi.DefaultConfig())
// 	if err != nil {
// 		Log.Errorf("Create New Consul Client ERROR: %v", err.Error())
// 		return
// 	}
//
// 	// get kv
// 	kv := client.KV()
//
// 	pair, _, err := kv.Get(key, nil)
//
// 	if err != nil {
// 		Log.Errorf("Get Consul KV ERROR: %v", err.Error())
// 		return
// 	}
// 	Log.Infof("%v %s", pair.Key, pair.Value)
// }