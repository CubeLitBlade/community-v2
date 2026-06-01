#!/usr/bin/env bash

declare -A MODULES_PATH
declare -A MODULES_CMD

# Services
MODULES_PATH[account]="./services/account"
MODULES_CMD[account]="./cmd/account"

MODULES_PATH[post]="./services/post"
MODULES_CMD[post]="./cmd/post"

# Packages
MODULES_PATH[common]="./pkg/common"
MODULES_PATH[platform]="./pkg/platform"

# Aliases
MODULES_ALIASES=("${!MODULES_PATH[@]}")