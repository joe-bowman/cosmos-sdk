package types

import (
	tmmath "github.com/tendermint/tendermint/libs/math"
	"net/url"
	"regexp"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// Message types for the IBC client
const (
	TypeMsgCreateWasmClient             string = "create_wasmclient"
	TypeMsgUpdateWasmClient             string = "update_wasmclient"
	TypeMsgSubmitWasmClientMisbehaviour string = "submit_wasmclient_misbehaviour"
	TypeMsgStoreClientCode              string = "store_wasmclient"
	TypeMsgWrappedData                  string = "wrapped_data"
	TODO                                string = "to be implemented."

	MaxWasmSize = 500 * 1024

	// MaxLabelSize is the longest label that can be used when Instantiating a contract
	MaxLabelSize = 128

	// BuildTagRegexp is a docker image regexp.
	// We only support max 128 characters, with at least one organization name (subset of all legal names).
	//
	// Details from https://docs.docker.com/engine/reference/commandline/tag/#extended-description :
	//
	// An image name is made up of slash-separated name components (optionally prefixed by a registry hostname).
	// Name components may contain lowercase characters, digits and separators.
	// A separator is defined as a period, one or two underscores, or one or more dashes. A name component may not start or end with a separator.
	//
	// A tag name must be valid ASCII and may contain lowercase and uppercase letters, digits, underscores, periods and dashes.
	// A tag name may not start with a period or a dash and may contain a maximum of 128 characters.
	BuildTagRegexp = "^[a-z0-9][a-z0-9._-]*[a-z0-9](/[a-z0-9][a-z0-9._-]*[a-z0-9])+:[a-zA-Z0-9_][a-zA-Z0-9_.-]*$"

	MaxBuildTagSize = 128
)

var (
	_ clientexported.MsgCreateClient     = MsgCreateWasmClient{}
	_ clientexported.MsgUpdateClient     = MsgUpdateWasmClient{}
	_ evidenceexported.MsgSubmitEvidence = MsgSubmitWasmClientMisbehaviour{}
)

// MsgCreateWasmClient defines a message to create an IBC client
type MsgCreateWasmClient struct {
	ClientID        string          `json:"client_id" yaml:"client_id"`
	Header          Header          `json:"header" yaml:"header"`
	Message         string          `json:"message" yaml:"message"`
	TrustLevel      tmmath.Fraction `json:"trust_level" yaml:"trust_level"`
	TrustingPeriod  time.Duration   `json:"trusting_period" yaml:"trusting_period"`
	UnbondingPeriod time.Duration   `json:"unbonding_period" yaml:"unbonding_period"`
	MaxClockDrift   time.Duration   `json:"max_clock_drift" yaml:"max_clock_drift"`
	Signer          sdk.AccAddress  `json:"address" yaml:"address"`
	WasmId          int             `json:"wasm_id" yaml:"wasm_id"`
}

// NewMsgCreateWasmClient creates a new MsgCreateWasmClient instance
func NewMsgCreateWasmClient(
	id string, header Header,
	trustLevel tmmath.Fraction,
	trustingPeriod, unbondingPeriod, maxClockDrift time.Duration, signer sdk.AccAddress, wasmId int,
) MsgCreateWasmClient {
	return MsgCreateWasmClient{
		ClientID:        id,
		Header:          header,
		Message:         "DEFAULT",
		TrustLevel:      trustLevel,
		TrustingPeriod:  trustingPeriod,
		UnbondingPeriod: unbondingPeriod,
		MaxClockDrift:   maxClockDrift,
		Signer:          signer,
		WasmId:          wasmId,
	}
}

// Route implements sdk.Msg
func (msg MsgCreateWasmClient) Route() string {
	return host.RouterKey
}

// Type implements sdk.Msg
func (msg MsgCreateWasmClient) Type() string {
	return TypeMsgCreateWasmClient
}

func (msg MsgCreateWasmClient) Reset()         {}
func (msg MsgCreateWasmClient) String() string { return TODO }
func (msg MsgCreateWasmClient) ProtoMessage()  {}

// ValidateBasic implements sdk.Msg
func (msg MsgCreateWasmClient) ValidateBasic() error {
	if msg.TrustingPeriod == 0 {
		return sdkerrors.Wrap(ErrInvalidTrustingPeriod, "duration cannot be 0")
	}
	if msg.UnbondingPeriod == 0 {
		return sdkerrors.Wrap(ErrInvalidUnbondingPeriod, "duration cannot be 0")
	}
	if msg.Signer.Empty() {
		return sdkerrors.ErrInvalidAddress
	}
	// ValidateBasic of provided header with self-attested chain-id
	//if err := msg.Header.ValidateBasic(msg.Header.ChainID); err != nil {
	//	return sdkerrors.Wrapf(ErrInvalidHeader, "header failed validatebasic with its own chain-id: %v", err)
	//}

	// check wasmId exists.

	return host.ClientIdentifierValidator(msg.ClientID)
}

// GetSignBytes implements sdk.Msg
func (msg MsgCreateWasmClient) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements sdk.Msg
func (msg MsgCreateWasmClient) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

// GetClientID implements clientexported.MsgCreateWasmClient
func (msg MsgCreateWasmClient) GetClientID() string {
	return msg.ClientID
}

// GetClientType implements clientexported.MsgCreateWasmClient
func (msg MsgCreateWasmClient) GetClientType() string {
	return clientexported.ClientTypeWasm
}

// GetConsensusState implements clientexported.MsgCreateWasmClient
func (msg MsgCreateWasmClient) GetConsensusState() clientexported.ConsensusState {
	// Construct initial consensus state from provided Header
	root := commitmenttypes.NewMerkleRoot(msg.Header.AppHash)
	return ConsensusState{
		Timestamp:    msg.Header.Time,
		Root:         root,
		Height:       uint64(msg.Header.Height),
		ValidatorSet: msg.Header.ValidatorSet,
	}
}

// MsgUpdateWasmClient defines a message to update an IBC client
type MsgUpdateWasmClient struct {
	ClientID string         `json:"client_id" yaml:"client_id"`
	Header   Header         `json:"header" yaml:"header"`
	Signer   sdk.AccAddress `json:"address" yaml:"address"`
}

// NewMsgUpdateWasmClient creates a new MsgUpdateWasmClient instance
func NewMsgUpdateWasmClient(id string, header Header, signer sdk.AccAddress) MsgUpdateWasmClient {
	return MsgUpdateWasmClient{
		ClientID: id,
		Header:   header,
		Signer:   signer,
	}
}

// Route implements sdk.Msg
func (msg MsgUpdateWasmClient) Route() string {
	return host.RouterKey
}

// Type implements sdk.Msg
func (msg MsgUpdateWasmClient) Type() string {
	return TypeMsgUpdateWasmClient
}

// dummy implementation of proto.Message
func (msg MsgUpdateWasmClient) Reset()         {}
func (msg MsgUpdateWasmClient) String() string { return TODO }
func (msg MsgUpdateWasmClient) ProtoMessage()  {}

// ValidateBasic implements sdk.Msg
func (msg MsgUpdateWasmClient) ValidateBasic() error {
	if msg.Signer.Empty() {
		return sdkerrors.ErrInvalidAddress
	}
	return host.ClientIdentifierValidator(msg.ClientID)
}

// GetSignBytes implements sdk.Msg
func (msg MsgUpdateWasmClient) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements sdk.Msg
func (msg MsgUpdateWasmClient) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

// GetClientID implements clientexported.MsgUpdateWasmClient
func (msg MsgUpdateWasmClient) GetClientID() string {
	return msg.ClientID
}

// GetHeader implements clientexported.MsgUpdateWasmClient
func (msg MsgUpdateWasmClient) GetHeader() clientexported.Header {
	return msg.Header
}

// MsgSubmitWasmClientMisbehaviour defines an sdk.Msg type that supports submitting
// Evidence for client misbehaviour.
type MsgSubmitWasmClientMisbehaviour struct {
	Evidence  evidenceexported.Evidence `json:"evidence" yaml:"evidence"`
	Submitter sdk.AccAddress            `json:"submitter" yaml:"submitter"`
}

// NewMsgSubmitWasmClientMisbehaviour creates a new MsgSubmitWasmClientMisbehaviour
// instance.
func NewMsgSubmitWasmClientMisbehaviour(e evidenceexported.Evidence, s sdk.AccAddress) MsgSubmitWasmClientMisbehaviour {
	return MsgSubmitWasmClientMisbehaviour{Evidence: e, Submitter: s}
}

// Route returns the MsgSubmitWasmClientMisbehaviour's route.
func (msg MsgSubmitWasmClientMisbehaviour) Route() string { return host.RouterKey }

// Type returns the MsgSubmitWasmClientMisbehaviour's type.
func (msg MsgSubmitWasmClientMisbehaviour) Type() string { return TypeMsgSubmitWasmClientMisbehaviour }

// dummy implementation of proto.Message
func (msg MsgSubmitWasmClientMisbehaviour) Reset()         {}
func (msg MsgSubmitWasmClientMisbehaviour) String() string { return TODO }
func (msg MsgSubmitWasmClientMisbehaviour) ProtoMessage()  {}

// ValidateBasic performs basic (non-state-dependant) validation on a MsgSubmitWasmClientMisbehaviour.
func (msg MsgSubmitWasmClientMisbehaviour) ValidateBasic() error {
	if msg.Evidence == nil {
		return sdkerrors.Wrap(evidencetypes.ErrInvalidEvidence, "missing evidence")
	}
	if err := msg.Evidence.ValidateBasic(); err != nil {
		return err
	}
	if msg.Submitter.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Submitter.String())
	}

	return nil
}

