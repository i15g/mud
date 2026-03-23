#!/usr/bin/env bash
# PostToolUse hook - run go tests when editing a .go source file

file_path=$(jq -r '.tool_input.file_path // .tool_response.filePath // empty')
[[ -n $file_path ]] || exit 0
[[ $file_path == *.go ]] || exit 0
[[ $file_path == *_test.go ]] && exit 0

test_file="${file_path%.go}_test.go"
[[ -f $test_file ]] || exit 0

if ! output=$(cd "$(dirname "$file_path")" && go test . 2>&1); then
    jq -n --arg msg "go test failed after editing $file_path:
$output" \
        '{"hookSpecificOutput": {"hookEventName": "PostToolUse", "additionalContext": $msg}}'
fi
