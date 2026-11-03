# desk shell integration for zsh. Load with: eval "$(desk init zsh)"

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

# Sets REPLY to "workspace/relative/path" when inside a workspace.
_desk_where() {
  local root="${DESK_HOME:-$HOME/.desk}/workspaces/"
  REPLY=""
  [[ $PWD == "$root"?* ]] || return 1
  REPLY="${PWD#"$root"}"
}

# For themes: PROMPT='$(desk_prompt_info) ...' with prompt_subst.
desk_prompt_info() {
  _desk_where && print -r -- "${REPLY//\%/%%}"
}

# Prefix the prompt with (workspace/path) while inside a workspace.
# Set DESK_PROMPT_DISABLE=1 to manage the prompt yourself.
_desk_precmd() {
  if [[ -n $_desk_prefix && $PROMPT == "$_desk_prefix"* ]]; then
    PROMPT="${PROMPT#"$_desk_prefix"}"
  fi
  _desk_prefix=""
  [[ -z $DESK_PROMPT_DISABLE ]] && _desk_where || return 0
  _desk_info="${REPLY//\%/%%}"
  if [[ -o prompt_subst ]]; then
    # Never substitute a directory name into the prompt as code.
    _desk_prefix='%F{cyan}(${_desk_info})%f '
  else
    _desk_prefix="%F{cyan}(${_desk_info})%f "
  fi
  PROMPT="$_desk_prefix$PROMPT"
}

autoload -Uz add-zsh-hook
add-zsh-hook precmd _desk_precmd

_desk() {
  if (( CURRENT == 2 )); then
    compadd new cd path list rm init version help
  elif (( CURRENT == 3 )); then
    case "$words[2]" in
      cd|path|rm) compadd -- "${DESK_HOME:-$HOME/.desk}"/workspaces/*(N/:t) ;;
      init) compadd zsh bash fish ;;
    esac
  fi
}
(( $+functions[compdef] )) && compdef _desk desk
