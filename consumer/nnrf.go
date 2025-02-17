// SPDX-FileCopyrightText: 2021 Open Networking Foundation <info@opennetworking.org>
// Copyright 2019 free5GC.org
//
// SPDX-License-Identifier: Apache-2.0

package consumer

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/antihax/optional"
	"github.com/free5gc/smf/metrics"
	"github.com/free5gc/smf/msgtypes/svcmsgtypes"
	"github.com/mohae/deepcopy"

	"github.com/free5gc/openapi"
	"github.com/free5gc/openapi/Nnrf_NFDiscovery"
	"github.com/free5gc/openapi/Nudm_SubscriberDataManagement"
	"github.com/free5gc/openapi/models"
	smf_context "github.com/free5gc/smf/context"
	"github.com/free5gc/smf/logger"
)

func SendNFRegistration() error {
	sNssais := []models.Snssai{}

	if len(*smf_context.SmfInfo.SNssaiSmfInfoList) == 0 {
		logger.ConsumerLog.Errorf("slice info not available, dropping NRF registration")
		return fmt.Errorf("slice info nil")
	}

	for _, snssaiSmfInfo := range *smf_context.SmfInfo.SNssaiSmfInfoList {
		sNssais = append(sNssais, *snssaiSmfInfo.SNssai)
	}

	// set nfProfile
	profile := models.NfProfile{
		NfInstanceId:  smf_context.SMF_Self().NfInstanceID,
		NfType:        models.NfType_SMF,
		NfStatus:      models.NfStatus_REGISTERED,
		Ipv4Addresses: []string{smf_context.SMF_Self().RegisterIPv4},
		NfServices:    smf_context.NFServices,
		SmfInfo:       smf_context.SmfInfo,
		SNssais:       &sNssais,
		PlmnList:      smf_context.SmfPlmnConfig(),
		AllowedPlmns:  smf_context.SmfPlmnConfig(),
	}

	var rep models.NfProfile
	var res *http.Response
	var err error

	// Check data (Use RESTful PUT)

	rep, res, err = smf_context.SMF_Self().
		NFManagementClient.
		NFInstanceIDDocumentApi.
		RegisterNFInstance(context.TODO(), smf_context.SMF_Self().NfInstanceID, profile)
	metrics.IncrementSvcNrfMsgStats(smf_context.SMF_Self().NfInstanceID, string(svcmsgtypes.NnrfNFRegister), "Out", "", "")

	if err != nil || res == nil {
		logger.ConsumerLog.Infof("SMF register to NRF Error[%s]", err.Error())
		metrics.IncrementSvcNrfMsgStats(smf_context.SMF_Self().NfInstanceID, string(svcmsgtypes.NnrfNFRegister), "In", "Failure", err.Error())
		return fmt.Errorf("NRF Registration failure")
	}

	if res != nil {
		defer func() {
			if resCloseErr := res.Body.Close(); resCloseErr != nil {
				logger.ConsumerLog.Errorf("RegisterNFInstance response body cannot close: %+v", resCloseErr)
			}
		}()
	}

	metrics.IncrementSvcNrfMsgStats(smf_context.SMF_Self().NfInstanceID, string(svcmsgtypes.NnrfNFRegister), "In", http.StatusText(res.StatusCode), "")

	status := res.StatusCode
	if status == http.StatusOK {
		// NFUpdate
		logger.ConsumerLog.Infof("NRF Registration success, status [%v]", http.StatusText(res.StatusCode))
	} else if status == http.StatusCreated {
		// NFRegister
		resourceUri := res.Header.Get("Location")
		// resouceNrfUri := resourceUri[strings.LastIndex(resourceUri, "/"):]
		smf_context.SMF_Self().NfInstanceID = resourceUri[strings.LastIndex(resourceUri, "/")+1:]
		logger.ConsumerLog.Infof("NRF Registration success, status [%v]", http.StatusText(res.StatusCode))
	} else {
		logger.ConsumerLog.Infof("handler returned wrong status code %d", status)
		// fmt.Errorf("NRF return wrong status code %d", status)
		logger.ConsumerLog.Errorf("NRF Registration failure, status [%v]", http.StatusText(res.StatusCode))
		return fmt.Errorf("NRF Registration failure, [%v]", http.StatusText(res.StatusCode))
	}

	logger.InitLog.Infof("SMF Registration to NRF %v", rep)
	return nil
}

