# desk shell integration for fish. Load with: desk init fish | source

function desk --description 'Workspaces for the work in between'
    switch "$argv[1]"
        case new
            set -l dir (command desk $argv); or return
            builtin cd -- $dir
        case cd
            set -l dir (command desk path $argv[2..-1]); or return
            builtin cd -- $dir
        case rm
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
complete -c desk -n __fish_use_subcommand -a 'new cd path list rm init version help'
complete -c desk -n '__fish_seen_subcommand_from cd path rm' -a '(for d in (_desk_root)*/; basename $d; end)'
complete -c desk -n '__fish_seen_subcommand_from init' -a 'zsh bash fish'
