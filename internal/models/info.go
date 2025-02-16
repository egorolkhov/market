package models

type Info struct {
	Coins        int     `json:"coins"`
	CoinsHistory History `json:"coinHistory"`
	Inventory    []Item  `json:"inventory"`
}

type Item struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

type History struct {
	Received []Transaction `json:"received"`
	Sent     []Transaction `json:"sent"`
}

type Transaction struct {
	FromUser string `json:"fromUser"`
	ToUser   string `json:"toUser"`
	Amount   int    `json:"amount"`
}
