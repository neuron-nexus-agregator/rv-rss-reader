package implementation

import "errors"

var ErrClosed error = errors.New("reader is closed")
var ErrNoItemsFound error = errors.New("no items found")
var ErrAlreadyStarted error = errors.New("already started")
