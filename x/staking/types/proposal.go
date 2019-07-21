package types


import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/dao"
	"io/ioutil"

	//"github.com/cosmos/cosmos-sdk/x/staking"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	daotypes "github.com/cosmos/cosmos-sdk/x/dao/types"
)

const (
	// ProposalTypeRebalancing defines the type for a RebalancingProposal
	ProposalTypeRebalancing = "Rebalancing"
)

// Assert RebalancingProposal implements govtypes.Content at compile-time
var _ daotypes.Content = RebalancingProposal{}

func init() {
	dao.RegisterProposalType(ProposalTypeRebalancing)
	daotypes.RegisterProposalTypeCodec(RebalancingProposal{}, "cosmos-sdk/RebalancingProposal")
}

// RebalancingProposal defines a proposal which contains multiple parameter
// changes.
type RebalancingProposal struct {
	Title       string        `json:"title" yaml:"title"`
	Description string        `json:"description" yaml:"description"`
	Rebalancing Rebalancing `json:"rebalancing yaml:rebalancing` // list of redelegation pairs
}

func NewRebalancingProposal(title, description string, Rebalancing Rebalancing) RebalancingProposal {
	return RebalancingProposal{title, description, Rebalancing}
}

// GetTitle returns the title of a parameter change proposal.
func (rp RebalancingProposal) GetTitle() string { return rp.Title }

// GetDescription returns the description of a parameter change proposal.
func (rp RebalancingProposal) GetDescription() string { return rp.Description }

// GetDescription returns the routing key of a parameter change proposal.
func (rp RebalancingProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a parameter change proposal.
func (rp RebalancingProposal) ProposalType() string { return ProposalTypeRebalancing }

// ValidateBasic validates the parameter change proposal
func (rp RebalancingProposal) ValidateBasic() sdk.Error {
	err := daotypes.ValidateAbstract(DefaultCodespace, rp)
	if err != nil {
		return err
	}

	return ValidateChanges(rp.Rebalancing)
}

// String implements the Stringer interface.
func (pcp RebalancingProposal) String() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf(`Rebalancing DAO Proposal:
  Title:       %s
  Description: %s
  Rebalancing:
`, pcp.Title, pcp.Description))

	for _, rp := range pcp.Rebalancing {
		b.WriteString(fmt.Sprintf(`    Rebalancing:
      ValidatorSrcAddress: %s
      ValidatorDstAddress:      %s
      Amount:   %s
`, rp.ValidatorSrcAddress, rp.ValidatorDstAddress, rp.Amount.String()))
	}

	return b.String()
}

// TODO: need DelegatorAddress?
type RedelegationPair struct {
	//DelegatorAddress    sdk.AccAddress `json:"delegator_address"`
	ValidatorSrcAddress sdk.ValAddress `json:"validator_src_address"`
	ValidatorDstAddress sdk.ValAddress `json:"validator_dst_address"`
	Amount              sdk.Coin       `json:"amount"`
}

type Rebalancing []RedelegationPair // list of redelegation pairs

// String implements the Stringer interface.
func (rp RedelegationPair) String() string {
	return fmt.Sprintf(`Param Change:
  ValidatorSrcAddress: %s
  ValidatorDstAddress: %s
  Amount:   %s
`, rp.ValidatorSrcAddress, rp.ValidatorDstAddress, rp.Amount.String())
}

// ValidateChange performs basic validation checks over a set of ParamChange. It
// returns an error if any ParamChange is invalid.
func ValidateChanges(rebalancing Rebalancing) sdk.Error {
	if len(rebalancing) == 0 {
		return ErrNoRedelegation(DefaultCodespace)
	}

	for _, r := range rebalancing {
		fmt.Println(r)
		if len(r.ValidatorSrcAddress) == 0 {
			return ErrBadRedelegationAddr(DefaultCodespace)
		}
		if len(r.ValidatorDstAddress) == 0 {
			return ErrBadRedelegationDst(DefaultCodespace)
		}
		if r.Amount.IsZero() {
			return ErrBadSharesAmount(DefaultCodespace)
		}
	}

	return nil
}


// ParseParamChangeProposalJSON reads and parses a ParamChangeProposalJSON from
// file.
func ParseRebalancingProposalJSON(cdc *codec.Codec, proposalFile string) (RebalancingProposal, error) {
	proposal := RebalancingProposal{}

	contents, err := ioutil.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	if err := cdc.UnmarshalJSON(contents, &proposal); err != nil {
		return proposal, err
	}

	return proposal, nil
}

