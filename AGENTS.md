# Overview

This is a Go CLI client for using the Wise API: https://docs.wise.com/api-reference

## Scope

The goal is to prepare payments with a personal API key.

## Structure

Instructions for fetching data go into queries/

Instructions for making changes go into commands/

## Naming

The binary is called `wise` (not `wise-cli`). The cache directory remains `~/.cache/wise-cli/` for backwards compatibility with existing installations.
