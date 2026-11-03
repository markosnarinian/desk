# desk shell integration for fish. Load with: desk init fish --print | source

function desk --description 'workspaces for temporary local work'
    switch "$argv[1]"
        case new amp
            set -l dir (command desk $argv); or return
            builtin cd -- "$dir"
        case switch
            set -l dir (command desk info --path $argv[2..-1]); or return
            builtin cd -- "$dir"; or return
            command desk _visit
        case rename
            set -l old (command desk info --path 2>/dev/null)
            set -l relative (string sub -s (math (string length -- "$old") + 1) -- "$PWD")
            set -l dir (command desk $argv); or return
            if not test -d "$PWD"
                builtin cd -- "$dir$relative"
            end
        case rm cleanup
            command desk $argv; or return
            test -d $PWD; or builtin cd ~
        case '*'
            command desk $argv
    end
end

function _desk_root
    if set -q DESK_HOME
        echo $DESK_HOME/workspaces/
    else
        echo $HOME/.desk/workspaces/
    end
end

# Prints "workspace/relative/path" when inside a workspace.
function desk_prompt_info
    set -l root (_desk_root)
    set -l n (string length -- $root)
    test (string length -- $PWD) -gt $n; or return 1
    test (string sub -l $n -- $PWD) = $root; or return 1
    string sub -s (math $n + 1) -- $PWD
end

function _desk_status
    return $argv[1]
end

# Prefix the prompt with (workspace/path) while inside a workspace.
# Set DESK_PROMPT_DISABLE to manage the prompt yourself.
if functions -q fish_prompt; and not functions -q _desk_original_fish_prompt
    functions -c fish_prompt _desk_original_fish_prompt
    function fish_prompt
        set -l last $status
        if not set -q DESK_PROMPT_DISABLE; and set -l info (desk_prompt_info)
            printf '%s(%s)%s ' (set_color cyan) $info (set_color normal)
        end
        _desk_status $last
        _desk_original_fish_prompt
    end
end

complete -c desk -f
complete -c desk -n __fish_use_subcommand -a 'new amp switch info list rename finish cleanup rm init version help'
complete -c desk -n '__fish_seen_subcommand_from switch info rm rename finish' -a '(for d in (_desk_root)*/; basename $d; end)'
complete -c desk -n '__fish_seen_subcommand_from info' -l path -d 'Print only the path'
complete -c desk -n '__fish_seen_subcommand_from new' -l fresh -d 'Create an independent clone'
complete -c desk -n '__fish_seen_subcommand_from new' -l choose -d 'Show workspace chooser'
complete -c desk -n '__fish_seen_subcommand_from new' -l reuse -d 'Remember most-recent reuse'
complete -c desk -n '__fish_seen_subcommand_from cleanup' -l yes -d 'Skip confirmation, not safety checks'
complete -c desk -n '__fish_seen_subcommand_from amp' -l copy -d 'Take a one-time copy'
complete -c desk -n '__fish_seen_subcommand_from amp' -l follow -d 'Follow live orb changes'
complete -c desk -n '__fish_seen_subcommand_from amp' -l repository -r -d 'Repository to clone'
complete -c desk -n '__fish_seen_subcommand_from init' -a 'zsh bash fish oh-my-zsh'
complete -c desk -n '__fish_seen_subcommand_from init' -l print -d 'Print without installing'
