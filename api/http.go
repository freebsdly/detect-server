package api

import (
	"detect-server/connector"
	"detect-server/detector"
	"detect-server/dispatcher"
	"detect-server/tools"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"net/http"
)

type IcmpDetectPayload struct {
	Timeout int      `json:"timeout"`
	Count   int      `json:"count"`
	Type    string   `json:"type"`
	Targets []string `json:"targets"`
}

type CommonResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func NewCommonResponse(code int, msg string, data any) CommonResponse {
	return CommonResponse{
		Code:    code,
		Message: msg,
		Data:    data,
	}
}

type HttpApiOptions struct {
	Listen           string
	MaxDetectTargets int
}

func NewHttpApiOptions() HttpApiOptions {
	var options = HttpApiOptions{
		Listen: viper.GetString("api.http.listen"),
	}

	if options.Listen == "" {
		options.Listen = "0.0.0.0:8080"
	}
	return options
}

type HttpApi struct {
	srv           *gin.Engine
	options       HttpApiOptions
	icmpPublisher connector.Publisher[dispatcher.Task[detector.IcmpOptions]]
}

func (api *HttpApi) AddIcmpPublisher(publisher connector.Publisher[dispatcher.Task[detector.IcmpOptions]]) {
	api.icmpPublisher = publisher
}

func convertPayloadToIcmpTask(payload IcmpDetectPayload) ([]dispatcher.Task[detector.IcmpOptions], error) {
	var tasks = make([]dispatcher.Task[detector.IcmpOptions], 0)
	var err error
	if payload.Type == "subnet" {
		for _, subnet := range payload.Targets {
			var ips []string
			ips, err = tools.ListIpsInNetwork(subnet)
			if err != nil {
				return nil, err
			}
			var options = detector.DetectOptions[detector.IcmpOptions]{
				Count:   payload.Count,
				Timeout: payload.Timeout,
			}
			var detects = make([]detector.DetectTarget[detector.IcmpOptions], 0)
			for _, target := range ips {
				detects = append(detects, detector.NewDetectTarget(detector.ICMPDetect, target, options))
			}
			tasks = append(tasks, dispatcher.NewTask[detector.IcmpOptions](subnet, detects))
		}
	} else {
		for _, target := range payload.Targets {
			var options = detector.DetectOptions[detector.IcmpOptions]{
				Count:   payload.Count,
				Timeout: payload.Timeout,
			}
			var detect = detector.NewDetectTarget(detector.ICMPDetect, target, options)
			tasks = append(tasks, dispatcher.NewTask[detector.IcmpOptions]("task", []detector.DetectTarget[detector.IcmpOptions]{detect}))
		}
	}

	return tasks, err
}

func (api *HttpApi) HandleIcmpDetect(ctx *gin.Context) {
	var payload = IcmpDetectPayload{}
	var err = ctx.BindJSON(&payload)
	if err != nil {
		ctx.JSON(http.StatusOK, NewCommonResponse(1, err.Error(), nil))
		return
	}

	tasks, err := convertPayloadToIcmpTask(payload)
	if err != nil {
		ctx.JSON(http.StatusOK, NewCommonResponse(1, err.Error(), nil))
		return
	}
	for _, task := range tasks {
		api.icmpPublisher.Publish() <- task
	}

	ctx.JSON(http.StatusOK, NewCommonResponse(0, "ok", nil))
}

func NewHttpApi(options HttpApiOptions) *HttpApi {
	var api = &HttpApi{
		srv:     gin.New(),
		options: options,
	}

	var group = api.srv.Group("/detects")
	group.POST("/icmp", api.HandleIcmpDetect)

	return api
}

func (api *HttpApi) Start() error {
	if err := api.srv.Run(api.options.Listen); err != nil {
		return err
	}
	return nil
}
