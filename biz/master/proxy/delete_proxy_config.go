package proxy

import (
	"context"
	"fmt"

	"github.com/VaalaCat/frp-panel/biz/master/client"
	"github.com/VaalaCat/frp-panel/common"
	"github.com/VaalaCat/frp-panel/dao"
	"github.com/VaalaCat/frp-panel/logger"
	"github.com/VaalaCat/frp-panel/pb"
	v1 "github.com/fatedier/frp/pkg/config/v1"
	"github.com/samber/lo"
)

func DeleteProxyConfig(c context.Context, req *pb.DeleteProxyConfigRequest) (*pb.DeleteProxyConfigResponse, error) {
	var (
		userInfo  = common.GetUserInfo(c)
		clientID  = req.GetClientId()
		serverID  = req.GetServerId()
		proxyName = req.GetName()
	)

	if len(clientID) == 0 || len(serverID) == 0 || len(proxyName) == 0 {
		return nil, fmt.Errorf("request invalid")
	}

	cli, err := dao.GetClientByClientID(userInfo, clientID)
	if err != nil {
		logger.Logger(c).WithError(err).Errorf("cannot get client, id: [%s]", clientID)
		return nil, err
	}
	if cli.ServerID != serverID {
		return nil, fmt.Errorf("client and server not match")
	}

	oldCfg, err := cli.GetConfigContent()
	if err != nil {
		logger.Logger(c).WithError(err).Errorf("cannot get client config content, id: [%s]", clientID)
		return nil, err
	}

	oldCfg.Proxies = lo.Filter(oldCfg.Proxies, func(p v1.TypedProxyConfig, _ int) bool {
		return p.GetBaseConfig().Name != proxyName
	})

	if err := cli.SetConfigContent(*oldCfg); err != nil {
		logger.Logger(c).WithError(err).Errorf("cannot set client config, id: [%s]", clientID)
		return nil, err
	}

	if err := dao.UpdateClient(userInfo, cli); err != nil {
		logger.Logger(c).WithError(err).Errorf("cannot update client, id: [%s]", clientID)
		return nil, err
	}

	rawCfg, err := cli.MarshalJSONConfig()
	if err != nil {
		logger.Logger(c).WithError(err).Errorf("cannot marshal client config, id: [%s]", clientID)
		return nil, err
	}

	_, err = client.UpdateFrpcHander(c, &pb.UpdateFRPCRequest{
		ClientId: &cli.ClientID,
		ServerId: &serverID,
		Config:   rawCfg,
		Comment:  &cli.Comment,
	})
	if err != nil {
		logger.Logger(c).WithError(err).Errorf("cannot update frpc, id: [%s]", clientID)
		return nil, err
	}

	if err := dao.DeleteProxyConfig(userInfo, clientID, proxyName); err != nil {
		logger.Logger(c).WithError(err).Errorf("cannot delete proxy config, id: [%s]", clientID)
		return nil, err
	}

	logger.Logger(c).Infof("delete proxy config, id: [%s], name: [%s]", clientID, proxyName)

	return &pb.DeleteProxyConfigResponse{}, nil
}
