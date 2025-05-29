package checker

import (
	"github.com/bestruirui/bestsub/proxy/info"
	"github.com/bestruirui/bestsub/utils/log"
)

type Checker struct {
	Proxy *info.Proxy
}

func NewChecker(proxy *info.Proxy) *Checker {
	return &Checker{
		Proxy: proxy,
	}
}

func (c *Checker) Close() {
	if c.Proxy != nil {
		c.Proxy.Close()
	}
}

func (c *Checker) CheckAll() {
	log.Info("开始代理检测: %v", c.Proxy.Raw["name"])
	
	// 基础连通性检测
	c.AliveTest("http://www.google.com/generate_204", 204)
	
	// 解锁检测
	c.GoogleTest()
	c.OpenaiTest()
	c.NetflixTest()
	c.DisneyTest()
	c.YoutubeTest()
	c.CloudflareTest()
	c.TiktokTest()
	c.SpotifyTest()
	c.AmazonTest()
	
	// IP相关检测 - 新增功能
	c.IPTest()
	
	// 速度检测
	c.CheckSpeed()
	
	// 网络质量检测 - 新增功能
	if err := c.CheckNetworkQuality(); err != nil {
		log.Debug("网络质量检测完成，部分测试可能失败: %v", err)
	} else {
		log.Info("网络质量检测完成: %v", c.Proxy.Raw["name"])
	}
	
	log.Info("代理检测完成: %v", c.Proxy.Raw["name"])
}
