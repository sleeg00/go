package account

import (
	"errors"
)

// Account struct
type Account struct { //banckAccount면 접근 불가.. Private
	owner   string //owner여도 접근 불가 소문자는 안돼!
	balance int
}

var err_nomoney = errors.New("can`t withdraw") // error이름은 err~추천

// NewAccount creates Account
func NewAccount(owner string) *Account { //account의 복사본이 된다
	account := Account{owner: owner, balance: 0}
	return &account //account 즉 복사본의 주소를 return 합니다, 새로운 객체
}

// Deposit x amount on your account
func (a *Account) Deposit(amount int) { //a는 인자값 타입은 Account *Account는 호출한
	//Account을 사용해라
	a.balance += amount
}

// Balance of your account
func (a Account) Balance() int {
	return a.balance
}

// Withdraw x amount from your account
func (a *Account) Withdraw(amount int) error {
	if a.balance < amount {
		return err_nomoney
		//return errors.New("Cant`t withdraw you are poor")
	}
	a.balance -= amount
	return nil // null or none을 뜻함
}
