package types

const (
	WithdrawSourceNoSource           = 0x00
	WithdrawSourceDelegationModified = 0x01
	WithdrawSourceExplicit           = 0x02
	WithdrawSourceBulkDaily          = 0x03
	WithdrawSourceBulkTenK           = 0x04
	WithdrawSourceBulkFirstBlock     = 0x05
	WithdrawSourceZero               = 0x06
)