// GetSignBytes returns the raw bytes a signer is expected to sign when submitting
// a MsgSubmitWasmClientMisbehaviour message.
func (msg MsgSubmitWasmClientMisbehaviour) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(msg))
}

// GetSigners returns the single expected signer for a MsgSubmitWasmClientMisbehaviour.
func (msg MsgSubmitWasmClientMisbehaviour) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Submitter}
}

func (msg MsgSubmitWasmClientMisbehaviour) GetEvidence() evidenceexported.Evidence {
	return msg.Evidence
}

func (msg MsgSubmitWasmClientMisbehaviour) GetSubmitter() sdk.AccAddress {
	return msg.Submitter
}

type MsgStoreClientCode struct {
	Sender sdk.AccAddress `json:"sender" yaml:"sender"`
	// WASMByteCode can be raw or gzip compressed
	WASMByteCode []byte `json:"wasm_byte_code" yaml:"wasm_byte_code"`
	// Source is a valid absolute HTTPS URI to the contract's source code, optional
	Source string `json:"source" yaml:"source"`
	// Builder is a valid docker image name with tag, optional
	Builder string `json:"builder" yaml:"builder"`
}

func (msg MsgStoreClientCode) Route() string {
	return host.RouterKey
}

func (msg MsgStoreClientCode) Type() string {
	return TypeMsgStoreClientCode
}

