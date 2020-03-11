bind C-F command-prompt -p "send favor:" -I "tmux.auto" 'run-shell -b "curl $ctx_dev/code/tmux/favor?cmds=%%"'
