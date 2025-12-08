package implementation

import "errors"

var ClosedError error = errors.New("reader is closed")
var NoItemsFoundError error = errors.New("no items found")
var AlreadyStartedError error = errors.New("already started")