func ReSendNFRegistration() {
	for {
		if err := SendNFRegistration(); err != nil {
			logger.ConsumerLog.Warnf("Send NFRegistration Failed, %v", err)
			time.Sleep(time.Second * 2)
			continue
		}
		return
	}
}

func SendNFDeregistration() error {
	// Check data (Use RESTful DELETE)
	res, localErr := smf_context.SMF_Self().
		NFManagementClient.
		NFInstanceIDDocumentApi.
		DeregisterNFInstance(context.TODO(), smf_context.SMF_Self().NfInstanceID)
	metrics.IncrementSvcNrfMsgStats(smf_context.SMF_Self().NfInstanceID, string(svcmsgtypes.NnrfNFDeRegister), "Out", "", "")
	if localErr != nil {
		logger.ConsumerLog.Warnln(localErr)
		metrics.IncrementSvcNrfMsgStats(smf_context.SMF_Self().NfInstanceID, string(svcmsgtypes.NnrfNFDeRegister), "In", "Failure", localErr.Error())
		return localErr
	}
	defer func() {
		if resCloseErr := res.Body.Close(); resCloseErr != nil {
			logger.ConsumerLog.Errorf("DeregisterNFInstance response body cannot close: %+v", resCloseErr)
		}
	}()
	if res != nil {
		metrics.IncrementSvcNrfMsgStats(smf_context.SMF_Self().NfInstanceID, string(svcmsgtypes.NnrfNFRegister), "In", http.StatusText(res.StatusCode), "")
		if status := res.StatusCode; status != http.StatusNoContent {
			logger.ConsumerLog.Warnln("handler returned wrong status code ", status)
			return openapi.ReportError("handler returned wrong status code %d", status)
		}
	}
	return nil
}

func SendNFDiscoveryUDM() (*models.ProblemDetails, error) {
	localVarOptionals := Nnrf_NFDiscovery.SearchNFInstancesParamOpts{}

	// Check data
	result, httpResp, localErr := smf_context.SMF_Self().
		NFDiscoveryClient.
		NFInstancesStoreApi.
		SearchNFInstances(context.TODO(), models.NfType_UDM, models.NfType_SMF, &localVarOptionals)
	metrics.IncrementSvcNrfMsgStats(smf_context.SMF_Self().NfInstanceID, string(svcmsgtypes.NnrfNFDiscoveryUdm), "Out", "", "")

	if localErr == nil {
		if result.NfInstances == nil {
			if status := httpResp.StatusCode; status != http.StatusOK {
				logger.ConsumerLog.Warnln("handler returned wrong status code", status)
			}

			logger.ConsumerLog.Warnln("NfInstances is nil")
			metrics.IncrementSvcNrfMsgStats(smf_context.SMF_Self().NfInstanceID, string(svcmsgtypes.NnrfNFDiscoveryUdm), "In", http.StatusText(httpResp.StatusCode), "NilInstance")
			return nil, openapi.ReportError("NfInstances is nil")
		}

		metrics.IncrementSvcNrfMsgStats(smf_context.SMF_Self().NfInstanceID, string(svcmsgtypes.NnrfNFDiscoveryUdm), "In", http.StatusText(httpResp.StatusCode), "")
		smf_context.SMF_Self().UDMProfile = result.NfInstances[0]

		for _, service := range *smf_context.SMF_Self().UDMProfile.NfServices {
			if service.ServiceName == models.ServiceName_NUDM_SDM {
				SDMConf := Nudm_SubscriberDataManagement.NewConfiguration()
				SDMConf.SetBasePath(service.ApiPrefix)
				smf_context.SMF_Self().SubscriberDataManagementClient = Nudm_SubscriberDataManagement.NewAPIClient(SDMConf)
			}
		}

		if smf_context.SMF_Self().SubscriberDataManagementClient == nil {
			logger.ConsumerLog.Warnln("sdm client failed")
		}
	} else if httpResp != nil {
		defer func() {
			if resCloseErr := httpResp.Body.Close(); resCloseErr != nil {
				logger.ConsumerLog.Errorf("SearchNFInstances response body cannot close: %+v", resCloseErr)
			}
		}()
		logger.ConsumerLog.Warnln("handler returned wrong status code ", httpResp.Status)
		if httpResp.Status != localErr.Error() {
			metrics.IncrementSvcNrfMsgStats(smf_context.SMF_Self().NfInstanceID, string(svcmsgtypes.NnrfNFDiscoveryUdm), "In", http.StatusText(httpResp.StatusCode), httpResp.Status)
			return nil, localErr
		}
		problem := localErr.(openapi.GenericOpenAPIError).Model().(models.ProblemDetails)
		metrics.IncrementSvcNrfMsgStats(smf_context.SMF_Self().NfInstanceID, string(svcmsgtypes.NnrfNFDiscoveryUdm), "In", http.StatusText(httpResp.StatusCode), localErr.Error())
		return &problem, nil
	} else {
		metrics.IncrementSvcNrfMsgStats(smf_context.SMF_Self().NfInstanceID, string(svcmsgtypes.NnrfNFDiscoveryUdm), "In", "Failure", "NoResponse")
		return nil, openapi.ReportError("server no response")
	}
	return nil, nil
}

