# App info
export app_name=bingo
export version=$(git describe --tags --match='v*' | sed 's/^v//' || echo '0.0.0')

# Build
export registry_prefix=bingo
export images=(bingo-apiserver bingo-watcher bingo-bot bingoctl)
export architecture=amd64
