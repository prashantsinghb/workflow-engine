package service

//go:generate protoc -I . -I ../ -I ../../third_party -I /opt/homebrew/Cellar/protobuf/33.2/include --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative --grpc-gateway_opt=Mservice.proto=github.com/prashantsinghb/workflow-engine/pkg/gateway --openapiv2_out=../../api/openapi --openapiv2_opt=logtostderr=true,allow_merge=true service.proto
