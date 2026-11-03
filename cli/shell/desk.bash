# desk shell integration for bash. Load with: eval "$(desk init bash)"

desk() {
  local dir
  case "$1" in
    new)
      dir="$(command desk "$@")" || return
      builtin cd -- "$dir"
      ;;
    cd)
      shift
      dir="$(command desk path "$@")" || return
      builtin cd -- "$dir"
      ;;
    rm)
      command desk "$@" || return
      [[ -d $PWD ]] || builtin cd -- ~
      ;;
    *)
      command desk "$@"
      ;;
  esac
}

# Sets _desk_info to "workspace/relative/path" when inside a workspace.
_desk_where() {
  local root="${DESK_HOME:-$HOME/.desk}/workspaces/"
  _desk_info=""
  [[ $PWD == "$root"?* ]] || return 1
  _desk_info="${PWD#"$root"}"
}

# For custom prompts: PS1='$(desk_prompt_info) ...'
desk_prompt_info() {
  _desk_where && printf '%s\n' "$_desk_info"
}

# Prefix PS1 with (workspace/path) while inside a workspace.
# Set DESK_PROMPT_DISABLE=1 to manage the prompt yourself.
_desk_prompt() {
  local status=$?
  if [[ -n $_desk_prefix && $PS1 == "$_desk_prefix"* ]]; then
    PS1="${PS1#"$_desk_prefix"}"
  fi
  _desk_prefix=""
  if [[ -z $DESK_PROMPT_DISABLE ]] && _desk_where; then
    if shopt -q promptvars; then
      # Expanded at display time, so a directory name is never run as code.
      _desk_prefix='\[\e[36m\](${_desk_info})\[\e[0m\] '
    else
      _desk_prefix="\[\e[36m\](${_desk_info//\\/\\\\})\[\e[0m\] "
    fi
    PS1="$_desk_prefix$PS1"
  fi
  return $status
}

if [[ $(declare -p PROMPT_COMMAND 2>/dev/null) == "declare -a"* ]]; then
  [[ " ${PROMPT_COMMAND[*]} " == *" _desk_prompt "* ]] || PROMPT_COMMAND+=(_desk_prompt)
elif [[ ";${PROMPT_COMMAND};" != *";_desk_prompt;"* ]]; then
  PROMPT_COMMAND="${PROMPT_COMMAND:+${PROMPT_COMMAND%;};}_desk_prompt"
fi

_desk_complete() {
  local cur="${COMP_WORDS[COMP_CWORD]}" d
  COMPREPLY=()
  if (( COMP_CWORD == 1 )); then
    COMPREPLY=($(compgen -W "new cd path list rm init version help" -- "$cur"))
  elif (( COMP_CWORD == 2 )); then
    case "${COMP_WORDS[1]}" in
      cd|path|rm)
        for d in "${DESK_HOME:-$HOME/.desk}/workspaces/$cur"*/; do
          [[ -d $d ]] && d="${d%/}" && COMPREPLY+=("${d##*/}")
        done
        ;;
      init) COMPREPLY=($(compgen -W "zsh bash fish" -- "$cur")) ;;
    esac
  fi
}
complete -F _desk_complete desk
