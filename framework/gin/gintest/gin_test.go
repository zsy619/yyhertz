package gin

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/gin"
)

func Test_Default(t *testing.T) {
	// Disable log's color
	gin.DisableConsoleColor()
	// Force log's color
	gin.ForceConsoleColor()

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		fmt.Print("Received a request")
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.Run()
}

func Test_AsciiJSON(t *testing.T) {
	r := gin.Default()

	r.GET("/someJSON", func(c *gin.Context) {
		data := map[string]any{
			"lang": "GO语言",
			"tag":  "<br>",
		}

		// will output : {"lang":"GO\u8bed\u8a00","tag":"\u003cbr\u003e"}
		c.AsciiJSON(http.StatusOK, data)
	})

	r.Run()
}

type StructD struct {
	NestedAnonyStruct struct {
		FieldX string `form:"field_x"`
	}
	FieldD string `form:"field_d"`
}

func GetDataD(c *gin.Context) {
	var b StructD
	c.Bind(&b)
	c.JSON(200, gin.H{
		"x": b.NestedAnonyStruct,
		"d": b.FieldD,
	})
}

func Test_Bind(t *testing.T) {
	r := gin.Default()
	r.GET("/getd", GetDataD)

	r.Run()
}

type myForm struct {
	Colors []string `form:"colors[]"`
}

func formHandler(c *gin.Context) {
	var fakeForm myForm
	c.ShouldBind(&fakeForm)
	fmt.Println(fakeForm.Colors)
	c.JSON(200, gin.H{"color": fakeForm.Colors})
}

func Test_FormCheckbox(t *testing.T) {
	r := gin.Default()
	r.Use(gin.CORS())
	r.POST("/checkbox", formHandler)
	r.Run()
}

type Person struct {
	Name     string    `form:"name"`
	Address  string    `form:"address"`
	Birthday time.Time `form:"birthday" time_format:"2006-01-02" time_utc:"1"`
}

func startPage(c *gin.Context) {
	var person Person
	// If `GET`, only `Form` binding engine (`query`) used.
	// If `POST`, first checks the `content-type` for `JSON` or `XML`, then uses `Form` (`form-data`).
	// See more at https://github.com/gin-gonic/gin/blob/master/binding/binding.go#L48
	if c.ShouldBind(&person) == nil {
		fmt.Println(person.Name)
		fmt.Println(person.Address)
		fmt.Println(person.Birthday)
	}

	c.String(200, "Success")
}

func Test_Person(t *testing.T) {
	r := gin.Default()
	r.Use(gin.CORS())
	r.POST("/person", startPage)
	r.Run()
}

type PersonURI struct {
	ID   string `uri:"id" binding:"required,uuid"`
	Name string `uri:"name" binding:"required"`
}

func Test_PersonURI(t *testing.T) {
	// 启用调试模式
	gin.SetMode(gin.DebugMode)

	route := gin.Default()
	route.GET("/:name/:id", func(c *gin.Context) {
		// 调试：输出路由参数状态
		fmt.Printf("=== 路由参数调试信息 ===\n")
		fmt.Printf("c.Params: %+v\n", c.Params)
		fmt.Printf("c.Params长度: %d\n", len(c.Params))

		// 尝试手动获取参数
		name := c.Param("name")
		id := c.Param("id")
		fmt.Printf("c.Param(\"name\"): %s\n", name)
		fmt.Printf("c.Param(\"id\"): %s\n", id)

		// 使用参数检查工具
		result := c.CheckUriParams([]string{"name", "id"})
		fmt.Printf("参数检查结果: %+v\n", result)

		var person PersonURI
		if err := c.ShouldBindUri(&person); err != nil {
			c.JSON(400, gin.H{
				"msg": err.Error(),
				"debug_info": map[string]interface{}{
					"params":       c.Params,
					"param_name":   name,
					"param_id":     id,
					"check_result": result,
				},
			})
			return
		}
		c.JSON(200, gin.H{"name": person.Name, "uuid": person.ID})
	})
	route.Run()
}

type PersonForm struct {
	Name     string    `form:"name" json:"name"`
	Address  string    `form:"address" json:"address"`
	Birthday time.Time ` json:"birthday" form:"birthday" time_format:"2006-01-02" time_utc:"1"`
}

func Test_FormData(t *testing.T) {
	r := gin.Default()
	r.Use(gin.CORS())
	r.Any("/formdata", func(c *gin.Context) {
		var person PersonForm
		// If `GET`, only `Form` binding engine (`query`) used.
		// If `POST`, first checks the `content-type` for `JSON` or `XML`, then uses `Form` (`form-data`).
		// See more at https://github.com/gin-gonic/gin/blob/master/binding/binding.go#L48
		if c.Bind(&person) == nil {
			fmt.Println("====== Bind String ======")
			fmt.Println(person.Name)
			fmt.Println(person.Address)
			fmt.Println(person.Birthday)
		}

		if c.BindQuery(&person) == nil {
			fmt.Println("====== Bind By Query String ======")
			fmt.Println(person.Name)
			fmt.Println(person.Address)
			fmt.Println(person.Birthday)
		}

		if c.BindJSON(&person) == nil {
			fmt.Println("====== Bind By JSON ======")
			fmt.Println(person.Name)
			fmt.Println(person.Address)
			fmt.Println(person.Birthday)
		}

		err := c.ShouldBind(&person)
		if err == nil {
			fmt.Println("====== Bind Should String ======")
			fmt.Println(person.Name)
			fmt.Println(person.Address)
			fmt.Println(person.Birthday)
			c.String(http.StatusOK, "Success")
		} else {
			c.String(http.StatusOK, err.Error())
		}
	})
	r.Run()
}

// 自定义日志文件
func Test_CustomLogger(t *testing.T) {
	router := gin.New()
	// LoggerWithFormatter middleware will write the logs to gin.DefaultWriter
	// By default gin.DefaultWriter = os.Stdout
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// your custom format
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))
	router.Use(gin.Recovery())
	router.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})
	router.Run()
}

// 优雅重启或停止
func Test_ServerShutdown(t *testing.T) {
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		time.Sleep(5 * time.Second)
		c.String(http.StatusOK, "Welcome Gin Server")
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router.Handler(),
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	// catching ctx.Done(). timeout of 5 seconds.
	select {
	case <-ctx.Done():
		log.Println("timeout of 5 seconds.")
	}
	log.Println("Server exiting")
}

// 作为查询字符串或 postform 参数的映射
func Test_QueryMap(t *testing.T) {
	router := gin.Default()

	router.POST("/post", func(c *gin.Context) {

		ids := c.QueryMap("ids")
		names := c.PostFormMap("names")

		fmt.Printf("ids: %v; names: %v", ids, names)
	})
	router.Run()
}

// HTML 渲染
func Test_HTMLRender(t *testing.T) {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	//router.LoadHTMLFiles("templates/template1.html", "templates/template2.html")
	router.GET("/index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Main website",
		})
	})
	router.Run()
}
