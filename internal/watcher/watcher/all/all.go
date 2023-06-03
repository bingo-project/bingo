package all

import (
	"bingo/internal/watcher/watcher"
	"bingo/internal/watcher/watcher/user"
)

func init() {
	watcher.Register("user", &user.UserWatcher{})
}
