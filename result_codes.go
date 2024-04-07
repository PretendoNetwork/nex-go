package nex

import (
	"reflect"
	"strconv"
)

var errorMask = 1 << 31

type resultCodes struct {
	Core struct {
		Unknown               uint32
		NotImplemented        uint32
		InvalidPointer        uint32
		OperationAborted      uint32
		Exception             uint32
		AccessDenied          uint32
		InvalidHandle         uint32
		InvalidIndex          uint32
		OutOfMemory           uint32
		InvalidArgument       uint32
		Timeout               uint32
		InitializationFailure uint32
		CallInitiationFailure uint32
		RegistrationError     uint32
		BufferOverflow        uint32
		InvalidLockState      uint32
		InvalidSequence       uint32
		SystemError           uint32
		Cancelled             uint32
	}

	DDL struct {
		InvalidSignature uint32
		IncorrectVersion uint32
	}

	RendezVous struct {
		ConnectionFailure                        uint32
		NotAuthenticated                         uint32
		InvalidUsername                          uint32
		InvalidPassword                          uint32
		UsernameAlreadyExists                    uint32
		AccountDisabled                          uint32
		AccountExpired                           uint32
		ConcurrentLoginDenied                    uint32
		EncryptionFailure                        uint32
		InvalidPID                               uint32
		MaxConnectionsReached                    uint32
		InvalidGID                               uint32
		InvalidControlScriptID                   uint32
		InvalidOperationInLiveEnvironment        uint32
		DuplicateEntry                           uint32
		ControlScriptFailure                     uint32
		ClassNotFound                            uint32
		SessionVoid                              uint32
		DDLMismatch                              uint32
		InvalidConfiguration                     uint32
		SessionFull                              uint32
		InvalidGatheringPassword                 uint32
		WithoutParticipationPeriod               uint32
		PersistentGatheringCreationMax           uint32
		PersistentGatheringParticipationMax      uint32
		DeniedByParticipants                     uint32
		ParticipantInBlackList                   uint32
		GameServerMaintenance                    uint32
		OperationPostpone                        uint32
		OutOfRatingRange                         uint32
		ConnectionDisconnected                   uint32
		InvalidOperation                         uint32
		NotParticipatedGathering                 uint32
		MatchmakeSessionUserPasswordUnmatch      uint32
		MatchmakeSessionSystemPasswordUnmatch    uint32
		UserIsOffline                            uint32
		AlreadyParticipatedGathering             uint32
		PermissionDenied                         uint32
		NotFriend                                uint32
		SessionClosed                            uint32
		DatabaseTemporarilyUnavailable           uint32
		InvalidUniqueID                          uint32
		MatchmakingWithdrawn                     uint32
		LimitExceeded                            uint32
		AccountTemporarilyDisabled               uint32
		PartiallyServiceClosed                   uint32
		ConnectionDisconnectedForConcurrentLogin uint32
	}

	PythonCore struct {
		Exception        uint32
		TypeError        uint32
		IndexError       uint32
		InvalidReference uint32
		CallFailure      uint32
		MemoryError      uint32
		KeyError         uint32
		OperationError   uint32
		ConversionError  uint32
		ValidationError  uint32
	}

	Transport struct {
		Unknown                       uint32
		ConnectionFailure             uint32
		InvalidURL                    uint32
		InvalidKey                    uint32
		InvalidURLType                uint32
		DuplicateEndpoint             uint32
		IOError                       uint32
		Timeout                       uint32
		ConnectionReset               uint32
		IncorrectRemoteAuthentication uint32
		ServerRequestError            uint32
		DecompressionFailure          uint32
		ReliableSendBufferFullFatal   uint32
		UPnPCannotInit                uint32
		UPnPCannotAddMapping          uint32
		NatPMPCannotInit              uint32
		NatPMPCannotAddMapping        uint32
		UnsupportedNAT                uint32
		DNSError                      uint32
		ProxyError                    uint32
		DataRemaining                 uint32
		NoBuffer                      uint32
		NotFound                      uint32
		TemporaryServerError          uint32
		PermanentServerError          uint32
		ServiceUnavailable            uint32
		ReliableSendBufferFull        uint32
		InvalidStation                uint32
		InvalidSubStreamID            uint32
		PacketBufferFull              uint32
		NatTraversalError             uint32
		NatCheckError                 uint32
	}

	DOCore struct {
		StationNotReached             uint32
		TargetStationDisconnect       uint32
		LocalStationLeaving           uint32
		ObjectNotFound                uint32
		InvalidRole                   uint32
		CallTimeout                   uint32
		RMCDispatchFailed             uint32
		MigrationInProgress           uint32
		NoAuthority                   uint32
		NoTargetStationSpecified      uint32
		JoinFailed                    uint32
		JoinDenied                    uint32
		ConnectivityTestFailed        uint32
		Unknown                       uint32
		UnfreedReferences             uint32
		JobTerminationFailed          uint32
		InvalidState                  uint32
		FaultRecoveryFatal            uint32
		FaultRecoveryJobProcessFailed uint32
		StationInconsitency           uint32
		AbnormalMasterState           uint32
		VersionMismatch               uint32
	}

	FPD struct {
		NotInitialized               uint32
		AlreadyInitialized           uint32
		NotConnected                 uint32
		Connected                    uint32
		InitializationFailure        uint32
		OutOfMemory                  uint32
		RmcFailed                    uint32
		InvalidArgument              uint32
		InvalidLocalAccountID        uint32
		InvalidPrincipalID           uint32
		InvalidLocalFriendCode       uint32
		LocalAccountNotExists        uint32
		LocalAccountNotLoaded        uint32
		LocalAccountAlreadyLoaded    uint32
		FriendAlreadyExists          uint32
		FriendNotExists              uint32
		FriendNumMax                 uint32
		NotFriend                    uint32
		FileIO                       uint32
		P2PInternetProhibited        uint32
		Unknown                      uint32
		InvalidState                 uint32
		AddFriendProhibited          uint32
		InvalidAccount               uint32
		BlacklistedByMe              uint32
		FriendAlreadyAdded           uint32
		MyFriendListLimitExceed      uint32
		RequestLimitExceed           uint32
		InvalidMessageID             uint32
		MessageIsNotMine             uint32
		MessageIsNotForMe            uint32
		FriendRequestBlocked         uint32
		NotInMyFriendList            uint32
		FriendListedByMe             uint32
		NotInMyBlacklist             uint32
		IncompatibleAccount          uint32
		BlockSettingChangeNotAllowed uint32
		SizeLimitExceeded            uint32
		OperationNotAllowed          uint32
		NotNetworkAccount            uint32
		NotificationNotFound         uint32
		PreferenceNotInitialized     uint32
		FriendRequestNotAllowed      uint32
	}

	Ranking struct {
		NotInitialized    uint32
		InvalidArgument   uint32
		RegistrationError uint32
		NotFound          uint32
		InvalidScore      uint32
		InvalidDataSize   uint32
		PermissionDenied  uint32
		Unknown           uint32
		NotImplemented    uint32
	}

	Authentication struct {
		NASAuthenticateError             uint32
		TokenParseError                  uint32
		HTTPConnectionError              uint32
		HTTPDNSError                     uint32
		HTTPGetProxySetting              uint32
		TokenExpired                     uint32
		ValidationFailed                 uint32
		InvalidParam                     uint32
		PrincipalIDUnmatched             uint32
		MoveCountUnmatch                 uint32
		UnderMaintenance                 uint32
		UnsupportedVersion               uint32
		ServerVersionIsOld               uint32
		Unknown                          uint32
		ClientVersionIsOld               uint32
		AccountLibraryError              uint32
		ServiceNoLongerAvailable         uint32
		UnknownApplication               uint32
		ApplicationVersionIsOld          uint32
		OutOfService                     uint32
		NetworkServiceLicenseRequired    uint32
		NetworkServiceLicenseSystemError uint32
		NetworkServiceLicenseError3      uint32
		NetworkServiceLicenseError4      uint32
	}

	DataStore struct {
		Unknown             uint32
		InvalidArgument     uint32
		PermissionDenied    uint32
		NotFound            uint32
		AlreadyLocked       uint32
		UnderReviewing      uint32
		Expired             uint32
		InvalidCheckToken   uint32
		SystemFileError     uint32
		OverCapacity        uint32
		OperationNotAllowed uint32
		InvalidPassword     uint32
		ValueNotEqual       uint32
	}

	ServiceItem struct {
		Unknown                  uint32
		InvalidArgument          uint32
		EShopUnknownHTTPError    uint32
		EShopResponseParseError  uint32
		NotOwned                 uint32
		InvalidLimitationType    uint32
		ConsumptionRightShortage uint32
	}

	MatchmakeReferee struct {
		Unknown                  uint32
		InvalidArgument          uint32
		AlreadyExists            uint32
		NotParticipatedGathering uint32
		NotParticipatedRound     uint32
		StatsNotFound            uint32
		RoundNotFound            uint32
		RoundArbitrated          uint32
		RoundNotArbitrated       uint32
	}

	Subscriber struct {
		Unknown          uint32
		InvalidArgument  uint32
		OverLimit        uint32
		PermissionDenied uint32
	}

	Ranking2 struct {
		Unknown         uint32
		InvalidArgument uint32
		InvalidScore    uint32
	}

	SmartDeviceVoiceChat struct {
		Unknown                       uint32
		InvalidArgument               uint32
		InvalidResponse               uint32
		InvalidAccessToken            uint32
		Unauthorized                  uint32
		AccessError                   uint32
		UserNotFound                  uint32
		RoomNotFound                  uint32
		RoomNotActivated              uint32
		ApplicationNotSupported       uint32
		InternalServerError           uint32
		ServiceUnavailable            uint32
		UnexpectedError               uint32
		UnderMaintenance              uint32
		ServiceNoLongerAvailable      uint32
		AccountTemporarilyDisabled    uint32
		PermissionDenied              uint32
		NetworkServiceLicenseRequired uint32
		AccountLibraryError           uint32
		GameModeNotFound              uint32
	}

	Screening struct {
		Unknown         uint32
		InvalidArgument uint32
		NotFound        uint32
	}

	Custom struct {
		Unknown uint32
	}

	Ess struct {
		Unknown                uint32
		GameSessionError       uint32
		GameSessionMaintenance uint32
	}
}

