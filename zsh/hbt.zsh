# To be able to use zsh hooks
autoload -Uz add-zsh-hook

export PATH="$HOME/Repositories/hbt/bin/darwin:$PATH"
export HBT_CACHE_PATH="$HOME/dotfiles/hbt/"
export HBT_PORT=43111
export HBT_SAVE_INTERVAL="60m"

function hbt_start() {
	pid=$(pgrep hbtsrv)
	if [ -z $pid ]; then
			echo "starting hbt, you can stop it with: hbt_stop"
			if [ "$1" = "--debug" ]; then
				hbtsrv --debug
			else
				nohup hbtsrv >/dev/null 2>&1 &
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

function _hbt_end_session() { echo -n "end\n$$" | nc localhost $HBT_PORT ; }
add-zsh-hook zshexit _hbt_end_session

function _hbt_track () { echo -n "track\n$$\n$(pwd)\n$1" | nc localhost $HBT_PORT ; }
add-zsh-hook preexec _hbt_track

# list dir with TAB, when there are only spaces/no text before cursor,
# or complete words, that are before cursor only (like in tcsh)
function _hbt_search () {
	if [[ -z ${LBUFFER// } ]]; then
		suggestion=$(echo -n "hint\n$$\n$(pwd)\n$1" | nc localhost $HBT_PORT)
		POSTDISPLAY="${suggestion#$BUFFER}"
		_zsh_autosuggest_highlight_reset
		_zsh_autosuggest_highlight_apply
	else
		 zle expand-or-complete-prefix;
	fi
}
zle -N _hbt_search
bindkey '^I' _hbt_search

function _hbt_clear() {
	if [[ -z ${LBUFFER// } ]]; then
		unset POSTDISPLAY
	else
		zle backward-delete-char
	fi
}
zle -N _hbt_clear
bindkey '^?' _hbt_clear

function _hbt_delsuggestion () {
	if [[ ! -z ${POSTDISPLAY} ]]; then
		$(echo -n "del\n$$\n$(pwd)\n$1" | nc localhost $HBT_PORT)
		unset POSTDISPLAY
	fi
}
zle -N _hbt_delsuggestion
bindkey '^[[3~' _hbt_delsuggestion

