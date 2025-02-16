package storage

import "errors"

var ErrUserNotFound = errors.New("user not found")
var ErrItemNotFound = errors.New("item not found")
var ErrNotEnoughBalance = errors.New("not enough tokens")
var ErrSendingToYourself = errors.New("can't send tokens to yourself")