// ResultNames contains a map of all the result code string names, indexed by the result code
var ResultNames = map[uint32]string{}

// ResultCodes provides a struct containing RDV result codes using dot-notation
var ResultCodes resultCodes

func initResultCodes() {
	ResultCodes.Core.Unknown = 0x00010001
	ResultCodes.Core.NotImplemented = 0x00010002
	ResultCodes.Core.InvalidPointer = 0x00010003
	ResultCodes.Core.OperationAborted = 0x00010004
	ResultCodes.Core.Exception = 0x00010005
	ResultCodes.Core.AccessDenied = 0x00010006
	ResultCodes.Core.InvalidHandle = 0x00010007
	ResultCodes.Core.InvalidIndex = 0x00010008
	ResultCodes.Core.OutOfMemory = 0x00010009
	ResultCodes.Core.InvalidArgument = 0x0001000A
	ResultCodes.Core.Timeout = 0x0001000B
	ResultCodes.Core.InitializationFailure = 0x0001000C
	ResultCodes.Core.CallInitiationFailure = 0x0001000D
	ResultCodes.Core.RegistrationError = 0x0001000E
	ResultCodes.Core.BufferOverflow = 0x0001000F
	ResultCodes.Core.InvalidLockState = 0x00010010
	ResultCodes.Core.InvalidSequence = 0x00010011
	ResultCodes.Core.SystemError = 0x00010012
	ResultCodes.Core.Cancelled = 0x00010013

	ResultCodes.DDL.InvalidSignature = 0x00020001
	ResultCodes.DDL.IncorrectVersion = 0x00020002

	ResultCodes.RendezVous.ConnectionFailure = 0x00030001
	ResultCodes.RendezVous.NotAuthenticated = 0x00030002
	ResultCodes.RendezVous.InvalidUsername = 0x00030064
	ResultCodes.RendezVous.InvalidPassword = 0x00030065
	ResultCodes.RendezVous.UsernameAlreadyExists = 0x00030066
	ResultCodes.RendezVous.AccountDisabled = 0x00030067
	ResultCodes.RendezVous.AccountExpired = 0x00030068
	ResultCodes.RendezVous.ConcurrentLoginDenied = 0x00030069
	ResultCodes.RendezVous.EncryptionFailure = 0x0003006A
	ResultCodes.RendezVous.InvalidPID = 0x0003006B
	ResultCodes.RendezVous.MaxConnectionsReached = 0x0003006C
	ResultCodes.RendezVous.InvalidGID = 0x0003006D
	ResultCodes.RendezVous.InvalidControlScriptID = 0x0003006E
	ResultCodes.RendezVous.InvalidOperationInLiveEnvironment = 0x0003006F
	ResultCodes.RendezVous.DuplicateEntry = 0x00030070
	ResultCodes.RendezVous.ControlScriptFailure = 0x00030071
	ResultCodes.RendezVous.ClassNotFound = 0x00030072
	ResultCodes.RendezVous.SessionVoid = 0x00030073
	ResultCodes.RendezVous.DDLMismatch = 0x00030075
	ResultCodes.RendezVous.InvalidConfiguration = 0x00030076
	ResultCodes.RendezVous.SessionFull = 0x000300C8
	ResultCodes.RendezVous.InvalidGatheringPassword = 0x000300C9
	ResultCodes.RendezVous.WithoutParticipationPeriod = 0x000300CA
	ResultCodes.RendezVous.PersistentGatheringCreationMax = 0x000300CB
	ResultCodes.RendezVous.PersistentGatheringParticipationMax = 0x000300CC
	ResultCodes.RendezVous.DeniedByParticipants = 0x000300CD
	ResultCodes.RendezVous.ParticipantInBlackList = 0x000300CE
	ResultCodes.RendezVous.GameServerMaintenance = 0x000300CF
	ResultCodes.RendezVous.OperationPostpone = 0x000300D0
	ResultCodes.RendezVous.OutOfRatingRange = 0x000300D1
	ResultCodes.RendezVous.ConnectionDisconnected = 0x000300D2
	ResultCodes.RendezVous.InvalidOperation = 0x000300D3
	ResultCodes.RendezVous.NotParticipatedGathering = 0x000300D4
	ResultCodes.RendezVous.MatchmakeSessionUserPasswordUnmatch = 0x000300D5
	ResultCodes.RendezVous.MatchmakeSessionSystemPasswordUnmatch = 0x000300D6
	ResultCodes.RendezVous.UserIsOffline = 0x000300D7
	ResultCodes.RendezVous.AlreadyParticipatedGathering = 0x000300D8
	ResultCodes.RendezVous.PermissionDenied = 0x000300D9
	ResultCodes.RendezVous.NotFriend = 0x000300DA
	ResultCodes.RendezVous.SessionClosed = 0x000300DB
	ResultCodes.RendezVous.DatabaseTemporarilyUnavailable = 0x000300DC
	ResultCodes.RendezVous.InvalidUniqueID = 0x000300DD
	ResultCodes.RendezVous.MatchmakingWithdrawn = 0x000300DE
	ResultCodes.RendezVous.LimitExceeded = 0x000300DF
	ResultCodes.RendezVous.AccountTemporarilyDisabled = 0x000300E0
	ResultCodes.RendezVous.PartiallyServiceClosed = 0x000300E1
	ResultCodes.RendezVous.ConnectionDisconnectedForConcurrentLogin = 0x000300E2

	ResultCodes.PythonCore.Exception = 0x00040001
	ResultCodes.PythonCore.TypeError = 0x00040002
	ResultCodes.PythonCore.IndexError = 0x00040003
	ResultCodes.PythonCore.InvalidReference = 0x00040004
	ResultCodes.PythonCore.CallFailure = 0x00040005
	ResultCodes.PythonCore.MemoryError = 0x00040006
	ResultCodes.PythonCore.KeyError = 0x00040007
	ResultCodes.PythonCore.OperationError = 0x00040008
	ResultCodes.PythonCore.ConversionError = 0x00040009
	ResultCodes.PythonCore.ValidationError = 0x0004000A

	ResultCodes.Transport.Unknown = 0x00050001
	ResultCodes.Transport.ConnectionFailure = 0x00050002
	ResultCodes.Transport.InvalidURL = 0x00050003
	ResultCodes.Transport.InvalidKey = 0x00050004
	ResultCodes.Transport.InvalidURLType = 0x00050005
	ResultCodes.Transport.DuplicateEndpoint = 0x00050006
	ResultCodes.Transport.IOError = 0x00050007
	ResultCodes.Transport.Timeout = 0x00050008
	ResultCodes.Transport.ConnectionReset = 0x00050009
	ResultCodes.Transport.IncorrectRemoteAuthentication = 0x0005000A
	ResultCodes.Transport.ServerRequestError = 0x0005000B
	ResultCodes.Transport.DecompressionFailure = 0x0005000C
	ResultCodes.Transport.ReliableSendBufferFullFatal = 0x0005000D
	ResultCodes.Transport.UPnPCannotInit = 0x0005000E
	ResultCodes.Transport.UPnPCannotAddMapping = 0x0005000F
	ResultCodes.Transport.NatPMPCannotInit = 0x00050010
	ResultCodes.Transport.NatPMPCannotAddMapping = 0x00050011
	ResultCodes.Transport.UnsupportedNAT = 0x00050013
	ResultCodes.Transport.DNSError = 0x00050014
	ResultCodes.Transport.ProxyError = 0x00050015
	ResultCodes.Transport.DataRemaining = 0x00050016
	ResultCodes.Transport.NoBuffer = 0x00050017
	ResultCodes.Transport.NotFound = 0x00050018
	ResultCodes.Transport.TemporaryServerError = 0x00050019
	ResultCodes.Transport.PermanentServerError = 0x0005001A
	ResultCodes.Transport.ServiceUnavailable = 0x0005001B
	ResultCodes.Transport.ReliableSendBufferFull = 0x0005001C
	ResultCodes.Transport.InvalidStation = 0x0005001D
	ResultCodes.Transport.InvalidSubStreamID = 0x0005001E
	ResultCodes.Transport.PacketBufferFull = 0x0005001F
	ResultCodes.Transport.NatTraversalError = 0x00050020
	ResultCodes.Transport.NatCheckError = 0x00050021

	ResultCodes.DOCore.StationNotReached = 0x00060001
	ResultCodes.DOCore.TargetStationDisconnect = 0x00060002
	ResultCodes.DOCore.LocalStationLeaving = 0x00060003
	ResultCodes.DOCore.ObjectNotFound = 0x00060004
	ResultCodes.DOCore.InvalidRole = 0x00060005
	ResultCodes.DOCore.CallTimeout = 0x00060006
	ResultCodes.DOCore.RMCDispatchFailed = 0x00060007
	ResultCodes.DOCore.MigrationInProgress = 0x00060008
	ResultCodes.DOCore.NoAuthority = 0x00060009
	ResultCodes.DOCore.NoTargetStationSpecified = 0x0006000A
	ResultCodes.DOCore.JoinFailed = 0x0006000B
	ResultCodes.DOCore.JoinDenied = 0x0006000C
	ResultCodes.DOCore.ConnectivityTestFailed = 0x0006000D
	ResultCodes.DOCore.Unknown = 0x0006000E
	ResultCodes.DOCore.UnfreedReferences = 0x0006000F
	ResultCodes.DOCore.JobTerminationFailed = 0x00060010
	ResultCodes.DOCore.InvalidState = 0x00060011
	ResultCodes.DOCore.FaultRecoveryFatal = 0x00060012
	ResultCodes.DOCore.FaultRecoveryJobProcessFailed = 0x00060013
	ResultCodes.DOCore.StationInconsitency = 0x00060014
	ResultCodes.DOCore.AbnormalMasterState = 0x00060015
	ResultCodes.DOCore.VersionMismatch = 0x00060016

	ResultCodes.FPD.NotInitialized = 0x00650000
	ResultCodes.FPD.AlreadyInitialized = 0x00650001
	ResultCodes.FPD.NotConnected = 0x00650002
	ResultCodes.FPD.Connected = 0x00650003
	ResultCodes.FPD.InitializationFailure = 0x00650004
	ResultCodes.FPD.OutOfMemory = 0x00650005
	ResultCodes.FPD.RmcFailed = 0x00650006
	ResultCodes.FPD.InvalidArgument = 0x00650007
	ResultCodes.FPD.InvalidLocalAccountID = 0x00650008
	ResultCodes.FPD.InvalidPrincipalID = 0x00650009
	ResultCodes.FPD.InvalidLocalFriendCode = 0x0065000A
	ResultCodes.FPD.LocalAccountNotExists = 0x0065000B
	ResultCodes.FPD.LocalAccountNotLoaded = 0x0065000C
	ResultCodes.FPD.LocalAccountAlreadyLoaded = 0x0065000D
	ResultCodes.FPD.FriendAlreadyExists = 0x0065000E
	ResultCodes.FPD.FriendNotExists = 0x0065000F
	ResultCodes.FPD.FriendNumMax = 0x00650010
	ResultCodes.FPD.NotFriend = 0x00650011
	ResultCodes.FPD.FileIO = 0x00650012
	ResultCodes.FPD.P2PInternetProhibited = 0x00650013
	ResultCodes.FPD.Unknown = 0x00650014
	ResultCodes.FPD.InvalidState = 0x00650015
	ResultCodes.FPD.AddFriendProhibited = 0x00650017
	ResultCodes.FPD.InvalidAccount = 0x00650019
	ResultCodes.FPD.BlacklistedByMe = 0x0065001A
	ResultCodes.FPD.FriendAlreadyAdded = 0x0065001C
	ResultCodes.FPD.MyFriendListLimitExceed = 0x0065001D
	ResultCodes.FPD.RequestLimitExceed = 0x0065001E
	ResultCodes.FPD.InvalidMessageID = 0x0065001F
	ResultCodes.FPD.MessageIsNotMine = 0x00650020
	ResultCodes.FPD.MessageIsNotForMe = 0x00650021
	ResultCodes.FPD.FriendRequestBlocked = 0x00650022
	ResultCodes.FPD.NotInMyFriendList = 0x00650023
	ResultCodes.FPD.FriendListedByMe = 0x00650024
	ResultCodes.FPD.NotInMyBlacklist = 0x00650025
	ResultCodes.FPD.IncompatibleAccount = 0x00650026
	ResultCodes.FPD.BlockSettingChangeNotAllowed = 0x00650027
	ResultCodes.FPD.SizeLimitExceeded = 0x00650028
	ResultCodes.FPD.OperationNotAllowed = 0x00650029
	ResultCodes.FPD.NotNetworkAccount = 0x0065002A
	ResultCodes.FPD.NotificationNotFound = 0x0065002B
	ResultCodes.FPD.PreferenceNotInitialized = 0x0065002C
	ResultCodes.FPD.FriendRequestNotAllowed = 0x0065002D

	ResultCodes.Ranking.NotInitialized = 0x00670001
	ResultCodes.Ranking.InvalidArgument = 0x00670002
	ResultCodes.Ranking.RegistrationError = 0x00670003
	ResultCodes.Ranking.NotFound = 0x00670005
	ResultCodes.Ranking.InvalidScore = 0x00670006
	ResultCodes.Ranking.InvalidDataSize = 0x00670007
	ResultCodes.Ranking.PermissionDenied = 0x00670009
	ResultCodes.Ranking.Unknown = 0x0067000A
	ResultCodes.Ranking.NotImplemented = 0x0067000B

	ResultCodes.Authentication.NASAuthenticateError = 0x00680001
	ResultCodes.Authentication.TokenParseError = 0x00680002
	ResultCodes.Authentication.HTTPConnectionError = 0x00680003
	ResultCodes.Authentication.HTTPDNSError = 0x00680004
	ResultCodes.Authentication.HTTPGetProxySetting = 0x00680005
	ResultCodes.Authentication.TokenExpired = 0x00680006
	ResultCodes.Authentication.ValidationFailed = 0x00680007
	ResultCodes.Authentication.InvalidParam = 0x00680008
	ResultCodes.Authentication.PrincipalIDUnmatched = 0x00680009
	ResultCodes.Authentication.MoveCountUnmatch = 0x0068000A
	ResultCodes.Authentication.UnderMaintenance = 0x0068000B
	ResultCodes.Authentication.UnsupportedVersion = 0x0068000C
	ResultCodes.Authentication.ServerVersionIsOld = 0x0068000D
	ResultCodes.Authentication.Unknown = 0x0068000E
	ResultCodes.Authentication.ClientVersionIsOld = 0x0068000F
	ResultCodes.Authentication.AccountLibraryError = 0x00680010
	ResultCodes.Authentication.ServiceNoLongerAvailable = 0x00680011
	ResultCodes.Authentication.UnknownApplication = 0x00680012
	ResultCodes.Authentication.ApplicationVersionIsOld = 0x00680013
	ResultCodes.Authentication.OutOfService = 0x00680014
	ResultCodes.Authentication.NetworkServiceLicenseRequired = 0x00680015
	ResultCodes.Authentication.NetworkServiceLicenseSystemError = 0x00680016
	ResultCodes.Authentication.NetworkServiceLicenseError3 = 0x00680017
	ResultCodes.Authentication.NetworkServiceLicenseError4 = 0x00680018

	ResultCodes.DataStore.Unknown = 0x00690001
	ResultCodes.DataStore.InvalidArgument = 0x00690002
	ResultCodes.DataStore.PermissionDenied = 0x00690003
	ResultCodes.DataStore.NotFound = 0x00690004
	ResultCodes.DataStore.AlreadyLocked = 0x00690005
	ResultCodes.DataStore.UnderReviewing = 0x00690006
	ResultCodes.DataStore.Expired = 0x00690007
	ResultCodes.DataStore.InvalidCheckToken = 0x00690008
	ResultCodes.DataStore.SystemFileError = 0x00690009
	ResultCodes.DataStore.OverCapacity = 0x0069000A
	ResultCodes.DataStore.OperationNotAllowed = 0x0069000B
	ResultCodes.DataStore.InvalidPassword = 0x0069000C
	ResultCodes.DataStore.ValueNotEqual = 0x0069000D

	ResultCodes.ServiceItem.Unknown = 0x006C0001
	ResultCodes.ServiceItem.InvalidArgument = 0x006C0002
	ResultCodes.ServiceItem.EShopUnknownHTTPError = 0x006C0003
	ResultCodes.ServiceItem.EShopResponseParseError = 0x006C0004
	ResultCodes.ServiceItem.NotOwned = 0x006C0005
	ResultCodes.ServiceItem.InvalidLimitationType = 0x006C0006
	ResultCodes.ServiceItem.ConsumptionRightShortage = 0x006C0007

	ResultCodes.MatchmakeReferee.Unknown = 0x006F0001
	ResultCodes.MatchmakeReferee.InvalidArgument = 0x006F0002
	ResultCodes.MatchmakeReferee.AlreadyExists = 0x006F0003
	ResultCodes.MatchmakeReferee.NotParticipatedGathering = 0x006F0004
	ResultCodes.MatchmakeReferee.NotParticipatedRound = 0x006F0005
	ResultCodes.MatchmakeReferee.StatsNotFound = 0x006F0006
	ResultCodes.MatchmakeReferee.RoundNotFound = 0x006F0007
	ResultCodes.MatchmakeReferee.RoundArbitrated = 0x006F0008
	ResultCodes.MatchmakeReferee.RoundNotArbitrated = 0x006F0009

	ResultCodes.Subscriber.Unknown = 0x00700001
	ResultCodes.Subscriber.InvalidArgument = 0x00700002
	ResultCodes.Subscriber.OverLimit = 0x00700003
	ResultCodes.Subscriber.PermissionDenied = 0x00700004

	ResultCodes.Ranking2.Unknown = 0x00710001
	ResultCodes.Ranking2.InvalidArgument = 0x00710002
	ResultCodes.Ranking2.InvalidScore = 0x00710003

	ResultCodes.SmartDeviceVoiceChat.Unknown = 0x00720001
	ResultCodes.SmartDeviceVoiceChat.InvalidArgument = 0x00720002
	ResultCodes.SmartDeviceVoiceChat.InvalidResponse = 0x00720003
	ResultCodes.SmartDeviceVoiceChat.InvalidAccessToken = 0x00720004
	ResultCodes.SmartDeviceVoiceChat.Unauthorized = 0x00720005
	ResultCodes.SmartDeviceVoiceChat.AccessError = 0x00720006
	ResultCodes.SmartDeviceVoiceChat.UserNotFound = 0x00720007
	ResultCodes.SmartDeviceVoiceChat.RoomNotFound = 0x00720008
	ResultCodes.SmartDeviceVoiceChat.RoomNotActivated = 0x00720009
	ResultCodes.SmartDeviceVoiceChat.ApplicationNotSupported = 0x0072000A
	ResultCodes.SmartDeviceVoiceChat.InternalServerError = 0x0072000B
	ResultCodes.SmartDeviceVoiceChat.ServiceUnavailable = 0x0072000C
	ResultCodes.SmartDeviceVoiceChat.UnexpectedError = 0x0072000D
	ResultCodes.SmartDeviceVoiceChat.UnderMaintenance = 0x0072000E
	ResultCodes.SmartDeviceVoiceChat.ServiceNoLongerAvailable = 0x0072000F
	ResultCodes.SmartDeviceVoiceChat.AccountTemporarilyDisabled = 0x00720010
	ResultCodes.SmartDeviceVoiceChat.PermissionDenied = 0x00720011
	ResultCodes.SmartDeviceVoiceChat.NetworkServiceLicenseRequired = 0x00720012
	ResultCodes.SmartDeviceVoiceChat.AccountLibraryError = 0x00720013
	ResultCodes.SmartDeviceVoiceChat.GameModeNotFound = 0x00720014

	ResultCodes.Screening.Unknown = 0x00730001
	ResultCodes.Screening.InvalidArgument = 0x00730002
	ResultCodes.Screening.NotFound = 0x00730003

	ResultCodes.Custom.Unknown = 0x00740001

	ResultCodes.Ess.Unknown = 0x00750001
	ResultCodes.Ess.GameSessionError = 0x00750002
	ResultCodes.Ess.GameSessionMaintenance = 0x00750003

	valueOfResultCodes := reflect.ValueOf(ResultCodes)
	typeOfResultCodes := valueOfResultCodes.Type()

	for i := 0; i < valueOfResultCodes.NumField(); i++ {
		category := typeOfResultCodes.Field(i).Name

		valueOfCategory := reflect.ValueOf(valueOfResultCodes.Field(i).Interface())
		typeOfCategory := valueOfCategory.Type()

		for j := 0; j < valueOfCategory.NumField(); j++ {
			name := typeOfCategory.Field(j).Name
			resultCode := valueOfCategory.Field(j).Interface().(uint32)

			ResultNames[resultCode] = category + "::" + name
		}
	}
}

// ResultCodeToName returns an error code string for the provided error code
func ResultCodeToName(resultCode uint32) string {
	name := ResultNames[resultCode]

	if name == "" {
		return "Invalid Result Code: " + strconv.Itoa(int(resultCode))
	}

	return name
}
