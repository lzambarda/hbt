autoload -Uz add-zsh-hook

HBT_BIN_PATH=~/Repositories/hbt/bin/hbtsrv

function hbt_start() {
	pid=$(pgrep hbtsrv)
	if [ -z $pid ]; then
			echo "starting hbt, you can stop it with: hbt_stop"
			if [ "$1" = "--debug" ]; then
				$HBT_BIN_PATH --cache $HBT_CACHE_PATH --debug true
			else
				nohup $HBT_BIN_PATH --cache $HBT_CACHE_PATH --debug false >/dev/null 2>&1 &
			fi
	fi
}
function hbt_stop() {
	pid=$(pgrep hbtsrv)
	if [ ! -z $pid ]; then
		kill -TERM $pid
	fi
}

# Start hbtsrv if it is not already started
if [ -z $(pgrep hbtsrv) ]; then
	hbt_start
fi

function _hbt_exit() {
	$HBT_BIN_PATH exit $$
	#echo -n "$$\nexit\n\n" | nc localhost $HBT_PORT
}
add-zsh-hook zshexit _hbt_exit

function _hbt_track () {
	$HBT_BIN_PATH track $$ $(pwd) $1
	#echo -n "$$\ntrack\n$(pwd)\n$1" | nc localhost $HBT_PORT
}
add-zsh-hook preexec _hbt_track

# list dir with TAB, when there are only spaces/no text before cursor,
# or complete words, that are before cursor only (like in tcsh)
function _hbt_search () {
	if [[ -z ${LBUFFER// } ]]; then
		suggestion=$($HBT_BIN_PATH hint $$ $(pwd) $1)
		#suggestion=$(echo -n "$$\nhint\n$(pwd)\n$1" | nc localhost $HBT_PORT)
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

# https://unix.stackexchange.com/questions/289883/binding-key-shortcuts-to-shell-functions-in-zsh
zle -N _hbt_search
zle -N _hbt_clear

bindkey '^I' _hbt_search
bindkey '^?' _hbt_clear
