package blockchain

import (
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn/chclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn/chmgmtclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn/resmgmtclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	"time"
)

const ChaincodeVersion = "1.0"

type FabricSetup struct {
	ConfigFile    string // SDK配置文件所在路径
	ChannelID     string // 应用通道名称
	ChannelConfig string // 应用通道交易配置文件所在路径
	OrgAdmin      string // 组织管理员名称
	OrgName       string // 组织名称

	initialized bool // 是否初始化

	admin resmgmtclient.ResourceMgmtClient // fabric环境中资源管理者
	sdk   *fabsdk.FabricSDK                // SDK实例

	// 链码相关
	ChaincodeID     string
	ChaincodeGoPath string
	ChaincodePath   string
	UserName        string
	Client          chclient.ChannelClient
}

// 1. 创建SDK实例并使用SDK实例创建应用通道, 将peer节点加入到创建的应用通道中
func (t *FabricSetup) Initialize() error {

	fmt.Println("初始化SDK...")

	if t.initialized {
		return fmt.Errorf("SDK已被实例化")
	}

	// 创建SDK实例对象
	sdk, err := fabsdk.New(config.FromFile(t.ConfigFile))
	if err != nil {
		return fmt.Errorf("实例化SDK失败, %v", err)
	}
	t.sdk = sdk

	// 创建一个具有管理权限的应用通道客户端管理对象
	// t.orgAdmin = Admin, t.orgName = org1
	chMgmtClient, err := t.sdk.NewClient(fabsdk.WithUser(t.OrgAdmin), fabsdk.WithOrg(t.OrgName)).ChannelMgmt()
	if err != nil {
		return fmt.Errorf("创建应用通道客户端管理对象失败, %v", err)
	}

	// 获取当前会话用户对象(Admin)
	session, err := t.sdk.NewClient(fabsdk.WithUser(t.OrgAdmin), fabsdk.WithOrg(t.OrgName)).Session()
	if err != nil {
		return fmt.Errorf("获取当前会话用户对象失败: %v", err)
	}
	orgAdminUser := session

	// 指定创建应用通道所需要的所有参数
	chReq := chmgmtclient.SaveChannelRequest{ChannelID: t.ChannelID, ChannelConfig: t.ChannelConfig, SigningIdentity: orgAdminUser}

	// 创建应用通道
	err = chMgmtClient.SaveChannel(chReq)

	if err != nil {
		return fmt.Errorf("创建应用通道失败: %v", err)
	}

	time.Sleep(time.Second * 5)

	// 创建一个管理资源的客户端对象
	t.admin, err = t.sdk.NewClient(fabsdk.WithUser(t.OrgAdmin)).ResourceMgmt()
	if err != nil {
		return fmt.Errorf("创建资源管理对象失败: %v", err)
	}

	// 将peers节点加入到应用通道中
	err = t.admin.JoinChannel(t.ChannelID)
	if err != nil {
		return fmt.Errorf("peers加入应用通道失败: %v", err)
	}

	fmt.Println("SDK初始化成功")
	t.initialized = true
	return nil
}

// 2. 安装及实例化链码

func (setup *FabricSetup) InstallAndInstantiateCC() error {
	fmt.Println("开始安装链码......")
	// 将指定的链码打包
	ccPkg, err := gopackager.NewCCPackage(setup.ChaincodePath, setup.ChaincodeGoPath)
	if err != nil {
		return fmt.Errorf("创建链码包失败: %v", err)
	}

	// 指定安装链码时所需要的各项参数
	installCCReq := resmgmtclient.InstallCCRequest{Name: setup.ChaincodeID, Path: setup.ChaincodePath, Version: ChaincodeVersion, Package: ccPkg}

	// 在OrgPeer上安装链码
	// -n -v -p
	_, err = setup.admin.InstallCC(installCCReq)
	if err != nil {
		return fmt.Errorf("安装链码失败: %v", err)
	}

	fmt.Println("指定的链码安装成功")
	fmt.Println("开始实例化链码......")

	// 指定链码策略
	ccPolicy := cauthdsl.SignedByAnyMember([]string{"Org1MSP"})

	// 指定实例化链码时所需要的各项参数
	// -n -v -C -c
	instantiateCCReq := resmgmtclient.InstantiateCCRequest{Name: setup.ChaincodeID, Path: setup.ChaincodePath, Version: ChaincodeVersion, Args: [][]byte{[]byte("init")}, Policy: ccPolicy}

	// 实例化链码
	err = setup.admin.InstantiateCC(setup.ChannelID, instantiateCCReq)
	if err != nil {
		return fmt.Errorf("实例化链码失败: %v", err)
	}

	fmt.Println("链码实例化成功")

	// 创建通道客户端, 用于调用链码进行查询与执行事务
	setup.Client, err = setup.sdk.NewClient(fabsdk.WithUser(setup.UserName)).Channel(setup.ChannelID)
	if err != nil {
		return fmt.Errorf("创建应用通道客户端失败: %v", err)
	}

	fmt.Println("创建客户端成功, 可以调用链码进行查询或执行事务")
	return nil
}
