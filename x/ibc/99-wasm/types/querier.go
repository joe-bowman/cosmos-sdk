package types

const (
	QueryShowCandidate    = "show-candidate"
	QueryGetContract      = "contract-info"
	QueryGetContractState = "contract-state"
	QueryGetCode          = "code"
	QueryListCandidate    = "list-candidate"
)

type QueryClientParams struct {
	ClientId int `json:"client_id" yaml:"client_id"`
}

// QueryAllClientsParams defines the parameters necessary for querying for all
// light client states.
type QueryAllCandidatesParams struct {
	Page  int `json:"page" yaml:"page"`
	Limit int `json:"limit" yaml:"limit"`
}

// NewQueryClientParams creates a new QueryClientParams instance.
func NewQueryClientParams(clientId int) QueryClientParams {
	return QueryClientParams{ClientId: clientId}
}

// NewQueryAllClientsParams creates a new QueryAllClientsParams instance.
func NewQueryAllCandidatesParams(page, limit int) QueryAllCandidatesParams {
	return QueryAllCandidatesParams{
		Page:  page,
		Limit: limit,
	}
}

// StateResponse defines the client response for a client state query.
// It includes the commitment proof and the height of the proof.
type QueryCandidateResponse struct {
}
