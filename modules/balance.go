package modules

type Balance struct {
	Users      map[string]int64
	Validators map[string]int64
	Transfers  []Transfer
	Stakes     []Stake
}

func NewBalance(balance *Balance) *Balance {
	return &Balance{
		Users:      nil,
		Validators: nil,
	}
}

func (balance *Balance) hash() []byte {
	return nil
}

func (balance *Balance) AddTransfer(transfer *Transfer) {

}

func (balance *Balance) AddStake(state *State) {

}

type Transfer struct {
	sender    []byte
	receiver  []byte
	amount    int64
	time      int64
	signature []byte
}

type Stake struct {
	user      []byte
	validator []byte
	amount    int64
	time      int64
	signature []byte
}
