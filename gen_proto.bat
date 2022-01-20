@echo off
protoc --proto_path=. --go_out=. uno.proto