func SendNFDiscoveryPCF() (problemDetails *models.ProblemDetails, err error) {
	// Set targetNfType
	targetNfType := models.NfType_PCF
	// Set requestNfType
	requesterNfType := models.NfType_SMF
	localVarOptionals := Nnrf_NFDiscovery.SearchNFInstancesParamOpts{}

	// Check data
	result, httpResp, localErr := smf_context.SMF_Self().
		NFDiscoveryClient.
		NFInstancesStoreApi.
		SearchNFInstances(context.TODO(), targetNfType, requesterNfType, &localVarOptionals)
	metrics.IncrementSvcNrfMsgStats(smf_context.SMF_Self().NfInstanceID, string(svcmsgtypes.NnrfNFDiscoveryPcf), "Out", "", "")

	if localErr == nil {
		logger.ConsumerLog.Traceln(result.NfInstances)
		metrics.IncrementSvcNrfMsgStats(smf_context.SMF_Self().NfInstanceID, string(svcmsgtypes.NnrfNFDiscoveryPcf), "In", http.StatusText(httpResp.StatusCode), "")
	} else if httpResp != nil {
		defer func() {
			if resCloseErr := httpResp.Body.Close(); resCloseErr != nil {
				logger.ConsumerLog.Errorf("SearchNFInstances response body cannot close: %+v", resCloseErr)
			}
		}()
		logger.ConsumerLog.Warnln("handler returned wrong status code ", httpResp.Status)
		if httpResp.Status != localErr.Error() {
			err = localErr
			metrics.IncrementSvcNrfMsgStats(smf_context.SMF_Self().NfInstanceID, string(svcmsgtypes.NnrfNFDiscoveryPcf), "In", http.StatusText(httpResp.StatusCode), httpResp.Status)
			return
		}
		metrics.IncrementSvcNrfMsgStats(smf_context.SMF_Self().NfInstanceID, string(svcmsgtypes.NnrfNFDiscoveryPcf), "In", http.StatusText(httpResp.StatusCode), localErr.Error())
		problem := localErr.(openapi.GenericOpenAPIError).Model().(models.ProblemDetails)
		problemDetails = &problem
	} else {
		metrics.IncrementSvcNrfMsgStats(smf_context.SMF_Self().NfInstanceID, string(svcmsgtypes.NnrfNFDiscoveryPcf), "In", "Failure", "NoResponse")
		err = openapi.ReportError("server no response")
	}

	return problemDetails, err
}

