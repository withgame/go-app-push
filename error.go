package go_app_push

import "errors"

var (
	OPPOMissingDeviceErr       = errors.New("missing device err")
	OPPOMissingMasterKeyErr    = errors.New("missing oppo masterKey err")
	OPPOMissingPushTypeErr     = errors.New("missing oppo push type err")
	VIVOMissingAppSecretKeyErr = errors.New("missing vivo appSecretKey err")
	VIVOMissingAppKeyErr       = errors.New("missing vivo appKey err")
	VIVOMissingAppIdErr        = errors.New("missing vivo appId err")
	VIVOMissingTargetErr       = errors.New("missing vivo alias or regid err")
	VIVOMissingBatchTargetErr  = errors.New("missing vivo aliases or regids err")
	HWMissingClientIdErr       = errors.New("missing clientId err")
	HWMissingClientSecretErr   = errors.New("missing clientSecret err")
	MissingMeiZuAppKeyErr      = errors.New("missing meizu appid err")
	MissingAppKeyErr           = errors.New("missing appkey err")
	MissingAppPkgNameErr       = errors.New("missing appPkgName err")
)
