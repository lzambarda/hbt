# hbt

A heuristic command suggestion system for zsh.

## Installation

```
function hbt_start() {
	pid=$(pgrep hbtsrv)
	if [ -z $pid ]; then
			echo "starting hbt, you can stop it with: hbt_stop"
			nohup ~/Repositories/hbt/bin/hbtsrv 1234 ~/dotfiles/hbt >/dev/null 2>&1 &
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
	#hbt_start
fi

#autoload -Uz add-zsh-hook

function _hbt_track () {
	echo "you run $1 at $(pwd)"
}

add-zsh-hook preexec _hbt_track

# list dir with TAB, when there are only spaces/no text before cursor,
# or complete words, that are before cursor only (like in tcsh)
function _hbt_search () {
	if [[ -z ${LBUFFER// } ]]; then
		suggestion=testone
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
# bindkey ^h _hbt_search # ctrl+h
```