func SendNFDiscoveryServingAMF(smContext *smf_context.SMContext) (*models.ProblemDetails, error) {
	targetNfType := models.NfType_AMF
	requesterNfType := models.NfType_SMF

	localVarOptionals := Nnrf_NFDiscovery.SearchNFInstancesParamOpts{}

	localVarOptionals.TargetNfInstanceId = optional.NewInterface(smContext.ServingNfId)

	// Check data
	result, httpResp, localErr := smf_context.SMF_Self().
		NFDiscoveryClient.
		NFInstancesStoreApi.
		SearchNFInstances(context.TODO(), targetNfType, requesterNfType, &localVarOptionals)
	metrics.IncrementSvcNrfMsgStats(smf_context.SMF_Self().NfInstanceID, string(svcmsgtypes.NnrfNFDiscoveryAmf), "Out", "", "")

	if localErr == nil {
		if result.NfInstances == nil {
			if status := httpResp.StatusCode; status != http.StatusOK {
				logger.ConsumerLog.Warnln("handler returned wrong status code", status)
			}

			logger.ConsumerLog.Warnln("NfInstances is nil")
			metrics.IncrementSvcNrfMsgStats(smf_context.SMF_Self().NfInstanceID, string(svcmsgtypes.NnrfNFDiscoveryAmf), "In", http.StatusText(httpResp.StatusCode), "NilInstance")
			return nil, openapi.ReportError("NfInstances is nil")
		}
		smContext.SubConsumerLog.Info("send NF Discovery Serving AMF Successful")
		smContext.AMFProfile = deepcopy.Copy(result.NfInstances[0]).(models.NfProfile)
		metrics.IncrementSvcNrfMsgStats(smf_context.SMF_Self().NfInstanceID, string(svcmsgtypes.NnrfNFDiscoveryAmf), "In", http.StatusText(httpResp.StatusCode), "")
	} else if httpResp != nil {
		defer func() {
			if resCloseErr := httpResp; resCloseErr != nil {
				logger.ConsumerLog.Errorf("SearchNFInstances response body cannot close: %+v", resCloseErr)
			}
		}()
		if httpResp.Status != localErr.Error() {
			metrics.IncrementSvcNrfMsgStats(smf_context.SMF_Self().NfInstanceID, string(svcmsgtypes.NnrfNFDiscoveryAmf), "In", http.StatusText(httpResp.StatusCode), httpResp.Status)
			return nil, localErr
		}
		metrics.IncrementSvcNrfMsgStats(smf_context.SMF_Self().NfInstanceID, string(svcmsgtypes.NnrfNFDiscoveryAmf), "In", http.StatusText(httpResp.StatusCode), localErr.Error())
		problem := localErr.(openapi.GenericOpenAPIError).Model().(models.ProblemDetails)
		return &problem, nil
	} else {
		metrics.IncrementSvcNrfMsgStats(smf_context.SMF_Self().NfInstanceID, string(svcmsgtypes.NnrfNFDiscoveryAmf), "In", "Failure", "NoResponse")
		return nil, openapi.ReportError("server no response")
	}

	return nil, nil
}

func SendDeregisterNFInstance() (*models.ProblemDetails, error) {
	logger.ConsumerLog.Infof("Send Deregister NFInstance")

	smfSelf := smf_context.SMF_Self()
	// Set client and set url

	res, err := smfSelf.
		NFManagementClient.
		NFInstanceIDDocumentApi.
		DeregisterNFInstance(context.Background(), smfSelf.NfInstanceID)
	metrics.IncrementSvcNrfMsgStats(smfSelf.NfInstanceID, string(svcmsgtypes.NnrfNFInstanceDeRegister), "Out", "", "")
	if err == nil {
		metrics.IncrementSvcNrfMsgStats(smfSelf.NfInstanceID, string(svcmsgtypes.NnrfNFInstanceDeRegister), "In", http.StatusText(res.StatusCode), "")
		return nil, err
	} else if res != nil {
		defer func() {
			if resCloseErr := res.Body.Close(); resCloseErr != nil {
				logger.ConsumerLog.Errorf("DeregisterNFInstance response body cannot close: %+v", resCloseErr)
			}
		}()
		if res.Status != err.Error() {
			metrics.IncrementSvcNrfMsgStats(smfSelf.NfInstanceID, string(svcmsgtypes.NnrfNFInstanceDeRegister), "In", http.StatusText(res.StatusCode), res.Status)
			return nil, err
		}
		metrics.IncrementSvcNrfMsgStats(smfSelf.NfInstanceID, string(svcmsgtypes.NnrfNFInstanceDeRegister), "In", http.StatusText(res.StatusCode), err.Error())
		problem := err.(openapi.GenericOpenAPIError).Model().(models.ProblemDetails)
		return &problem, err
	} else {
		metrics.IncrementSvcNrfMsgStats(smfSelf.NfInstanceID, string(svcmsgtypes.NnrfNFInstanceDeRegister), "In", "Failure", "NoResponse")
		return nil, openapi.ReportError("server no response")
	}
}
