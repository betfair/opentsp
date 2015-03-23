// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package nitro

// response represents api response.
type response interface {
	errorCode() int
	message() string
}

type responseSessionID struct {
	ErrorCode int
	Message   string

	SessionID string
}

func (r *responseSessionID) errorCode() int {
	return r.ErrorCode
}

func (r *responseSessionID) message() string {
	return r.Message
}

type ResponseStat struct {
	ErrorCode int
	Message   string

	LBVServer []LBVServer

	Service []Service

	ResponderPolicy []struct {
		Name          *string
		HitsRate      *float64 `json:"pipolicyhitsrate"`
		UndefHitsRate *float64 `json:"pipolicyundefhitsrate"`
	}

	ProtocolHTTP *struct {
		HTTPErrIncompleteRequests  *uint64 `json:",string"`
		HTTPErrIncompleteResponses *uint64 `json:",string"`
		HTTPErrServerBusy          *uint64 `json:",string"`
		HTTPTotGets                *uint64 `json:",string"`
		HTTPTotOthers              *uint64 `json:",string"`
		HTTPTotPosts               *uint64 `json:",string"`
		HTTPTotResponses           *uint64 `json:",string"`
		HTTPTotRxRequestBytes      *uint64 `json:",string"`
		HTTPTotTxResponseBytes     *uint64 `json:",string"`
	}

	ProtocolTCP *struct {
		ActiveServerConn                         *uint64  `json:"tcpactiveserverconn,string"`
		ClientConnOpenedRate                     *float64 `json:"tcpclientconnopenedrate"`
		CltFinRate                               *float64 `json:"tcpcltfinrate"`
		CurClientConn                            *uint64  `json:"tcpcurclientconn,string"`
		CurClientConnClosing                     *uint64  `json:"tcpcurclientconnclosing,string"`
		CurClientConnEstablished                 *uint64  `json:"tcpcurclientconnestablished,string"`
		CurClientConnOpening                     *uint64  `json:"tcpcurclientconnopening,string"`
		CurServerConn                            *uint64  `json:"tcpcurserverconn,string"`
		CurServerConnClosing                     *uint64  `json:"tcpcurserverconnclosing,string"`
		CurServerConnEstablished                 *uint64  `json:"tcpcurserverconnestablished,string"`
		CurServerConnOpening                     *uint64  `json:"tcpcurserverconnopening,string"`
		FinWaitClosedRate                        *float64 `json:"tcpfinwaitclosedrate"`
		PcbZombieCallRate                        *float64
		RxBytesRate                              *float64 `json:"tcprxbytesrate"`
		RxPktsRate                               *float64 `json:"tcprxpktsrate"`
		ServerConnOpenedRate                     *float64 `json:"tcpserverconnopenedrate"`
		SpareConn                                *uint64  `json:"tcpspareconn,string"`
		SurgeQueueLen                            *uint64  `json:"tcpsurgequeuelen,string"`
		SvrFinRate                               *float64 `json:"tcpsvrfinrate"`
		SynFlushRate                             *float64 `json:"tcpsynflushrate"`
		SynHeldRate                              *float64 `json:"tcpsynheldrate"`
		SynProbeRate                             *float64 `json:"tcpsynproberate"`
		SynRate                                  *float64 `json:"tcpsynrate"`
		TotServerConnOpened                      *uint64  `json:"tcptotserverconnopened,string"`
		TxBytesRate                              *float64 `json:"tcptxbytesrate"`
		TxPktsRate                               *float64 `json:"tcptxpktsrate"`
		WaitToSynRate                            *float64 `json:"tcpwaittosynrate"`
		WaitToDataRate                           *float64 `json:"tcpwaittodatarate"`
		ZombieActiveHalfCloseCltConnFlushedRate  *float64 `json:"tcpzombieactivehalfclosecltconnflushedrate"`
		ZombieActivehalfCloseSvrConnFlushedRate  *float64 `json:"tcpzombieactivehalfclosesvrconnflushedrate"`
		ZombieCltConnFlushedRate                 *float64 `json:"tcpzombiecltconnflushedrate"`
		ZombieHalfOpenCltConnFlushedRate         *float64 `json:"tcpzombiehalfopencltconnflushedrate"`
		ZombieHalfOpenSvrConnFlushedRate         *float64 `json:"tcpzombiehalfopensvrconnflushedrate"`
		ZombiePassiveHalfClosecltConnFlushedRate *float64 `json:"tcpzombiepassivehalfclosecltconnflushedrate"`
		ZombiePassivehalfCloseSrvConnFlushedRate *float64 `json:"tcpzombiepassivehalfclosesrvconnflushedrate"`
		ZombieSvrConnFlushedRate                 *float64 `json:"tcpzombiesvrconnflushedrate"`

		ErrAnyPortFailRate            *float64 `json:"tcperranyportfailrate"`
		ErrBadChecksumRate            *float64 `json:"tcperrbadchecksumrate"`
		ErrBadstateConnRate           *float64 `json:"tcperrbadstateconnrate"`
		ErrCipAllocRate               *float64 `json:"tcperrcipallocrate"`
		ErrCltHoleRate                *float64 `json:"tcperrcltholerate"`
		ErrCltOutOfOrderRate          *float64 `json:"tcperrcltoutoforderrate"`
		ErrCltRetrasmitRate           *float64 `json:"tcperrcltretrasmitrate"`
		ErrCookiePktMssRejectRate     *float64 `json:"tcperrcookiepktmssrejectrate"`
		ErrCookiePktSeqDropRate       *float64 `json:"tcperrcookiepktseqdroprate"`
		ErrCookiePktSeqRejectRate     *float64 `json:"tcperrcookiepktseqrejectrate"`
		ErrCookiePktSigRejectRate     *float64 `json:"tcperrcookiepktsigrejectrate"`
		ErrDataAfterFinRate           *float64 `json:"tcperrdataafterfinrate"`
		ErrFastRetransmissionsRate    *float64 `json:"tcperrfastretransmissionsrate"`
		ErrFifthRetransmissionsRate   *float64 `json:"tcperrfifthretransmissionsrate"`
		ErrFinGiveUpRate              *float64 `json:"tcperrfingiveuprate"`
		ErrFinRetryRate               *float64 `json:"tcperrfinretryrate"`
		ErrFirstRetransmissionsRate   *float64 `json:"tcperrfirstretransmissionsrate"`
		ErrFourthRetransmissionsRate  *float64 `json:"tcperrforthretransmissionsrate"`
		ErrFullRetrasmitRate          *float64 `json:"tcperrfullretrasmitrate"`
		ErrIpPortFailRate             *float64 `json:"tcperripportfailrate"`
		ErrOutOfWindowPktsRate        *float64 `json:"tcperroutofwindowpktsrate"`
		ErrPartialRetrasmitRate       *float64 `json:"tcperrpartialretrasmitrate"`
		ErrRetransmitGiveUpRate       *float64 `json:"tcperrretransmitgiveuprate"`
		ErrRetransmitRate             *float64 `json:"tcperrretransmitrate"`
		ErrRstInTimeWaitRate          *float64 `json:"tcperrrstintimewaitrate"`
		ErrRstNonEstRate              *float64 `json:"tcperrrstnonestrate"`
		ErrRstOutOfWindowRate         *float64 `json:"tcperrrstoutofwindowrate"`
		ErrRstRate                    *float64 `json:"tcperrrstrate"`
		ErrRstThresholdRate           *float64 `json:"tcperrrstthresholdrate"`
		ErrSecondRetransmissionsRate  *float64 `json:"tcperrsecondretransmissionsrate"`
		ErrSentRstRate                *float64 `json:"tcperrsentrstrate"`
		ErrSeventhRetransmissionsRate *float64 `json:"tcperrseventhretransmissionsrate"`
		ErrSixthRetransmissionsRate   *float64 `json:"tcperrsixthretransmissionsrate"`
		ErrStrayPktRate               *float64 `json:"tcperrstraypktrate"`
		ErrSvrHoleRate                *float64 `json:"tcperrsvrholerate"`
		ErrSvrOutOfOrderRate          *float64 `json:"tcperrsvroutoforderrate"`
		ErrSvrRetrasmitRate           *float64 `json:"tcperrsvrretrasmitrate"`
		ErrSynDroppedCongestionRate   *float64 `json:"tcperrsyndroppedcongestionrate"`
		ErrSynGiveupRate              *float64 `json:"tcperrsyngiveuprate"`
		ErrSynInSynrcvdRate           *float64 `json:"tcperrsyninsynrcvdrate"`
		ErrSynRetryRate               *float64 `json:"tcperrsynretryrate"`
		ErrSynSentBadackRate          *float64 `json:"tcperrsynsentbadackrate"`
		ErrSyninestRate               *float64 `json:"tcperrsyninestrate"`
		ErrThirdRetransmissionsRate   *float64 `json:"tcperrthirdretransmissionsrate"`
	}

	ProtocolUDP *struct {
		RxPktsRate              *float64 `json:"udprxpktsrate"`
		RxBytesRate             *float64 `json:"udprxbytesrate"`
		TxPktsRate              *float64 `json:"udptxpktsrate"`
		TxBytesRate             *float64 `json:"udptxbytesrate"`
		CurRateThreshold        *float64 `json:"udpcurratethreshold,string"`
		TotUnknownSvcPkts       *uint64  `json:"udptotunknownsvcpkts,string"`
		BadChecksum             *uint64  `json:"udpbadchecksum,string"`
		CurRateThresholdExceeds *float64 `json:"udpcurratethresholdexceeds,string"`
	}

	SSL *struct {
		// Global gauges.
		NewSessionsRate            float64 `json:"sslnewsessionsrate"`
		SessionHitsRate            float64 `json:"sslsessionhitsrate"`
		SessionMissRate            float64 `json:"sslsessionmissrate"`
		SSLv3RenegSessionsRate     float64 `json:"sslsslv3renegsessionsrate"`
		TLSv1RenegSessionsRate     float64 `json:"ssltlsv1renegsessionsrate"`
		OffloadRSAKeyExchangesRate float64 `json:"ssloffloadrsakeyexchangesrate"`
		OffloadDHKeyExchangesRate  float64 `json:"ssloffloaddhkeyexchangesrate"`
		OffloadSignRSARate         float64 `json:"ssloffloadsignrsarate"`
		OffloadBulkRC4Rate         float64 `json:"ssloffloadbulkrc4rate"`
		OffloadBulkDESRate         float64 `json:"ssloffloadbulkdesrate"`
		OffloadBulkAESRate         float64 `json:"ssloffloadbulkaesrate"`

		// Client-side gauges.
		ClientSSLv2SessionsRate              float64 `json:"sslsslv2sessionsrate"`
		ClientSSLv3SessionsRate              float64 `json:"sslsslv3sessionsrate"`
		ClientTLSv1SessionsRate              float64 `json:"ssltlsv1sessionsrate"`
		ClientSSLv2TransactionsRate          float64 `json:"sslsslv2transactionsrate"`
		ClientSSLv3TransactionsRate          float64 `json:"sslsslv3transactionsrate"`
		ClientTLSv1TransactionsRate          float64 `json:"ssltlsv1transactionsrate"`
		ClientHWEncFERate                    float64 `json:"sslhwencferate"`
		ClientSWEncFERate                    float64 `json:"sslswencferate"`
		ClientHWDecFERate                    float64 `json:"sslhwdecferate"`
		ClientSWDecFERate                    float64 `json:"sslswdecferate"`
		ClientRSA512KeyExchangesRate         float64 `json:"sslrsa512keyexchangesrate"`
		ClientRSA1024KeyExchangesRate        float64 `json:"sslrsa1024keyexchangesrate"`
		ClientRSA2048KeyExchangesRate        float64 `json:"sslrsa2048keyexchangesrate"`
		ClientRSA4096KeyExchangesRate        float64 `json:"sslrsa4096keyexchangesrate"`
		ClientDH512KeyExchangesRate          float64 `json:"ssldh512keyexchangesrate"`
		ClientDH1024KeyExchangesRate         float64 `json:"ssldh1024keyexchangesrate"`
		ClientDH2048KeyExchangesRate         float64 `json:"ssldh2048keyexchangesrate"`
		Client40BitRC4CiphersRate            float64 `json:"ssl40bitrc4ciphersrate"`
		Client56BitRC4CiphersRate            float64 `json:"ssl56bitrc4ciphersrate"`
		Client64BitRC4CiphersRate            float64 `json:"ssl64bitrc4ciphersrate"`
		Client128BitRC4CiphersRate           float64 `json:"ssl128bitrc4ciphersrate"`
		Client40BitDESCiphersRate            float64 `json:"ssl40bitdesciphersrate"`
		Client56BitDESCiphersRate            float64 `json:"ssl56bitdesciphersrate"`
		Client168Bit3DESCiphersRate          float64 `json:"ssl168bit3desciphersrate"`
		ClientCipherAES128Rate               float64 `json:"sslcipheraes128rate"`
		ClientCipherAES256Rate               float64 `json:"sslcipheraes256rate"`
		Client40BitRC2CiphersRate            float64 `json:"ssl40bitrc2ciphersrate"`
		Client56BitRC2CiphersRate            float64 `json:"ssl56bitrc2ciphersrate"`
		Client128BitRC2CiphersRate           float64 `json:"ssl128bitrc2ciphersrate"`
		ClientNULLCiphersRate                float64 `json:"sslnullciphersrate"`
		Client128BitIDEACiphersRate          float64 `json:"ssl128bitideaciphersrate"`
		ClientMD5MacRate                     float64 `json:"sslmd5macrate"`
		ClientSHAMacRate                     float64 `json:"sslshamacrate"`
		ClientSSLv2HandshakesRate            float64 `json:"sslsslv2handshakesrate"`
		ClientSSLv3HandshakesRate            float64 `json:"sslsslv3handshakesrate"`
		ClientTLSv1HandshakesRate            float64 `json:"ssltlsv1handshakesrate"`
		ClientSSLv2ClientAuthenticationsRate float64 `json:"sslsslv2clientauthenticationsrate"`
		ClientSSLv3ClientAuthenticationsRate float64 `json:"sslsslv3clientauthenticationsrate"`
		ClientTLSv1ClientAuthenticationsRate float64 `json:"ssltlsv1clientauthenticationsrate"`
		ClientRSAAuthorizationsRate          float64 `json:"sslrsaauthorizationsrate"`
		ClientDHAuthorizationsRate           float64 `json:"ssldhauthorizationsrate"`
		ClientDSSAuthorizationsRate          float64 `json:"ssldssauthorizationsrate"`
		ClientNULLAuthorizationsRate         float64 `json:"sslnullauthorizationsrate"`

		// Server-side gauges.
		ServerSSLv3SessionsRate                  float64 `json:"sslbesslv3sessionsrate"`
		ServerTLSv1SessionsRate                  float64 `json:"sslbetlsv1sessionsrate"`
		ServerSessionMultiplexAttemptSuccessRate float64 `json:"sslbesessionmultiplexattemptsuccessrate"`
		ServerSessionMultiplexAttemptFailsRate   float64 `json:"sslbesessionmultiplexattemptfailsrate"`
		ServerRSA512KeyExchangesRate             float64 `json:"sslbersa512keyexchangesrate"`
		ServerRSA1024KeyExchangesRate            float64 `json:"sslbersa1024keyexchangesrate"`
		ServerRSA2048KeyExchangesRate            float64 `json:"sslbersa2048keyexchangesrate"`
		ServerRSA4096KeyExchangesRate            float64 `json:"sslbersa4096keyexchangesrate"`
		ServerDH512KeyExchangesRate              float64 `json:"sslbedh512keyexchangesrate"`
		ServerDH1024KeyExchangesRate             float64 `json:"sslbedh1024keyexchangesrate"`
		ServerDH2048KeyExchangesRate             float64 `json:"sslbedh2048keyexchangesrate"`
		Server40BitRC4CiphersRate                float64 `json:"sslbe40bitrc4ciphersrate"`
		Server56BitRC4CiphersRate                float64 `json:"sslbe56bitrc4ciphersrate"`
		Server64BitRC4CiphersRate                float64 `json:"sslbe64bitrc4ciphersrate"`
		Server128BitRC4CiphersRate               float64 `json:"sslbe128bitrc4ciphersrate"`
		Server40BitDESCiphersRate                float64 `json:"sslbe40bitdesciphersrate"`
		Server56BitDESCiphersRate                float64 `json:"sslbe56bitdesciphersrate"`
		Server168Bit3DESCiphersRate              float64 `json:"sslbe168bit3desciphersrate"`
		ServerCipherAES128Rate                   float64 `json:"sslbecipheraes128rate"`
		ServerCipherAES256Rate                   float64 `json:"sslbecipheraes256rate"`
		Server40BitRC2CiphersRate                float64 `json:"sslbe40bitrc2ciphersrate"`
		Server56BitRC2CiphersRate                float64 `json:"sslbe56bitrc2ciphersrate"`
		Server128BitRC2CiphersRate               float64 `json:"sslbe128bitrc2ciphersrate"`
		ServerNULLCiphersRate                    float64 `json:"sslbenullciphersrate"`
		Server128BitIDEACiphersRate              float64 `json:"sslbe128bitideaciphersrate"`
		ServerMD5MacRate                         float64 `json:"sslbemd5macrate"`
		ServerSHAMacRate                         float64 `json:"sslbeshamacrate"`
		ServerSSLv3HandshakesRate                float64 `json:"sslbesslv3handshakesrate"`
		ServerTLSv1HandshakesRate                float64 `json:"sslbetlsv1handshakesrate"`
		ServerSSLv3ClientAuthenticationsRate     float64 `json:"sslbesslv3clientauthenticationsrate"`
		ServerTLSv1ClientAuthenticationsRate     float64 `json:"sslbetlsv1clientauthenticationsrate"`
		ServerRSAAuthorizationsRate              float64 `json:"sslbersaauthorizationsrate"`
		ServerDHAuthorizationsRate               float64 `json:"sslbedhauthorizationsrate"`
		ServerDSSAuthorizationsRate              float64 `json:"sslbedssauthorizationsrate"`
		ServerNULLAuthorizationsRate             float64 `json:"sslbenullauthorizationsrate"`
		ServerMultiplexedSessionsRate            float64 `json:"sslbemultiplexedsessionsrate"`
		ServerHWEncBERate                        float64 `json:"sslhwencberate"`
		ServerSWEncBERate                        float64 `json:"sslswencberate"`
		ServerHWDecBERate                        float64 `json:"sslhwdecberate"`
		ServerSWDecBERate                        float64 `json:"sslswdecberate"`
	}

	NS *struct {
		HTTPToTRequests  uint64 `json:",string"`
		MgmtCPUUsagePcnt *float64
		PktCPUUsagePcnt  *float64
		ResMemUsage      *uint64 `json:",string"`
		MemUsagePcnt     *float64
	}

	Interface []struct {
		ID                  string
		TotRxBytes          *uint64 `json:",string"`
		TotTxBytes          *uint64 `json:",string"`
		TotRxPkts           *uint64 `json:",string"`
		TotTxPkts           *uint64 `json:",string"`
		ErrIfInDiscards     *uint64 `json:",string"`
		NicErrIfOutDiscards *uint64 `json:",string"`
		ErrPktRx            *uint64 `json:",string"`
		ErrPktTx            *uint64 `json:",string"`
		// missing: Speed
		// missing: PacketsQueued
		// missing: InUnknownProtos
	}
}

