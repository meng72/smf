// SPDX-FileCopyrightText: 2021 Open Networking Foundation <info@opennetworking.org>
// Copyright 2019 free5GC.org
//
// SPDX-License-Identifier: Apache-2.0

package upf

import (
	"time"

	"github.com/free5gc/pfcp"
	"github.com/free5gc/smf/context"
	"github.com/free5gc/smf/logger"
	"github.com/free5gc/smf/metrics"
	"github.com/free5gc/smf/msgtypes/pfcpmsgtypes"
	"github.com/free5gc/smf/pfcp/message"
)

const (
	maxHeartbeatRetry        = 3  //sec
	maxHeartbeatInterval     = 10 //sec
	maxUpfProbeRetryInterval = 10 //sec
)

func InitPfcpHeartbeatRequest(userplane *context.UserPlaneInformation) {
	//Iterate through all UPFs and send heartbeat to active UPFs
	for {
		time.Sleep(maxHeartbeatInterval * time.Second)
		for _, upf := range userplane.UPFs {
			upf.UPF.UpfLock.Lock()
			if (upf.UPF.UPFStatus == context.AssociatedSetUpSuccess) && upf.UPF.NHeartBeat < maxHeartbeatRetry {
				err := message.SendHeartbeatRequest(upf.NodeID)
				if err != nil {
					logger.PfcpLog.Errorf("Send PFCP Heartbeat Request failed: %v for UPF: %v", err, upf.NodeID)
				} else {
					upf.UPF.NHeartBeat++
				}
			} else if upf.UPF.NHeartBeat == maxHeartbeatRetry {
				metrics.IncrementN4MsgStats(context.SMF_Self().NfInstanceID, pfcpmsgtypes.PfcpMsgTypeString(pfcp.PFCP_HEARTBEAT_REQUEST), "Out", "Failure", "Timeout")
				upf.UPF.UPFStatus = context.NotAssociated
			}
			upf.UPF.UpfLock.Unlock()
		}
	}
}

func ProbeInactiveUpfs(upfs *context.UserPlaneInformation) {
	//Iterate through all UPFs and send PFCP request to inactive UPFs
	for {
		time.Sleep(maxUpfProbeRetryInterval * time.Second)
		for _, upf := range upfs.UPFs {
			if upf.UPF.UPFStatus == context.NotAssociated {
				message.SendPfcpAssociationSetupRequest(upf.NodeID)
			}
		}
	}
}
