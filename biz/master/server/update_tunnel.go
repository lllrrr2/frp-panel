package server

import (
	"context"
	"fmt"

	"github.com/VaalaCat/frp-panel/common"
	"github.com/VaalaCat/frp-panel/dao"
	"github.com/VaalaCat/frp-panel/pb"
	"github.com/VaalaCat/frp-panel/rpc"
	"github.com/VaalaCat/frp-panel/utils"
	"github.com/sirupsen/logrus"
)

func UpdateFrpsHander(c context.Context, req *pb.UpdateFRPSRequest) (*pb.UpdateFRPSResponse, error) {
	logrus.Infof("update frps, req: [%+v]", req)
	var (
		serverID  = req.GetServerId()
		configStr = req.GetConfig()
		userInfo  = common.GetUserInfo(c)
	)

	if len(configStr) == 0 || len(serverID) == 0 {
		return nil, fmt.Errorf("request invalid")
	}

	srv, err := dao.GetServerByServerID(userInfo, serverID)
	if srv == nil || err != nil {
		logrus.WithError(err).Errorf("cannot get server, id: [%s]", serverID)
		return nil, err
	}

	srvCfg, err := utils.LoadServerConfig(configStr, true)
	if srvCfg == nil || err != nil {
		logrus.WithError(err).Errorf("cannot load server config")
		return nil, err
	}

	if err := srv.SetConfigContent(srvCfg); err != nil {
		logrus.WithError(err).Errorf("cannot set server config")
		return nil, err
	}

	if err := dao.UpdateServer(userInfo, srv); err != nil {
		logrus.WithError(err).Errorf("cannot update server, id: [%s]", serverID)
		return nil, err
	}

	go func() {
		resp, err := rpc.CallClient(context.Background(), req.GetServerId(), pb.Event_EVENT_UPDATE_FRPS, req)
		if err != nil {
			logrus.WithError(err).Errorf("update event send to server error, server id: [%s]", req.GetServerId())
		}
		if resp == nil {
			logrus.Errorf("cannot get response, server id: [%s]", req.GetServerId())
		}
	}()

	logrus.Infof("update frps success, id: [%s]", serverID)
	return &pb.UpdateFRPSResponse{
		Status: &pb.Status{Code: pb.RespCode_RESP_CODE_SUCCESS, Message: "ok"},
	}, nil
}
