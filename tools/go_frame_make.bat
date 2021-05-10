set frame_in_path=./frameproto
set out_path="../frame/framepb"

rd /s /Q %out_path%
md %out_path%

rem go has import protobuf error must manual write
protoc.exe --plugin=protoc-gen-go=.\protoc-gen-go.exe --proto_path %frame_in_path%  --go_out  %out_path%  frame.proto

pause