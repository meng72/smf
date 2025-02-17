// SPDX-FileCopyrightText: 2021 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package pfcpmsgtypes

import (
	"github.com/free5gc/pfcp"
)

var msgTypeText = map[pfcp.MessageType]string{
	pfcp.PFCP_HEARTBEAT_REQUEST:              "PFCP_HEARTBEAT_REQUEST",
	pfcp.PFCP_HEARTBEAT_RESPONSE:             "PFCP_HEARTBEAT_RESPONSE",
	pfcp.PFCP_PFD_MANAGEMENT_REQUEST:         "PFCP_PFD_MANAGEMENT_REQUEST",
	pfcp.PFCP_PFD_MANAGEMENT_RESPONSE:        "PFCP_PFD_MANAGEMENT_RESPONSE",
	pfcp.PFCP_ASSOCIATION_SETUP_REQUEST:      "PFCP_ASSOCIATION_SETUP_REQUEST",
	pfcp.PFCP_ASSOCIATION_SETUP_RESPONSE:     "PFCP_ASSOCIATION_SETUP_RESPONSE",
	pfcp.PFCP_ASSOCIATION_UPDATE_REQUEST:     "PFCP_ASSOCIATION_UPDATE_REQUEST",
	pfcp.PFCP_ASSOCIATION_UPDATE_RESPONSE:    "PFCP_ASSOCIATION_UPDATE_RESPONSE",
	pfcp.PFCP_ASSOCIATION_RELEASE_REQUEST:    "PFCP_ASSOCIATION_RELEASE_REQUEST",
	pfcp.PFCP_ASSOCIATION_RELEASE_RESPONSE:   "PFCP_ASSOCIATION_RELEASE_RESPONSE",
	pfcp.PFCP_VERSION_NOT_SUPPORTED_RESPONSE: "PFCP_VERSION_NOT_SUPPORTED_RESPONSE",
	pfcp.PFCP_NODE_REPORT_REQUEST:            "PFCP_NODE_REPORT_REQUEST",
	pfcp.PFCP_NODE_REPORT_RESPONSE:           "PFCP_NODE_REPORT_RESPONSE",
	pfcp.PFCP_SESSION_SET_DELETION_REQUEST:   "PFCP_SESSION_SET_DELETION_REQUEST",
	pfcp.PFCP_SESSION_SET_DELETION_RESPONSE:  "PFCP_SESSION_SET_DELETION_RESPONSE",

	pfcp.PFCP_SESSION_ESTABLISHMENT_REQUEST:  "PFCP_SESSION_ESTABLISHMENT_REQUEST",
	pfcp.PFCP_SESSION_ESTABLISHMENT_RESPONSE: "PFCP_SESSION_ESTABLISHMENT_RESPONSE",
	pfcp.PFCP_SESSION_MODIFICATION_REQUEST:   "PFCP_SESSION_MODIFICATION_REQUEST",
	pfcp.PFCP_SESSION_MODIFICATION_RESPONSE:  "PFCP_SESSION_MODIFICATION_RESPONSE",
	pfcp.PFCP_SESSION_DELETION_REQUEST:       "PFCP_SESSION_DELETION_REQUEST",
	pfcp.PFCP_SESSION_DELETION_RESPONSE:      "PFCP_SESSION_DELETION_RESPONSE",
	pfcp.PFCP_SESSION_REPORT_REQUEST:         "PFCP_SESSION_REPORT_REQUEST",
	pfcp.PFCP_SESSION_REPORT_RESPONSE:        "PFCP_SESSION_REPORT_RESPONSE",
}

func PfcpMsgTypeString(code pfcp.MessageType) string {
	return msgTypeText[code]
}