func (msg MsgStoreClientCode) Reset()         {}
func (msg MsgStoreClientCode) String() string { return TODO }
func (msg MsgStoreClientCode) ProtoMessage()  {}

func (msg MsgStoreClientCode) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(msg.Sender); err != nil {
		return err
	}

	if len(msg.WASMByteCode) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty wasm code")
	}

	if len(msg.WASMByteCode) > MaxWasmSize {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "wasm code too large")
	}

	if msg.Source != "" {
		u, err := url.Parse(msg.Source)
		if err != nil {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "source should be a valid url")
		}
		if !u.IsAbs() {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "source should be an absolute url")
		}
		if u.Scheme != "https" {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "source must use https")
		}
	}

	return validateBuilder(msg.Builder)
}

func (msg MsgStoreClientCode) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgStoreClientCode) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

func validateBuilder(buildTag string) error {
	if len(buildTag) > MaxBuildTagSize {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "builder tag longer than 128 characters")
	}

	if buildTag != "" {
		ok, err := regexp.MatchString(BuildTagRegexp, buildTag)
		if err != nil || !ok {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid tag supplied for builder")
		}
	}

	return nil
}

type MsgWrappedData struct {
	Sender sdk.AccAddress `json:"sender" yaml:"sender"`
	Target Address        `json:"target" yaml:"target"`
	// Data can be gzipped or raw.
	Data []byte `json:"data" yaml:"data"`
}

func (msg MsgWrappedData) Route() string {
	return host.RouterKey
}

func (msg MsgWrappedData) Type() string {
	return TypeMsgWrappedData
}

func (msg MsgWrappedData) Reset()         {}
func (msg MsgWrappedData) String() string { return TODO }
func (msg MsgWrappedData) ProtoMessage()  {}

func (msg MsgWrappedData) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(msg.Sender); err != nil {
		return err
	}

	return nil
}

func (msg MsgWrappedData) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgWrappedData) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
