// Code generated by "stringer -type Ident"; DO NOT EDIT.

package mqtt

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[PayloadFormatIndicator-1]
	_ = x[MessageExpiryInterval-2]
	_ = x[ContentType-3]
	_ = x[ResponseTopic-8]
	_ = x[CorrelationData-9]
	_ = x[SubIdent-11]
	_ = x[SessionExpiryInterval-17]
	_ = x[AssignedClientIdent-18]
	_ = x[ServerKeepAlive-19]
	_ = x[AuthMethod-21]
	_ = x[AuthData-22]
	_ = x[RequestProblemInfo-23]
	_ = x[WillDelayInterval-24]
	_ = x[RequestResponseInfo-25]
	_ = x[ResponseInformation-26]
	_ = x[ServerReference-28]
	_ = x[ReasonString-31]
	_ = x[ReceiveMax-33]
	_ = x[TopicAliasMax-34]
	_ = x[TopicAlias-35]
	_ = x[MaximumQoS-36]
	_ = x[RetainAvailable-37]
	_ = x[UserProperty-38]
	_ = x[MaxPacketSize-39]
	_ = x[WildcardSubAvailable-40]
	_ = x[SubIdentAvailable-41]
	_ = x[SharedSubAvailable-48]
}

const (
	_Ident_name_0 = "PayloadFormatIndicatorMessageExpiryIntervalContentType"
	_Ident_name_1 = "ResponseTopicCorrelationData"
	_Ident_name_2 = "SubIdent"
	_Ident_name_3 = "SessionExpiryIntervalAssignedClientIdentServerKeepAlive"
	_Ident_name_4 = "AuthMethodAuthDataRequestProblemInfoWillDelayIntervalRequestResponseInfoResponseInformation"
	_Ident_name_5 = "ServerReference"
	_Ident_name_6 = "ReasonString"
	_Ident_name_7 = "ReceiveMaxTopicAliasMaxTopicAliasMaximumQoSRetainAvailableUserPropertyMaxPacketSizeWildcardSubAvailableSubIdentAvailable"
	_Ident_name_8 = "SharedSubAvailable"
)

var (
	_Ident_index_0 = [...]uint8{0, 22, 43, 54}
	_Ident_index_1 = [...]uint8{0, 13, 28}
	_Ident_index_3 = [...]uint8{0, 21, 40, 55}
	_Ident_index_4 = [...]uint8{0, 10, 18, 36, 53, 72, 91}
	_Ident_index_7 = [...]uint8{0, 10, 23, 33, 43, 58, 70, 83, 103, 120}
)

func (i Ident) String() string {
	switch {
	case 1 <= i && i <= 3:
		i -= 1
		return _Ident_name_0[_Ident_index_0[i]:_Ident_index_0[i+1]]
	case 8 <= i && i <= 9:
		i -= 8
		return _Ident_name_1[_Ident_index_1[i]:_Ident_index_1[i+1]]
	case i == 11:
		return _Ident_name_2
	case 17 <= i && i <= 19:
		i -= 17
		return _Ident_name_3[_Ident_index_3[i]:_Ident_index_3[i+1]]
	case 21 <= i && i <= 26:
		i -= 21
		return _Ident_name_4[_Ident_index_4[i]:_Ident_index_4[i+1]]
	case i == 28:
		return _Ident_name_5
	case i == 31:
		return _Ident_name_6
	case 33 <= i && i <= 41:
		i -= 33
		return _Ident_name_7[_Ident_index_7[i]:_Ident_index_7[i+1]]
	case i == 48:
		return _Ident_name_8
	default:
		return "Ident(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}