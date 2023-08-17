package handlers

import "github.com/StanzaSystems/sdk-go/hub"

const (
	GuardSuccess = iota
	GuardFailure
	GuardUnknown
)

type Guard struct {
	Success     int
	Failure     int
	Unknown     int
	finalStatus int
	err         error

	quotaToken   string
	quotaStatus  int
	quotaMessage string
	quotaReason  string
}

func (g *Guard) Allowed() bool {
	return g.quotaStatus != hub.CheckQuotaBlocked
}

func (g *Guard) Blocked() bool {
	return g.quotaStatus == hub.CheckQuotaBlocked
}

func (g *Guard) BlockMessage() string {
	return g.quotaMessage
}

func (g *Guard) BlockReason() string {
	return g.quotaReason
}

func (g *Guard) Token() string {
	return g.quotaToken
}

func (g *Guard) Error() error {
	return g.err
}

func (g *Guard) End(status int) {
	g.finalStatus = status
}
