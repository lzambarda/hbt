autoload -Uz add-zsh-hook

HBT_PORT=1234
HBT_BIN_PATH=~/Repositories/hbt/bin/hbtsrv
HBT_CACHE_PATH=~/dotfiles/hbt

function hbt_start() {
	pid=$(pgrep hbtsrv)
	if [ -z $pid ]; then
			echo "starting hbt, you can stop it with: hbt_stop"
			#nohup $HBT_BIN_PATH $HBT_PORT $HBT_CACHE_PATH >/dev/null 2>&1 &
			$HBT_BIN_PATH $HBT_PORT $HBT_CACHE_PATH --verbose
	fi
}
function hbt_stop() {
	pid=$(pgrep hbtsrv)
	if [ -z $pid ]; then
		echo "already stopped"
	else
		echo "stopping"
		kill -TERM $pid
	fi
}

if [ -z $(pgrep hbtsrv) ]; then
	hbt_start
fi

function _hbt_track () {
	echo "you run $1 at $(pwd)"
	echo -n "$$\ntrack\n$(pwd)\n$1" | nc localhost $HBT_PORT
}

add-zsh-hook preexec _hbt_track

# list dir with TAB, when there are only spaces/no text before cursor,
# or complete words, that are before cursor only (like in tcsh)
function _hbt_search () {
	if [[ -z ${LBUFFER// } ]]; then
		suggestion=$(echo -n "$$\nhint\n$(pwd)\n$1" | nc localhost $HBT_PORT)
		POSTDISPLAY="${suggestion#$BUFFER}"
		_zsh_autosuggest_highlight_apply
	else
		 zle expand-or-complete-prefix;
	fi
}

function _hbt_clear() {
	if [[ -z ${LBUFFER// } ]]; then
		unset POSTDISPLAY
	else
		zle backward-delete-char
	fi
}

function _hbt_exit() {
	echo -n "$$\nexit\n\n" | nc localhost $HBT_PORT
}

add-zsh-hook zshexit _hbt_exit

# https://unix.stackexchange.com/questions/289883/binding-key-shortcuts-to-shell-functions-in-zsh
zle -N _hbt_search
zle -N _hbt_clear

bindkey '^I' _hbt_search
bindkey '^?' _hbt_clear
