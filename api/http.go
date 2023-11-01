package api

import (
	"detect-server/detector"
	"github.com/gin-gonic/gin"
	"net/http"
)

type HttpApiOptions struct {
	Listen           string
	MaxDetectTargets int
}

type HttpApi struct {
	srv          *gin.Engine
	options      HttpApiOptions
	icmpReceiver chan<- detector.Task[detector.IcmpDetect]
}

type IcmpDetectPayload struct {
	Timeout int      `json:"timeout"`
	Count   int      `json:"count"`
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

func (api *HttpApi) HandleIcmpDetect(ctx *gin.Context) {
	var payload = IcmpDetectPayload{}
	var err = ctx.BindJSON(&payload)
	if err != nil {
		ctx.JSON(http.StatusOK, NewCommonResponse(1, err.Error(), nil))
		return
	}

	var targets = make([]detector.IcmpDetect, 0)
	for _, target := range payload.Targets {
		var options = detector.IcmpDetect{
			Target:  target,
			Count:   payload.Count,
			Timeout: payload.Timeout,
		}

		targets = append(targets, options)
	}
	api.icmpReceiver <- detector.Task[detector.IcmpDetect]{Targets: targets}
	ctx.JSON(http.StatusOK, NewCommonResponse(0, "ok", nil))
}

func (api *HttpApi) SetIcmpReceiver(receiver chan<- detector.Task[detector.IcmpDetect]) {
	api.icmpReceiver = receiver
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
