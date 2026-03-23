#!/usr/bin/env bash
# PostToolUse hook for Claude Code - runs prek on edited files
# Receives JSON on stdin with tool_input.file_path

file_path=$(jq -r '.tool_input.file_path // .tool_response.filePath // empty')
[[ -n $file_path ]] || exit 0

prek run --files "$file_path" || true