func (r *ResponseStat) errorCode() int {
	return r.ErrorCode
}

func (r *ResponseStat) message() string {
	return r.Message
}

type LBVServer struct {
	Name   string
	Type   string
	State  string
	Health *uint64 `json:"vslbhealth,string"`

	TotalPktsSent *uint64 `json:",string"`

	RequestsRate       *float64
	ResponsesRate      *float64
	RequestBytesRate   *float64
	ResponseBytesRate  *float64
	CurClntConnections *uint64 `json:",string"`
	CurSrvrConnections *uint64 `json:",string"`

	// NB: present in documentation but absent in Nitro API.
	// SvrEstablishedConn *uint64 `json:",string"`

	SpilloverThreshold *uint64 `json:"sothreshold,string"`
	Spillovers         *uint64 `json:"totspillovers,string"`
}

type Service struct {
	Name        string
	ServiceType string
	State       string

	MaxClients         *uint64  `json:",string"`
	ActiveTransactions *uint64  `json:",string"`
	AvgSvrTTFB         *float64 `json:",string"`
	CurClntConnections *uint64  `json:",string"`
	CurReusePool       *uint64  `json:",string"`
	CurSrvrConnections *uint64  `json:",string"`
	RequestsRate       *float64
	ResponsesRate      *float64
	SurgeCount         *uint64 `json:",string"`
	SvrEstablishedConn *uint64 `json:",string"`
	RequestBytesRate   *float64
	ResponseBytesRate  *float64
}

type ResponseConfig struct {
	ErrorCode int
	Message   string

	LBVServerServiceBinding []struct {
		ServiceName string
	} `json:"lbvserver_service_binding"`

	FilterPolicy []struct {
		Name string
		Hits *uint64 `json:",string"`
	}
}

func (r *ResponseConfig) errorCode() int {
	return r.ErrorCode
}

func (r *ResponseConfig) message() string {
	return r.Message
}
